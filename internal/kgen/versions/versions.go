// Package versions generates the per-service version-gate declarations that pin
// the Kion release range each resource / data source supports.
//
// For every resource and data source declared in codegen/generator_config.yaml
// (merged over by codegen/config_overrides.yaml), it picks the DEFINING
// operation (create if present, else read) and scans the SDK's per-version
// generated clients (generated/<v>/oas_client_gen.go) to find which versions
// contain that exact (METHOD, path). The contiguous range of matching versions
// becomes the [min, max] support range, and, for entries with a bounded range,
// it writes internal/service/<name>/<name>_version_gen.go declaring:
//
//	var (
//		minKionVersion = conns.MustParseKionVersion("3.16.0")
//		maxKionVersion = conns.KionVersion{} // unbounded
//	)
//
// so the runtime version gate (framework.RequireKionVersionInRange) survives
// schema regeneration. See the file header rules in computeRange for the exact
// min/max/open-ended semantics.
//
// The single external boundary, the filesystem (kfs.FS). Is injected so the
// whole pipeline (config + SDK client reads and _version_gen.go writes) is
// unit-testable against a mock instead of touching disk.
package versions

import (
	"bufio"
	"fmt"
	"go/format"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"terraform-provider-kion/internal/kgen/crud"
	"terraform-provider-kion/internal/kgen/kfs"

	"gopkg.in/yaml.v3"
)

// Options configures a version-gate generation run. Empty fields fall back to
// the defaults below.
type Options struct {
	SDKDir        string // path to the kion-sdk-go module (default ../kion-sdk-go)
	ServiceRoot   string // OUTPUT root for <name>/<name>_version_gen.go (default internal/service)
	ConfigPath    string // generator_config.yaml (default codegen/generator_config.yaml)
	Overrides     string // config_overrides.yaml merged over the config (default codegen/config_overrides.yaml)
	AttrOverrides string // attr_version_overrides.yaml merged over the derived windows
	Force         bool   // overwrite existing _version_gen.go files (reserved; writes always overwrite today)
}

const (
	defaultSDKDir        = "../kion-sdk-go"
	defaultServiceRoot   = "internal/service"
	defaultConfigPath    = "codegen/generator_config.yaml"
	defaultAttrOverrides = "codegen/attr_version_overrides.yaml"
	defaultOverrides     = "codegen/config_overrides.yaml"
)

// TrackedVersion is one Kion release the provider generates version gates for.
// Entries are ascending; master is excluded.
type TrackedVersion struct {
	Dir   string
	Minor int
}

// TrackedVersions is the release window: the current Kion release plus three
// back. kion-sdk-go ships exactly this set.
//
// A version listed here but absent from the SDK stops generation outright
// (loadVersionOps errors on the missing client). One the SDK ships but this
// omits is silently ignored, leaving resources gated against a window that
// excludes the current release -- the quieter and more dangerous direction.
//
// Declared once and consumed by cmd/kversions rather than duplicated. The SDK
// had this same set spread across nine places; a missed one caused a two-week
// pipeline outage and a release that shipped without the new version.
var TrackedVersions = []TrackedVersion{
	{Dir: "v3_14", Minor: 14},
	{Dir: "v3_15", Minor: 15},
	{Dir: "v3_16", Minor: 16},
	{Dir: "v3_17", Minor: 17},
}

var trackedVersions = func() []version {
	out := make([]version, 0, len(TrackedVersions))
	for _, tv := range TrackedVersions {
		out = append(out, version{dir: tv.Dir, minor: tv.Minor})
	}
	return out
}()

// CurrentSDK is the sub-package the provider compiles against: the newest
// tracked version.
func CurrentSDK() string { return TrackedVersions[len(TrackedVersions)-1].Dir }

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

// mergeOverrides merges override entries OVER base. For each resource /
// data-source present in override, the individual operation sub-keys that are
// set in the override replace the corresponding base operation; base operations
// not mentioned in the override are preserved.
func mergeOverrides(base, override *config) *config {
	return &config{
		Resources:   mergeSection(base.Resources, override.Resources),
		DataSources: mergeSection(base.DataSources, override.DataSources),
	}
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

// support is the derived version-support range for one entry. Each field is
// either a "3.NN.0" string or empty, where empty means "unbounded" on that side.
type support struct {
	Min string
	Max string
}

// rangeResult captures the outcome of deriving a range for one entry.
type rangeResult struct {
	// minIdx/maxIdx index into trackedVersions.
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
//     (no gate needed) so no file is written.
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

// generator holds the injected filesystem boundary. Use Generate for the real
// thing.
type generator struct {
	fs kfs.FS
	// src reads the SDK's generated Go for attribute derivation. Injected
	// alongside fs so a test can drive the attribute path with a stub instead
	// of needing real SDK files on disk.
	src crud.Source
}

// Generate runs the full pipeline with the real filesystem, returning the number
// of _version_gen.go files written.
func Generate(opts Options) (int, error) {
	return (&generator{fs: kfs.OS{}, src: crud.NewFileSource()}).generate(opts)
}

func (g *generator) generate(opts Options) (int, error) {
	sdkDir := orDefault(opts.SDKDir, defaultSDKDir)
	serviceRoot := orDefault(opts.ServiceRoot, defaultServiceRoot)
	configPath := orDefault(opts.ConfigPath, defaultConfigPath)
	overridesPath := orDefault(opts.Overrides, defaultOverrides)

	logw := os.Stderr

	base, err := g.loadConfig(configPath)
	if err != nil {
		return 0, fmt.Errorf("loading generator config: %w", err)
	}
	merged := base
	if ov, err := g.loadConfig(overridesPath); err != nil {
		if !os.IsNotExist(err) {
			return 0, fmt.Errorf("loading overrides: %w", err)
		}
		fmt.Fprintf(logw, "note: no overrides file at %s; using generator config as-is\n", overridesPath)
	} else {
		merged = mergeOverrides(base, ov)
	}

	versionOps := make([]map[op]struct{}, len(trackedVersions))
	for i, v := range trackedVersions {
		ops, err := g.loadVersionOps(sdkDir, v)
		if err != nil {
			return 0, err
		}
		versionOps[i] = ops
	}

	floors, err := LoadFloors(g.fs.ReadFile, orDefault(opts.AttrOverrides, defaultAttrOverrides),
		SortedKeys(merged.Resources), SortedKeys(merged.DataSources))
	if err != nil {
		return 0, err
	}

	resources, err := deriveSection("resources", merged.Resources, versionOps, floors, logw)
	if err != nil {
		return 0, err
	}
	dataSources, err := deriveSection("data_sources", merged.DataSources, versionOps, floors, logw)
	if err != nil {
		return 0, err
	}

	// Attribute gating is independent of resource gating: a resource present in
	// every tracked version can still carry a field only newer Kions accept.
	src := g.src
	if src == nil {
		src = crud.NewFileSource()
	}
	attrMins := deriveAttrWindows(src, sdkDir, serviceRoot, merged.Resources, logw)
	applyAttrOverrides(attrMins, orDefault(opts.AttrOverrides, defaultAttrOverrides), g.fs, logw)

	// So a package needing only attribute gates still gets a file. Its support
	// range is unbounded on both sides, making the resource-level check a no-op.
	gated := make(map[string]bool, len(resources))
	for _, e := range resources {
		gated[e.name] = true
	}
	for name := range attrMins {
		if !gated[name] {
			resources = append(resources, resolvedEntry{name: name})
		}
	}

	// A resource and its data source share a service package; both derive from
	// the same defining op family, so gate on the union keyed by service name.
	// Later (data_sources) entries do not clobber earlier ones with the same
	// name because the range is derived per defining op, but we dedupe by name
	// to avoid writing the same file twice.
	// Only resources get a ModifyPlan: data sources have no plan to modify, and
	// their struct is named differently (gcpRegionsDataSource, not
	// gcp_regionsDataSource), so the <name>Resource convention does not apply.
	isResource := make(map[string]bool, len(resources))
	for _, e := range resources {
		isResource[e.name] = true
	}

	seen := make(map[string]bool)
	var written int
	for _, e := range append(resources, dataSources...) {
		if seen[e.name] {
			continue
		}
		seen[e.name] = true
		dir := filepath.Join(serviceRoot, e.name)
		path := filepath.Join(dir, e.name+"_version_gen.go")

		// Build safety: some service packages still declare an inline
		// var minKionVersion / maxKionVersion in a hand-written source file.
		// Writing our generated file into such a package would be a duplicate
		// declaration => compile error. Skip (and log) those; migrating them to
		// the generated file is a separate step. Our own _version_gen.go is
		// excluded from the scan so re-runs are idempotent.
		if !opts.Force {
			if who, err := g.inlineVersionVarFile(dir, e.name); err != nil {
				return written, err
			} else if who != "" {
				fmt.Fprintf(logw, "skip: %s already declares minKionVersion/maxKionVersion in %s; "+
					"remove the inline var to let `kgen versions` own it\n", e.name, who)
				continue
			}
		}

		content, err := renderVersionGen(e.name, e.support, isResource[e.name], pruneRedundant(attrMins[e.name], e.support.Min))
		if err != nil {
			return written, fmt.Errorf("rendering %s: %w", e.name, err)
		}
		if err := g.fs.MkdirAll(dir, 0o750); err != nil {
			return written, fmt.Errorf("creating %s: %w", dir, err)
		}
		if err := g.fs.WriteFile(path, content, 0o600); err != nil {
			return written, fmt.Errorf("writing %s: %w", path, err)
		}
		fmt.Fprintf(logw, "wrote %s (min=%s max=%s)\n", path, orDash(e.support.Min), orDash(e.support.Max))
		written++
	}
	return written, nil
}

// inlineVarRE matches a hand-written top-level declaration of minKionVersion or
// maxKionVersion (with or without a var group), e.g.
//
//	var minKionVersion = conns.MustParseKionVersion("3.16.0")
//	    maxKionVersion = conns.KionVersion{}
var inlineVarRE = regexp.MustCompile(`(?m)^\s*(?:var\s+)?(?:min|max)KionVersion\s*=`)

// inlineVersionVarFile reports the base name of a NON-generated .go file in dir
// that already declares minKionVersion / maxKionVersion, or "" if none does.
// The generated <name>_version_gen.go is ignored so re-runs are idempotent. A
// missing directory is treated as "no collision".
func (g *generator) inlineVersionVarFile(dir, name string) (string, error) {
	entries, err := g.fs.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("scanning %s: %w", dir, err)
	}
	generated := name + "_version_gen.go"
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		fn := e.Name()
		if !strings.HasSuffix(fn, ".go") || fn == generated || strings.HasSuffix(fn, "_test.go") {
			continue
		}
		b, err := g.fs.ReadFile(filepath.Join(dir, fn))
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", filepath.Join(dir, fn), err)
		}
		if inlineVarRE.Match(b) {
			return fn, nil
		}
	}
	return "", nil
}

// loadConfig reads and parses a generator_config.yaml / config_overrides.yaml via
// the injected filesystem.
func (g *generator) loadConfig(path string) (*config, error) {
	b, err := g.fs.ReadFile(path)
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

// loadVersionOps reads and parses the client for one tracked version via the
// injected filesystem.
func (g *generator) loadVersionOps(sdkDir string, v version) (map[op]struct{}, error) {
	p := filepath.Join(sdkDir, "generated", v.dir, "oas_client_gen.go")
	b, err := g.fs.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading client for %s: %w", v.dir, err)
	}
	return parseClientOps(string(b)), nil
}

// deriveSection derives support ranges for all entries in one config section
// (resources or data_sources). It returns the entries that should be emitted
// (sorted by name) and logs unresolved / non-contiguous / fully-supported
// entries to the writer.
//
// A floor raises the derived minimum, and forces a gate for an entry that would
// otherwise need none. An entry a floor names that derives nothing is an error.
func deriveSection(section string, entries map[string]entry, versionOps []map[op]struct{}, floors map[string]string, logw io.Writer) ([]resolvedEntry, error) {
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
			if _, _, err := ApplyFloor(name, section, floors[name], FloorNoOp, "", "", ""); err != nil {
				return nil, err
			}
			fmt.Fprintf(logw, "skip: %s/%s has no create or read operation\n", section, name)
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
			if _, _, err := ApplyFloor(name, section, floors[name], FloorUnresolved, defOp.String(), "", ""); err != nil {
				return nil, err
			}
			fmt.Fprintf(logw, "skip: %s/%s defining op %q not found in any tracked SDK version\n", section, name, defOp)
			continue
		}
		raised, forced, err := ApplyFloor(name, section, floors[name], FloorResolved, "", r.support.Min, r.support.Max)
		if err != nil {
			return nil, err
		}
		r.support.Min = raised
		r.emit = r.emit || forced
		if !r.contiguous {
			fmt.Fprintf(logw, "warning: %s/%s defining op %q present in a non-contiguous set of versions; using overall min/max (%s..%s)\n",
				section, name, defOp, versionString(trackedVersions[r.minIdx]), versionString(trackedVersions[r.maxIdx]))
		}
		if !r.emit {
			// Fully supported (oldest..open): no gate needed, no file written.
			continue
		}
		out = append(out, resolvedEntry{name: name, support: r.support})
	}
	return out, nil
}

// renderVersionGen renders (and gofmts) the <name>_version_gen.go source for a
// service package with the given derived support range. A bounded min emits
// MustParseKionVersion; an unbounded side emits conns.KionVersion{}.
func renderVersionGen(name string, s support, isResource bool, attrMins map[string][]attrWindow) ([]byte, error) {
	var minExpr, maxComment string
	if s.Min != "" {
		minExpr = fmt.Sprintf("conns.MustParseKionVersion(%q)", s.Min)
	} else {
		minExpr = "conns.KionVersion{}"
	}
	var maxExpr string
	if s.Max != "" {
		maxExpr = fmt.Sprintf("conns.MustParseKionVersion(%q)", s.Max)
	} else {
		maxExpr = "conns.KionVersion{}"
		maxComment = " // unbounded"
	}

	src := fmt.Sprintf(`// Code generated by `+"`kgen versions`"+`; DO NOT EDIT.
package %s

import "terraform-provider-kion/internal/conns"

var (
	minKionVersion = %s
	maxKionVersion = %s%s
)
`, name, minExpr, maxExpr, maxComment)

	if isResource {
		// Always declared, empty or not: ModifyPlan references it unconditionally.
		attrBlock := "\n// attrKionVersions is the Kion version window accepting each attribute. An\n" +
			"// attribute absent here is accepted for as long as the operation exists.\n" +
			"var attrKionVersions = map[string]framework.AttrVersions{\n" +
			renderAttrMins(attrMins) + "}\n"

		src = fmt.Sprintf(`// Code generated by `+"`kgen versions`"+`; DO NOT EDIT.
package %s

import (
	"context"

	"terraform-provider-kion/internal/conns"
	"terraform-provider-kion/internal/framework"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	minKionVersion = %s
	maxKionVersion = %s%s
)

%s
var _ resource.ResourceWithModifyPlan = &%sResource{}

// ModifyPlan reports an unsupported Kion version during plan. The CRUD methods
// gate too, but they run at apply, by which point the practitioner has already
// been told the change is going ahead.
//
// A destroy plan is exempt: refusing it would strand the resource in state with
// no way out but terraform state rm.
func (r *%sResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(framework.RequireKionVersionInRange(r.Meta(), minKionVersion, maxKionVersion, "kion_%s")...)
	resp.Diagnostics.Append(framework.RequireAttrKionVersions(r.Meta(), req.Plan, attrKionVersions, "kion_%s")...)
}
`, name, minExpr, maxExpr, maxComment, attrBlock, name, name, name, name)
	}

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return nil, fmt.Errorf("gofmt %s_version_gen.go: %w", name, err)
	}
	return formatted, nil
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func orDash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}
