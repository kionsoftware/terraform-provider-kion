// Command kversions derives the range of Kion versions that support each
// resource / data source and writes it to codegen/version_support.yaml.
//
// For every resource and data source declared in codegen/generator_config.yaml
// (merged over by codegen/config_overrides.yaml), it picks the DEFINING
// operation (create if present, else read) and scans the SDK's per-version
// generated clients (generated/<v>/oas_client_gen.go) to find which versions
// contain that exact (METHOD, path). The contiguous range of matching versions
// becomes the [min, max] support range. See the generated file's header and the
// task spec for the exact min/max/open-ended rules.
//
// This tool is read-only against its inputs; it only writes version_support.yaml.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"terraform-provider-kion/internal/kgen/versions"

	"gopkg.in/yaml.v3"
)

// trackedVersions are the SDK support versions, ascending. master is excluded.
// Each entry maps a version directory name to its minor number.
// trackedVersions is the canonical set, defined once in internal/kgen/versions.
// Keeping a second copy here is what the SDK's own nine-copy version set is a
// cautionary tale about.
var trackedVersions = func() []version {
	out := make([]version, 0, len(versions.TrackedVersions))
	for _, v := range versions.TrackedVersions {
		out = append(out, version{dir: v.Dir, minor: v.Minor})
	}
	return out
}()

type version struct {
	dir   string
	minor int
}

// versionString maps a tracked version to its "3.NN.0" release string.
func versionString(v version) string {
	return fmt.Sprintf("3.%d.0", v.minor)
}

// op is a single (METHOD, path) API operation. METHOD is upper-cased; path
// retains its {id} placeholders and any synthetic __qs__ segments verbatim.
type op struct {
	method string
	path   string
}

func (o op) String() string { return o.method + " " + o.path }

// opLineRE matches a client doc line like "// POST /v3/ou-note".
var opLineRE = regexp.MustCompile(`^//\s+(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+(\S+)\s*$`)

// clientFuncRE matches a top-level client method: "func (c *Client) Name(".
var clientFuncRE = regexp.MustCompile(`^func \(c \*Client\) [A-Z]`)

// parseClientOps parses the contents of an oas_client_gen.go file into the set
// of (METHOD, path) operations it exposes. An operation is recognized as a
// "// METHOD /path" doc line immediately followed by a top-level
// "func (c *Client) <OpName>(" declaration.
func parseClientOps(src string) map[op]struct{} {
	ops := make(map[op]struct{})
	sc := bufio.NewScanner(strings.NewReader(src))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var prevOp *op
	var prevMatched bool
	for sc.Scan() {
		line := sc.Text()
		if clientFuncRE.MatchString(line) {
			if prevMatched && prevOp != nil {
				ops[*prevOp] = struct{}{}
			}
			prevMatched = false
			prevOp = nil
			continue
		}
		if m := opLineRE.FindStringSubmatch(line); m != nil {
			o := op{method: strings.ToUpper(m[1]), path: m[2]}
			prevOp = &o
			prevMatched = true
			continue
		}
		// Any other line breaks the "immediately preceding" adjacency.
		prevMatched = false
		prevOp = nil
	}
	return ops
}

// loadVersionOps reads and parses the client for one tracked version.
func loadVersionOps(sdkDir string, v version) (map[op]struct{}, error) {
	p := filepath.Join(sdkDir, "generated", v.dir, "oas_client_gen.go")
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading client for %s: %w", v.dir, err)
	}
	return parseClientOps(string(b)), nil
}

// opRef is a {path, method} pair from the generator config.
type opRef struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

// entry is a single resource / data-source config entry. Only the operation
// sub-keys are decoded; schema and other keys are ignored.
type entry struct {
	Create *opRef `yaml:"create"`
	Read   *opRef `yaml:"read"`
	Update *opRef `yaml:"update"`
	Delete *opRef `yaml:"delete"`
}

// config is the decoded shape of generator_config.yaml / config_overrides.yaml.
type config struct {
	Resources   map[string]entry `yaml:"resources"`
	DataSources map[string]entry `yaml:"data_sources"`
}

func loadConfig(path string) (*config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if c.Resources == nil {
		c.Resources = map[string]entry{}
	}
	if c.DataSources == nil {
		c.DataSources = map[string]entry{}
	}
	return &c, nil
}

// mergeOverrides merges override entries OVER base. For each resource /
// data-source present in override, the individual operation sub-keys that are
// set in the override replace the corresponding base operation; base operations
// not mentioned in the override are preserved.
func mergeOverrides(base, override *config) *config {
	merged := &config{
		Resources:   mergeSection(base.Resources, override.Resources),
		DataSources: mergeSection(base.DataSources, override.DataSources),
	}
	return merged
}

func mergeSection(base, override map[string]entry) map[string]entry {
	out := make(map[string]entry, len(base))
	maps.Copy(out, base)
	for k, ov := range override {
		be := out[k] // zero entry if not present
		if ov.Create != nil {
			be.Create = ov.Create
		}
		if ov.Read != nil {
			be.Read = ov.Read
		}
		if ov.Update != nil {
			be.Update = ov.Update
		}
		if ov.Delete != nil {
			be.Delete = ov.Delete
		}
		out[k] = be
	}
	return out
}

// definingOp picks the operation that defines a resource's existence: create if
// present, otherwise read. Returns ok=false if neither is set.
func definingOp(e entry) (op, bool) {
	var ref *opRef
	switch {
	case e.Create != nil:
		ref = e.Create
	case e.Read != nil:
		ref = e.Read
	default:
		return op{}, false
	}
	if ref.Path == "" || ref.Method == "" {
		return op{}, false
	}
	return op{method: strings.ToUpper(ref.Method), path: ref.Path}, true
}

// support is the derived version-support range for one entry.
type support struct {
	Min string `yaml:"min,omitempty"`
	Max string `yaml:"max,omitempty"`
}

// rangeResult captures the outcome of deriving a range for one entry.
type rangeResult struct {
	// present indexes into trackedVersions: present[i] is true when the
	// defining op exists in trackedVersions[i].
	minIdx, maxIdx int
	resolved       bool // false when the op is in no tracked version
	contiguous     bool // whether the matching versions form a contiguous run
	// emit is false when the range is "fully supported" (min == oldest AND
	// max open) and therefore needs no gate.
	emit    bool
	support support
}

// computeRange derives the [min, max] support range from a per-version presence
// slice aligned with trackedVersions (ascending). See the file header rules:
//   - empty => unresolved (skip).
//   - min = "3.<minMinor>.0".
//   - max omitted when maxIdx is the newest tracked version (still current).
//   - otherwise max = "3.<maxMinor>.0".
//   - an entry that is min==oldest AND max==open is fully supported: emit=false
//     (no gate needed) so the output file stays small.
func computeRange(present []bool) rangeResult {
	minIdx, maxIdx := -1, -1
	for i, p := range present {
		if p {
			if minIdx == -1 {
				minIdx = i
			}
			maxIdx = i
		}
	}
	if minIdx == -1 {
		return rangeResult{resolved: false}
	}

	// Contiguity: every version between min and max inclusive is present.
	contiguous := true
	for i := minIdx; i <= maxIdx; i++ {
		if !present[i] {
			contiguous = false
			break
		}
	}

	res := rangeResult{
		minIdx:     minIdx,
		maxIdx:     maxIdx,
		resolved:   true,
		contiguous: contiguous,
	}

	newest := len(trackedVersions) - 1
	oldest := 0

	var s support
	s.Min = versionString(trackedVersions[minIdx])
	openMax := maxIdx == newest
	if !openMax {
		s.Max = versionString(trackedVersions[maxIdx])
	}
	res.support = s

	// Fully supported (min oldest AND max open) => no gate needed.
	res.emit = minIdx != oldest || !openMax
	return res
}

// resolvedEntry pairs a config entry name with its derived support range.
type resolvedEntry struct {
	name    string
	support support
}

// deriveSection derives support ranges for all entries in one config section
// (resources or data_sources). It returns the entries that should be emitted
// (sorted by name) and logs unresolved / non-contiguous entries to stderr.
// Floors come from versions.ParseResourceFloors and raise the derived minimum;
// an entry with no bounded range but a floor still gets a gate.
func deriveSection(section string, entries map[string]entry, versionOps []map[op]struct{}, floors map[string]string, logw *diag) ([]resolvedEntry, error) {
	names := make([]string, 0, len(entries))
	for k := range entries {
		names = append(names, k)
	}
	sort.Strings(names)

	var out []resolvedEntry
	for _, name := range names {
		e := entries[name]
		defOp, ok := definingOp(e)
		if !ok {
			if _, _, err := versions.ApplyFloor(name, section, floors[name], versions.FloorNoOp, "", "", ""); err != nil {
				return nil, err
			}
			logw.printf("unresolved: %s/%s has no create or read operation\n", section, name)
			continue
		}
		present := make([]bool, len(trackedVersions))
		for i, ops := range versionOps {
			if _, found := ops[defOp]; found {
				present[i] = true
			}
		}
		r := computeRange(present)
		if !r.resolved {
			if _, _, err := versions.ApplyFloor(name, section, floors[name], versions.FloorUnresolved, defOp.String(), "", ""); err != nil {
				return nil, err
			}
			logw.printf("unresolved: %s/%s defining op %q not found in any tracked SDK version\n", section, name, defOp)
			continue
		}
		raised, forced, err := versions.ApplyFloor(name, section, floors[name], versions.FloorResolved, "", r.support.Min, r.support.Max)
		if err != nil {
			return nil, err
		}
		r.support.Min = raised
		r.emit = r.emit || forced
		if !r.contiguous {
			logw.printf("warning: %s/%s defining op %q present in a non-contiguous set of versions; using overall min/max (%s..%s)\n",
				section, name, defOp, versionString(trackedVersions[r.minIdx]), versionString(trackedVersions[r.maxIdx]))
		}
		if !r.emit {
			// Fully supported (oldest..open): no gate needed, keep file small.
			continue
		}
		out = append(out, resolvedEntry{name: name, support: r.support})
	}
	return out, nil
}

// versionSupportHeader is the header comment for the generated file.
const versionSupportHeader = `# Generated by ` + "`make version-support`" + ` (cmd/kversions) from the SDK's per-version
# generated clients. Do not edit by hand, re-run after refreshing the SDK.
#
# Each entry gives the Kion version range in which a resource / data source's
# defining API operation (create if present, else read) exists:
#   min: oldest tracked version that has the op ("3.NN.0").
#   max: newest version that has the op; OMITTED when the op is still present in
#        the newest tracked version (unbounded above / still current).
# Resources whose op is present in every tracked version (min oldest, max open)
# need no gate and are omitted entirely to keep this file small.
`

// marshalOutput renders the resolved entries as YAML with the file header,
// deterministic key ordering, and quoted version strings.
func marshalOutput(resources, dataSources []resolvedEntry) ([]byte, error) {
	root := &yaml.Node{Kind: yaml.MappingNode}

	if len(resources) > 0 {
		root.Content = append(root.Content,
			scalarNode("resources"), sectionNode(resources))
	}
	if len(dataSources) > 0 {
		root.Content = append(root.Content,
			scalarNode("data_sources"), sectionNode(dataSources))
	}

	body, err := yaml.Marshal(root)
	if err != nil {
		return nil, err
	}
	out := versionSupportHeader + string(body)
	return []byte(out), nil
}

func sectionNode(entries []resolvedEntry) *yaml.Node {
	m := &yaml.Node{Kind: yaml.MappingNode}
	for _, e := range entries {
		sup := &yaml.Node{Kind: yaml.MappingNode}
		sup.Content = append(sup.Content, scalarNode("min"), quotedScalar(e.support.Min))
		if e.support.Max != "" {
			sup.Content = append(sup.Content, scalarNode("max"), quotedScalar(e.support.Max))
		}
		m.Content = append(m.Content, scalarNode(e.name), sup)
	}
	return m
}

func scalarNode(v string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: v}
}

func quotedScalar(v string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Value: v}
}

// diag is a best-effort diagnostic writer: it remembers the first write error
// so individual call sites don't each have to check it (mirrors
// internal/kalign's errWriter). Diagnostics are non-critical, so the remembered
// error is only used to short-circuit further writes, never surfaced.
type diag struct {
	w   io.Writer
	err error
}

func (d *diag) printf(format string, a ...any) {
	if d.err != nil {
		return
	}
	_, d.err = fmt.Fprintf(d.w, format, a...)
}

func (d *diag) println(a ...any) {
	if d.err != nil {
		return
	}
	_, d.err = fmt.Fprintln(d.w, a...)
}

// run is the testable entry point: it takes the CLI arguments (excluding the
// program name) and its output streams, and returns the process exit code
// instead of calling os.Exit directly. Progress and diagnostics go to stderr;
// the derived support map is written to the -out file. stdout is accepted for
// signature parity with the other CLI tools but is currently unused.
func run(args []string, _ io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("kversions", flag.ContinueOnError)
	fs.SetOutput(stderr)
	sdkDir := fs.String("sdk", "../kion-sdk-go", "path to the kion-sdk-go module (with generated/<v>/oas_client_gen.go)")
	genConfig := fs.String("config", "codegen/generator_config.yaml", "path to generator_config.yaml")
	overrides := fs.String("overrides", "codegen/config_overrides.yaml", "path to config_overrides.yaml (merged over the config)")
	attrOverrides := fs.String("attr-overrides", "codegen/attr_version_overrides.yaml", "path to attr_version_overrides.yaml (its resource_floors raise the derived minimums)")
	outPath := fs.String("out", "codegen/version_support.yaml", "output path for the derived version-support yaml")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	d := &diag{w: stderr}

	base, err := loadConfig(*genConfig)
	if err != nil {
		d.println("kversions:", fmt.Errorf("loading generator config: %w", err))
		return 1
	}
	// config_overrides.yaml is optional; a missing file means no overrides.
	merged := base
	if ov, err := loadConfig(*overrides); err != nil {
		if !os.IsNotExist(err) {
			d.println("kversions:", fmt.Errorf("loading overrides: %w", err))
			return 1
		}
		d.printf("note: no overrides file at %s; using generator config as-is\n", *overrides)
	} else {
		merged = mergeOverrides(base, ov)
	}

	versionOps := make([]map[op]struct{}, len(trackedVersions))
	for i, v := range trackedVersions {
		ops, err := loadVersionOps(*sdkDir, v)
		if err != nil {
			d.println("kversions:", err)
			return 1
		}
		versionOps[i] = ops
	}

	floors, err := versions.LoadFloors(os.ReadFile, *attrOverrides,
		versions.SortedKeys(merged.Resources), versions.SortedKeys(merged.DataSources))
	if err != nil {
		d.println("kversions:", err)
		return 1
	}

	resources, err := deriveSection("resources", merged.Resources, versionOps, floors, d)
	if err != nil {
		d.println("kversions:", err)
		return 1
	}
	dataSources, err := deriveSection("data_sources", merged.DataSources, versionOps, floors, d)
	if err != nil {
		d.println("kversions:", err)
		return 1
	}

	out, err := marshalOutput(resources, dataSources)
	if err != nil {
		d.println("kversions:", fmt.Errorf("marshaling output: %w", err))
		return 1
	}
	if err := os.WriteFile(*outPath, out, 0o644); err != nil {
		d.println("kversions:", fmt.Errorf("writing %s: %w", *outPath, err))
		return 1
	}
	d.printf("wrote %s: %d resources, %d data sources\n", *outPath, len(resources), len(dataSources))
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
