// Package config derives the tfplugingen generator config (codegen/generator_config.yaml)
// from what's already in the repo: each service package's resource/data-source
// code (the SDK ops its CRUD methods call) mapped to spec paths. Implemented
// resources are authoritative; stubs fall back to the spec naming convention and
// are flagged for review.
//
//	kgen config gen    # print the derived config (or --write it)
//	kgen config check  # report drift between the config and the code (CI gate)
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Options configures a run.
type Options struct {
	Spec        string // OpenAPI 3 spec (default spec/openapi3.json)
	ConfigPath  string // generator config to write/compare (default codegen/generator_config.yaml)
	ServiceRoot string // default internal/service
	Overrides   string // hand-tuned entries merged over the derivation (default codegen/config_overrides.yaml)
}

const defaultOverrides = "codegen/config_overrides.yaml"

// CRUD holds the four resolved operations (nil when absent/unresolved).
type CRUD struct {
	Create, Read, Update, Delete *Op
}

func (c *CRUD) empty() bool {
	return c == nil || (c.Create == nil && c.Read == nil && c.Update == nil && c.Delete == nil)
}

// ServiceConfig is the derived config for one service package.
type ServiceConfig struct {
	Name                   string
	Resource               *CRUD
	ResourceHeuristic      bool     // derived from spec naming, not code
	Overridden             bool     // paths supplied by codegen/config_overrides.yaml (config-verified, not code-implemented)
	ExtraIgnores           []string // extra OAS attributes to drop, from config_overrides (appended to baked-in ignores)
	DataSourceRead         *Op
	DataSourceExtraIgnores []string // as ExtraIgnores, for the data-source schema
	// DataSourceReadMatch filters which parents a parent-scoped read is made
	// under; the import manifest reads it from the rendered config.
	DataSourceReadMatch map[string]string
}

var acronyms = map[string]string{
	"ou": "OU", "gcp": "GCP", "aws": "AWS", "iam": "IAM", "saml": "SAML",
	"idms": "IDMS", "cft": "CFT", "ami": "AMI", "arm": "ARM", "oci": "OCI",
	"api": "API", "scp": "SCP", "id": "ID",
}

func toPascal(snake string) string {
	parts := strings.Split(snake, "_")
	for i, p := range parts {
		switch {
		case p == "":
		case acronyms[p] != "":
			parts[i] = acronyms[p]
		default:
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

// pickVerb returns the candidate op present in the index whose method is one of
// the wanted methods, preferring one that addresses the RECORD rather than a
// sub-resource hanging off it.
//
// The preference is not cosmetic. A service's update method may call more than
// one write op: kion_cloud_rule and kion_ou both PATCH the record AND PUT its
// labels, because label sync is part of Update. Taking the first match made
// `kconfig gen` rewrite their update to
//
//	path: /v3/cloud-rule/{cloud_rule_id}/labels
//	method: PUT
//
// which is a different endpoint doing a different thing -- so regenerating the
// config silently broke both resources, and `make codegen-check` could not pass
// on a clean tree.
//
// A record-addressing path ends at its id parameter; a sub-resource has static
// segments after it. That distinction holds for every op here and does not
// depend on knowing which sub-resources exist.
func pickVerb(candidates []string, ops map[string]Op, methods ...string) *Op {
	var fallback *Op
	for _, c := range candidates {
		op, ok := ops[c]
		if !ok || !slices.Contains(methods, op.Method) {
			continue
		}
		o := op
		if addressesRecord(o.Path) {
			return &o
		}
		if fallback == nil {
			fallback = &o
		}
	}
	// Every candidate is a sub-resource op. Returning one is still better than
	// dropping the verb, and a create legitimately has no id in its path.
	return fallback
}

// addressesRecord reports whether a path ends at a path parameter, i.e. names
// the record itself rather than something beneath it.
func addressesRecord(path string) bool {
	p := strings.TrimSuffix(path, "/")
	i := strings.LastIndexByte(p, '/')
	if i < 0 {
		return false
	}
	last := p[i+1:]
	return strings.HasPrefix(last, "{") && strings.HasSuffix(last, "}")
}

func heuristic(name string, ops map[string]Op) *CRUD {
	p := toPascal(name)
	c := &CRUD{}
	if op, ok := ops["Post"+p]; ok {
		c.Create = &op
	}
	if op, ok := ops["Get"+p]; ok {
		c.Read = &op
	}
	if op, ok := ops["Patch"+p]; ok {
		c.Update = &op
	}
	if op, ok := ops["Delete"+p]; ok {
		c.Delete = &op
	}
	return c
}

// Derive builds the config for every service package, preferring the ops the
// code actually calls and falling back to the spec naming convention for stubs.
func Derive(src Source, opts Options) ([]ServiceConfig, error) {
	ops, err := src.Operations(orDefault(opts.Spec, "spec/openapi3.json"))
	if err != nil {
		return nil, fmt.Errorf("reading spec operations: %w", err)
	}
	svcs, err := src.ServiceOps(orDefault(opts.ServiceRoot, "internal/service"))
	if err != nil {
		return nil, fmt.Errorf("reading service ops: %w", err)
	}

	var out []ServiceConfig
	for _, so := range svcs {
		sc := ServiceConfig{Name: so.Name}
		r := &CRUD{
			// A raw create has no operationId to look up; the parser
			// reconstructs it from the call itself.
			Create: orRaw(pickVerb(so.Create, ops, "POST"), so.RawCreate),
			Read:   orRaw(pickVerb(so.Read, ops, "GET"), so.RawRead),
			Update: orRaw(pickVerb(so.Update, ops, "PATCH", "PUT"), so.RawUpdate),
			Delete: orRaw(pickVerb(so.Delete, ops, "DELETE"), so.RawDelete),
		}
		if r.empty() {
			if h := heuristic(so.Name, ops); !h.empty() {
				r = h
				sc.ResourceHeuristic = true
			}
		}
		if !r.empty() {
			sc.Resource = r
		}
		sc.DataSourceRead = pickVerb(so.DataSourceRead, ops, "GET")
		if sc.Resource != nil || sc.DataSourceRead != nil {
			out = append(out, sc)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Gen renders the derived config (with overrides merged in) to w.
func Gen(src Source, opts Options, w io.Writer) error {
	configs, err := Derive(src, opts)
	if err != nil {
		return err
	}
	if configs, err = withOverrides(configs, opts); err != nil {
		return err
	}
	_, err = io.WriteString(w, render(configs))
	return err
}

func render(configs []ServiceConfig) string {
	var b strings.Builder
	b.WriteString("# Generated by `kgen config gen` from the service packages + spec/openapi3.json.\n")
	b.WriteString("# Entries flagged `# heuristic` were derived from the spec naming convention\n")
	b.WriteString("# (the resource is a stub with no SDK calls yet); verify their paths.\n")
	b.WriteString("# `# WARN: no read` means tfplugingen will reject the resource until a read is set.\n\n")
	b.WriteString("provider:\n  name: kion\n\n")

	b.WriteString("resources:\n")
	for _, c := range configs {
		if c.Resource == nil {
			continue
		}
		// tfplugingen requires both create and read. List incomplete resources
		// as a comment (kept inactive) so generation still works on the rest;
		// supply the missing op(s) via codegen/config_overrides.yaml to enable.
		var missing []string
		if c.Resource.Create == nil {
			missing = append(missing, "create")
		}
		if c.Resource.Read == nil {
			missing = append(missing, "read")
		}
		if len(missing) > 0 {
			fmt.Fprintf(&b, "  # %s: INCOMPLETE, missing %s (supply via codegen/config_overrides.yaml)\n",
				c.Name, strings.Join(missing, "+"))
			continue
		}
		flag := ""
		if c.ResourceHeuristic {
			flag = "  # heuristic, verify"
		}
		fmt.Fprintf(&b, "  %s:%s\n", c.Name, flag)
		writeOp(&b, "create", c.Resource.Create)
		writeOp(&b, "read", c.Resource.Read)
		writeOp(&b, "update", c.Resource.Update)
		writeOp(&b, "delete", c.Resource.Delete)
		ignores := dedupe(append([]string{"status", "record_id", "data"}, c.ExtraIgnores...))
		fmt.Fprintf(&b, "    schema:\n      ignores: [%s]\n", strings.Join(ignores, ", "))
	}

	b.WriteString("\ndata_sources:\n")
	for _, c := range configs {
		if c.DataSourceRead == nil {
			continue
		}
		fmt.Fprintf(&b, "  %s:\n", c.Name)
		writeOp(&b, "read", c.DataSourceRead)
		writeMatch(&b, c.DataSourceReadMatch)
		dsIgnores := dedupe(append([]string{"status", "record_id"}, c.DataSourceExtraIgnores...))
		fmt.Fprintf(&b, "    schema:\n      ignores: [%s]\n", strings.Join(dsIgnores, ", "))
	}

	var impl, stub, incomplete, ds int
	for _, c := range configs {
		if c.Resource != nil {
			switch {
			case c.Resource.Create == nil || c.Resource.Read == nil:
				incomplete++ // missing create/read -> commented out, not generated
			case c.ResourceHeuristic || c.Overridden:
				stub++ // no SDK code yet (heuristic guess or override-supplied paths)
			default:
				impl++ // CRUD actually calls the SDK
			}
		}
		if c.DataSourceRead != nil {
			ds++
		}
	}
	fmt.Fprintf(&b, "\n# summary: %d generatable (%d implemented, %d stub), %d incomplete; %d data sources\n",
		impl+stub, impl, stub, incomplete, ds)
	return b.String()
}

// dedupe removes repeats while preserving first-seen order. Overrides commonly
// restate the baked-in status/record_id/data alongside their own additions, and
// emitting each twice makes the rendered config noisy and its diffs misleading.
func dedupe(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := in[:0:0]
	for _, v := range in {
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

// writeMatch renders a read's match under it, sorted and with every value
// quoted so a number stays a string.
func writeMatch(b *strings.Builder, match map[string]string) {
	if len(match) == 0 {
		return
	}
	keys := make([]string, 0, len(match))
	for k := range match {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b.WriteString("      match:\n")
	for _, k := range keys {
		fmt.Fprintf(b, "        %s: %q\n", k, match[k])
	}
}

func writeOp(b *strings.Builder, verb string, op *Op) {
	if op == nil {
		return
	}
	fmt.Fprintf(b, "    %s:\n      path: %s\n      method: %s\n", verb, op.Path, op.Method)
}

// --- check (drift) ---

type yOp struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
	// Match is valid only on a data source's parent-scoped read; see
	// validateMatches.
	Match map[string]string `yaml:"match"`
}
type yEntry struct {
	Create yOp `yaml:"create"`
	Read   yOp `yaml:"read"`
}
type yCfg struct {
	Resources   map[string]yEntry `yaml:"resources"`
	DataSources map[string]yEntry `yaml:"data_sources"`
}

// Check compares the committed config against what the code implies and reports
// drift to w, returning the number of drifts found (0 == in sync).
func Check(src Source, opts Options, w io.Writer) (int, error) {
	configs, err := Derive(src, opts)
	if err != nil {
		return 0, err
	}
	if configs, err = withOverrides(configs, opts); err != nil {
		return 0, err
	}
	cfgPath := orDefault(opts.ConfigPath, "codegen/generator_config.yaml")
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		return 0, fmt.Errorf("reading %s: %w", cfgPath, err)
	}
	var cur yCfg
	if err := yaml.Unmarshal(raw, &cur); err != nil {
		return 0, fmt.Errorf("parsing %s: %w", cfgPath, err)
	}

	derived := make(map[string]ServiceConfig, len(configs))
	for _, c := range configs {
		derived[c.Name] = c
	}

	// emit writes a report line, remembering the first write error so the caller
	// can surface it once rather than after every line.
	var werr error
	emit := func(format string, a ...any) {
		if werr != nil {
			return
		}
		_, werr = fmt.Fprintf(w, format, a...)
	}

	problems := 0
	// Implemented (non-heuristic) resources missing from, or stale in, the config.
	for _, c := range configs {
		if c.Resource == nil || c.ResourceHeuristic {
			continue
		}
		ce, ok := cur.Resources[c.Name]
		if !ok {
			emit("missing: resource %q is implemented but not in the config\n", c.Name)
			problems++
			continue
		}
		if c.Resource.Create != nil && ce.Create.Path != "" && ce.Create.Path != c.Resource.Create.Path {
			emit("stale: resource %q create path config=%s code=%s\n", c.Name, ce.Create.Path, c.Resource.Create.Path)
			problems++
		}
	}
	// Config resources that no longer map to a service package. Derive only sees
	// packages whose CRUD calls the SDK, so a resource served entirely over raw
	// HTTP (internal/conns, declared in codegen/private_endpoints.yaml) resolves
	// no ops and is absent from derived even though its package is right there.
	// funding_source_note is one, and it was reported extra for exactly that
	// reason. Check the claim being made: does the directory exist?
	serviceRoot := orDefault(opts.ServiceRoot, "internal/service")
	for name := range cur.Resources {
		if _, ok := derived[name]; ok {
			continue
		}
		if fi, err := os.Stat(filepath.Join(serviceRoot, name)); err == nil && fi.IsDir() {
			continue
		}
		emit("extra: config resource %q has no matching service package\n", name)
		problems++
	}
	if problems == 0 {
		emit("config in sync with code\n")
	}
	return problems, werr
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// --- overrides: hand-tuned entries the derivation can't reproduce ---

type ovEntry struct {
	Create  *yOp     `yaml:"create"`
	Read    *yOp     `yaml:"read"`
	Update  *yOp     `yaml:"update"`
	Delete  *yOp     `yaml:"delete"`
	Ignores []string `yaml:"ignores"` // extra OAS attributes to drop (appended to the baked-in ignores)

	// Skip drops an entry derivation produced but that should not exist. Needed
	// because derivation picks the first GET a package's code touches, which for
	// a polymorphic data source is whichever entity it happened to reach first.
	// codegen/import_manifest.json takes its list paths from data_sources, so a
	// wrong entry there is not cosmetic.
	Skip bool `yaml:"skip"`

	// Schema accepts the same `schema: { ignores: [...] }` shape that the rendered
	// generator_config.yaml uses. Authors reasonably mirror the output format when
	// writing an override, and before this field existed those entries parsed
	// cleanly and were then discarded, the ignores never reached the generator and
	// the dropped attributes silently reappeared in the schema. Both spellings are
	// honored and merged; see allIgnores.
	Schema *ovSchema `yaml:"schema"`
}

type ovSchema struct {
	Ignores []string `yaml:"ignores"`
}

// allIgnores returns the entry's extra ignores from either spelling.
func (e ovEntry) allIgnores() []string {
	out := append([]string(nil), e.Ignores...)
	if e.Schema != nil {
		out = append(out, e.Schema.Ignores...)
	}
	return out
}

type overridesFile struct {
	Resources   map[string]ovEntry `yaml:"resources"`
	DataSources map[string]ovEntry `yaml:"data_sources"`
}

func (o *yOp) toOp() *Op {
	if o == nil || o.Path == "" {
		return nil
	}
	return &Op{Method: o.Method, Path: o.Path}
}

// withOverrides reads the (optional) overrides file and overlays it onto the
// derived configs so hand-tuned entries survive regeneration.
func withOverrides(configs []ServiceConfig, opts Options) ([]ServiceConfig, error) {
	path := orDefault(opts.Overrides, defaultOverrides)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return configs, nil
		}
		return nil, fmt.Errorf("reading overrides %s: %w", path, err)
	}
	var ov overridesFile
	// KnownFields: a misspelled or misplaced key in this file used to be dropped
	// in silence, which is how `schema: { ignores: ... }` entries went unapplied
	// for four resources without any signal. Fail the generate instead.
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&ov); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parsing overrides %s: %w", path, err)
	}
	if err := validateMatches(ov); err != nil {
		return nil, fmt.Errorf("overrides %s: %w", path, err)
	}
	return applyOverrides(configs, ov), nil
}

// validateMatches rejects a `match` that could never be applied: one on a
// resource (only data source reads are list reads), one on a read with no
// parent placeholder, or one with an empty map, field or value.
func validateMatches(ov overridesFile) error {
	for name, e := range ov.Resources {
		for verb, op := range map[string]*yOp{"create": e.Create, "read": e.Read, "update": e.Update, "delete": e.Delete} {
			if op != nil && op.Match != nil {
				return fmt.Errorf("resource %q %s: match is only valid on a data source read", name, verb)
			}
		}
	}
	for name, e := range ov.DataSources {
		if e.Read == nil || e.Read.Match == nil {
			continue
		}
		if len(e.Read.Match) == 0 {
			return fmt.Errorf("data source %q read: match is empty", name)
		}
		if !strings.Contains(e.Read.Path, "{") || strings.HasSuffix(e.Read.Path, "}") {
			return fmt.Errorf("data source %q read: match needs a parent-scoped path (/parent/{id}/child), got %q", name, e.Read.Path)
		}
		for k, v := range e.Read.Match {
			if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
				return fmt.Errorf("data source %q read: match field and value must be non-empty, got %q: %q", name, k, v)
			}
		}
	}
	return nil
}

// applyOverrides overlays overrides onto the derived configs: overridden verbs
// win, an overridden resource is no longer treated as heuristic, and entries
// present only in the overrides are added.
func applyOverrides(configs []ServiceConfig, ov overridesFile) []ServiceConfig {
	m := make(map[string]ServiceConfig, len(configs))
	for _, c := range configs {
		m[c.Name] = c
	}
	for name, e := range ov.Resources {
		sc := m[name]
		sc.Name = name
		if sc.Resource == nil {
			sc.Resource = &CRUD{}
		}
		if op := e.Create.toOp(); op != nil {
			sc.Resource.Create = op
		}
		if op := e.Read.toOp(); op != nil {
			sc.Resource.Read = op
		}
		if op := e.Update.toOp(); op != nil {
			sc.Resource.Update = op
		}
		if op := e.Delete.toOp(); op != nil {
			sc.Resource.Delete = op
		}
		sc.ResourceHeuristic = false // human-verified via override (no "# heuristic" annotation)
		sc.Overridden = true         // but its CRUD is not code-implemented
		sc.ExtraIgnores = append(sc.ExtraIgnores, e.allIgnores()...)
		m[name] = sc
	}
	for name, e := range ov.DataSources {
		sc := m[name]
		sc.Name = name
		if e.Skip {
			sc.DataSourceRead = nil
			m[name] = sc
			continue
		}
		if op := e.Read.toOp(); op != nil {
			sc.DataSourceRead = op
			sc.DataSourceReadMatch = e.Read.Match
		}
		sc.DataSourceExtraIgnores = append(sc.DataSourceExtraIgnores, e.allIgnores()...)
		m[name] = sc
	}
	out := make([]ServiceConfig, 0, len(m))
	for _, c := range m {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// orRaw prefers a typed op, falling back to one reconstructed from a raw HTTP
// call. Without it a resource using the raw_create archetype looked as though it
// had no create at all, and the derived config reported it INCOMPLETE.
func orRaw(typed, raw *Op) *Op {
	if typed != nil {
		return typed
	}
	return raw
}
