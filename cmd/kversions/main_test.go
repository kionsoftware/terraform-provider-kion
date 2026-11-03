package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseClientOps(t *testing.T) {
	src := `package generated

// Client is the API client.
type Client struct{}

// POST /v3/ou-note
func (c *Client) PostOUNote(ctx context.Context, request *OUNoteCreate) (PostOUNoteRes, error) {
	return nil, nil
}

// GET /v3/ou-note/{id}
func (c *Client) GetOUNote(ctx context.Context, params GetOUNoteParams) (GetOUNoteRes, error) {
	return nil, nil
}

// POST /v3/account-cache/create/__qs__/account-type/aws
func (c *Client) PostAccountCacheAWS(ctx context.Context) error {
	return nil
}

// this doc line is not adjacent to the func below

func (c *Client) NotAnOp(ctx context.Context) error { return nil }

// a lower-case receiver method must be ignored
// PATCH /v3/should-not-count
func (c *client) internalHelper() {}
`

	ops := parseClientOps(src)

	require.Contains(t, ops, op{method: "POST", path: "/v3/ou-note"})
	require.Contains(t, ops, op{method: "GET", path: "/v3/ou-note/{id}"})
	require.Contains(t, ops, op{method: "POST", path: "/v3/account-cache/create/__qs__/account-type/aws"})

	// A doc line separated from its func by a blank line is not an op.
	require.NotContains(t, ops, op{method: "PATCH", path: "/v3/should-not-count"})

	require.Len(t, ops, 3)
}

func TestParseClientOps_MethodUppercased(t *testing.T) {
	// The regex only accepts upper-case verbs, but ToUpper is applied so the
	// stored key is always upper. Confirm the canonical form.
	src := `// DELETE /v3/thing/{id}
func (c *Client) DeleteThing() {}
`
	ops := parseClientOps(src)
	require.Contains(t, ops, op{method: "DELETE", path: "/v3/thing/{id}"})
}

func TestVersionString(t *testing.T) {
	cases := []struct {
		v    version
		want string
	}{
		{version{dir: "v3_14", minor: 14}, "3.14.0"},
		{version{dir: "v3_17", minor: 17}, "3.17.0"},
		{version{dir: "v3_15", minor: 15}, "3.15.0"},
	}
	for _, c := range cases {
		require.Equal(t, c.want, versionString(c.v))
	}
}

func TestTrackedVersionsMapping(t *testing.T) {
	// Guard the assumption used throughout: 4 tracked versions v3_14..v3_17,
	// the current Kion release plus three back.
	require.Len(t, trackedVersions, 4)
	require.Equal(t, "3.14.0", versionString(trackedVersions[0]))
	require.Equal(t, "3.17.0", versionString(trackedVersions[len(trackedVersions)-1]))
}

func TestComputeRange(t *testing.T) {
	// Presence slices are aligned with trackedVersions: [v3_14, v3_15, v3_16, v3_17].
	cases := []struct {
		name       string
		present    []bool
		resolved   bool
		contiguous bool
		emit       bool
		min        string
		max        string
	}{
		{
			name:     "not present anywhere -> unresolved",
			present:  []bool{false, false, false, false},
			resolved: false,
		},
		{
			name:       "new endpoint only in newest -> min 3.17.0, open max, emit",
			present:    []bool{false, false, false, true},
			resolved:   true,
			contiguous: true,
			emit:       true,
			min:        "3.17.0",
			max:        "",
		},
		{
			name:       "dropped endpoint present only in v3_14,v3_15 -> min 3.14.0 max 3.15.0",
			present:    []bool{true, true, false, false},
			resolved:   true,
			contiguous: true,
			emit:       true,
			min:        "3.14.0",
			max:        "3.15.0",
		},
		{
			name:       "fully supported all versions -> no gate, emit false",
			present:    []bool{true, true, true, true},
			resolved:   true,
			contiguous: true,
			emit:       false,
			min:        "3.14.0",
			max:        "",
		},
		{
			name:       "added mid-range and still current -> min 3.16.0 open max, emit",
			present:    []bool{false, false, true, true},
			resolved:   true,
			contiguous: true,
			emit:       true,
			min:        "3.16.0",
			max:        "",
		},
		{
			name:       "non-contiguous still uses overall min/max and warns",
			present:    []bool{true, false, true, false},
			resolved:   true,
			contiguous: false,
			emit:       true,
			min:        "3.14.0",
			max:        "3.16.0",
		},
		{
			name:       "dropped just before newest -> min 3.14.0 max 3.16.0",
			present:    []bool{true, true, true, false},
			resolved:   true,
			contiguous: true,
			emit:       true,
			min:        "3.14.0",
			max:        "3.16.0",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := computeRange(c.present)
			require.Equal(t, c.resolved, r.resolved)
			if !c.resolved {
				return
			}
			require.Equal(t, c.contiguous, r.contiguous)
			require.Equal(t, c.emit, r.emit)
			require.Equal(t, c.min, r.support.Min)
			require.Equal(t, c.max, r.support.Max)
		})
	}
}

func TestDefiningOp(t *testing.T) {
	create := &opRef{Path: "/v3/thing", Method: "post"}
	read := &opRef{Path: "/v3/thing/{id}", Method: "get"}

	// create wins over read.
	o, ok := definingOp(entry{Create: create, Read: read})
	require.True(t, ok)
	require.Equal(t, op{method: "POST", path: "/v3/thing"}, o)

	// read used when no create.
	o, ok = definingOp(entry{Read: read})
	require.True(t, ok)
	require.Equal(t, op{method: "GET", path: "/v3/thing/{id}"}, o)

	// neither -> not ok.
	_, ok = definingOp(entry{})
	require.False(t, ok)

	// present but empty fields -> not ok.
	_, ok = definingOp(entry{Create: &opRef{}})
	require.False(t, ok)
}

func TestMergeOverrides(t *testing.T) {
	base := &config{
		Resources: map[string]entry{
			"thing": {
				Read:   &opRef{Path: "/v3/thing/{id}", Method: "GET"},
				Delete: &opRef{Path: "/v3/thing/{id}", Method: "DELETE"},
			},
		},
		DataSources: map[string]entry{},
	}
	override := &config{
		Resources: map[string]entry{
			"thing": {
				// Override the read op only; delete must be preserved.
				Read: &opRef{Path: "/beta/thing/{id}", Method: "GET"},
			},
			"brand_new": {
				// A brand-new entry exercises the create/update/delete override branches.
				Create: &opRef{Path: "/v3/brand-new", Method: "POST"},
				Read:   &opRef{Path: "/v3/brand-new/{id}", Method: "GET"},
				Update: &opRef{Path: "/v3/brand-new/{id}", Method: "PATCH"},
				Delete: &opRef{Path: "/v3/brand-new/{id}", Method: "DELETE"},
			},
		},
		DataSources: map[string]entry{},
	}

	merged := mergeOverrides(base, override)

	// Overridden read wins.
	require.Equal(t, "/beta/thing/{id}", merged.Resources["thing"].Read.Path)
	// Non-overridden op preserved.
	require.NotNil(t, merged.Resources["thing"].Delete)
	require.Equal(t, "/v3/thing/{id}", merged.Resources["thing"].Delete.Path)
	// New resource from override is added, with all four ops applied.
	bn := merged.Resources["brand_new"]
	require.Equal(t, "/v3/brand-new", bn.Create.Path)
	require.Equal(t, "/v3/brand-new/{id}", bn.Read.Path)
	require.Equal(t, "/v3/brand-new/{id}", bn.Update.Path)
	require.Equal(t, "/v3/brand-new/{id}", bn.Delete.Path)
}

func TestDeriveSectionAndMarshal(t *testing.T) {
	entries := map[string]entry{
		"ou_note":     {Create: &opRef{Path: "/v3/ou-note", Method: "POST"}},     // only newest
		"forecast_v2": {Create: &opRef{Path: "/v3/forecast-v2", Method: "POST"}}, // dropped
		"account":     {Read: &opRef{Path: "/v3/account/{id}", Method: "GET"}},   // all versions -> omitted
		"ghost":       {Read: &opRef{Path: "/v3/does-not-exist", Method: "GET"}}, // unresolved
	}

	// Build per-version op sets.
	ouNote := op{method: "POST", path: "/v3/ou-note"}
	forecast := op{method: "POST", path: "/v3/forecast-v2"}
	account := op{method: "GET", path: "/v3/account/{id}"}

	versionOps := make([]map[op]struct{}, len(trackedVersions))
	for i := range versionOps {
		versionOps[i] = map[op]struct{}{}
		versionOps[i][account] = struct{}{} // present everywhere
	}
	// ou-note only in v3_17 (index 3, the newest).
	versionOps[3][ouNote] = struct{}{}
	// forecast-v2 only in v3_14,v3_15 (indexes 0,1).
	versionOps[0][forecast] = struct{}{}
	versionOps[1][forecast] = struct{}{}

	out, err := deriveSection("resources", entries, versionOps, nil, &diag{w: io.Discard})
	require.NoError(t, err)

	// account is fully supported (omitted); ghost is unresolved (omitted).
	byName := map[string]support{}
	for _, e := range out {
		byName[e.name] = e.support
	}
	require.Len(t, out, 2)
	require.Equal(t, support{Min: "3.17.0"}, byName["ou_note"])
	require.Equal(t, support{Min: "3.14.0", Max: "3.15.0"}, byName["forecast_v2"])

	// Marshaled output is deterministic and quotes versions.
	b, err := marshalOutput(out, nil)
	require.NoError(t, err)
	s := string(b)
	require.Contains(t, s, "Generated by")
	require.Contains(t, s, "resources:")
	require.Contains(t, s, `min: "3.17.0"`)
	require.Contains(t, s, `max: "3.15.0"`)
	require.NotContains(t, s, "data_sources:") // none emitted

	// Deterministic ordering: forecast_v2 sorts before ou_note.
	require.Less(t, indexOf(s, "forecast_v2"), indexOf(s, "ou_note"))
}

func indexOf(haystack, needle string) int {
	return strings.Index(haystack, needle)
}

// writeClient writes a minimal oas_client_gen.go under sdkDir for one version,
// exposing exactly the given ops.
func writeClient(t *testing.T, sdkDir, verDir string, ops []op) {
	t.Helper()
	dir := filepath.Join(sdkDir, "generated", verDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	var b strings.Builder
	b.WriteString("package generated\n\ntype Client struct{}\n\n")
	for i, o := range ops {
		fmt.Fprintf(&b, "// %s %s\n", o.method, o.path)
		fmt.Fprintf(&b, "func (c *Client) Op%c() {}\n\n", 'A'+i)
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "oas_client_gen.go"), []byte(b.String()), 0o644))
}

// TestRunEndToEnd exercises run() over a synthetic SDK + config + overrides,
// covering the flag wiring, loading, merging, derivation, and file output.
func TestRunEndToEnd(t *testing.T) {
	root := t.TempDir()
	sdkDir := filepath.Join(root, "sdk")

	ouNote := op{method: "POST", path: "/v3/ou-note"}
	forecast := op{method: "POST", path: "/v3/forecast-v2"}
	account := op{method: "GET", path: "/v3/account/{id}"}
	betaThing := op{method: "GET", path: "/beta/thing/{id}"}

	// Per-version presence:
	//   ou-note: only v3_17 -> min 3.17.0
	//   forecast-v2: v3_14,v3_15 -> min 3.14.0 max 3.15.0
	//   account: all -> omitted (fully supported)
	//   beta thing (from override read): all -> omitted
	writeClient(t, sdkDir, "v3_14", []op{forecast, account, betaThing})
	writeClient(t, sdkDir, "v3_15", []op{forecast, account, betaThing})
	writeClient(t, sdkDir, "v3_16", []op{account, betaThing})
	writeClient(t, sdkDir, "v3_17", []op{ouNote, account, betaThing})

	cfgPath := filepath.Join(root, "generator_config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`
resources:
  ou_note:
    create:
      path: /v3/ou-note
      method: POST
  forecast_v2:
    create:
      path: /v3/forecast-v2
      method: POST
  account:
    read:
      path: /v3/account/{id}
      method: GET
  thing:
    read:
      path: /v3/thing/{id}
      method: GET
data_sources:
  gcp_regions:
    read:
      path: /v3/ou-note
      method: POST
`), 0o644))

	ovPath := filepath.Join(root, "config_overrides.yaml")
	require.NoError(t, os.WriteFile(ovPath, []byte(`
resources:
  thing:
    read:
      path: /beta/thing/{id}
      method: GET
`), 0o644))

	outPath := filepath.Join(root, "version_support.yaml")

	code := run([]string{
		"-sdk", sdkDir,
		"-config", cfgPath,
		"-overrides", ovPath,
		"-out", outPath,
	}, io.Discard, io.Discard)
	require.Equal(t, 0, code)

	b, err := os.ReadFile(outPath)
	require.NoError(t, err)

	var doc struct {
		Resources   map[string]support `yaml:"resources"`
		DataSources map[string]support `yaml:"data_sources"`
	}
	require.NoError(t, yaml.Unmarshal(b, &doc))

	require.Equal(t, support{Min: "3.17.0"}, doc.Resources["ou_note"])
	require.Equal(t, support{Min: "3.14.0", Max: "3.15.0"}, doc.Resources["forecast_v2"])
	// account fully supported -> omitted.
	require.NotContains(t, doc.Resources, "account")
	// thing's read was overridden to /beta/thing/{id}, present everywhere -> omitted.
	require.NotContains(t, doc.Resources, "thing")
	// data source gcp_regions points at ou-note op (only v3_16) -> min 3.16.0.
	require.Equal(t, support{Min: "3.17.0"}, doc.DataSources["gcp_regions"])
}

// floorFixture writes a one-resource SDK (op in the named versions), a config
// with a resource and data source named thing, and the given attr overrides.
func floorFixture(t *testing.T, presentIn []string, attrOverrides string) (args []string, outPath string) {
	t.Helper()
	root := t.TempDir()
	sdkDir := filepath.Join(root, "sdk")
	thing := op{method: "POST", path: "/v3/thing"}
	in := map[string]bool{}
	for _, d := range presentIn {
		in[d] = true
	}
	for _, v := range trackedVersions {
		if in[v.dir] {
			writeClient(t, sdkDir, v.dir, []op{thing})
		} else {
			writeClient(t, sdkDir, v.dir, nil)
		}
	}
	cfgPath := filepath.Join(root, "generator_config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`
resources:
  thing:
    create: {path: /v3/thing, method: POST}
data_sources:
  thing:
    read: {path: /v3/thing, method: POST}
`), 0o644))
	attrPath := filepath.Join(root, "attr.yaml")
	require.NoError(t, os.WriteFile(attrPath, []byte(attrOverrides), 0o644))
	outPath = filepath.Join(root, "version_support.yaml")
	return []string{"-sdk", sdkDir, "-config", cfgPath, "-overrides", filepath.Join(root, "none.yaml"),
		"-attr-overrides", attrPath, "-out", outPath}, outPath
}

func readSupport(t *testing.T, path string) (res, ds map[string]support) {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	var doc struct {
		Resources   map[string]support `yaml:"resources"`
		DataSources map[string]support `yaml:"data_sources"`
	}
	require.NoError(t, yaml.Unmarshal(b, &doc))
	return doc.Resources, doc.DataSources
}

func TestRun_resourceFloorRaisesMin(t *testing.T) {
	args, out := floorFixture(t, []string{"v3_17"}, `resource_floors: { thing: "3.17.3" }`)
	require.Equal(t, 0, run(args, io.Discard, io.Discard))
	res, ds := readSupport(t, out)
	require.Equal(t, support{Min: "3.17.3"}, res["thing"])
	require.Equal(t, support{Min: "3.17.3"}, ds["thing"], "the data source takes the floor too")
}

func TestRun_resourceFloorOnUngatedEntryAppears(t *testing.T) {
	args, out := floorFixture(t, []string{"v3_14", "v3_15", "v3_16", "v3_17"}, `resource_floors: { thing: "3.17.3" }`)
	require.Equal(t, 0, run(args, io.Discard, io.Discard))
	res, ds := readSupport(t, out)
	require.Equal(t, support{Min: "3.17.3"}, res["thing"])
	require.Equal(t, support{Min: "3.17.3"}, ds["thing"])
}

func TestRun_resourceFloorErrors(t *testing.T) {
	for name, tc := range map[string]struct{ yaml, want string }{
		"below derived": {`resource_floors: { thing: "3.16.0" }`, "below the derived minimum"},
		"unknown name":  {`resource_floors: { nope: "3.17.3" }`, "neither a resource nor a data source"},
		"malformed":     {`resource_floors: { thing: "3.17" }`, "not a 3.N.P version"},
	} {
		t.Run(name, func(t *testing.T) {
			args, _ := floorFixture(t, []string{"v3_17"}, tc.yaml)
			var errb bytes.Buffer
			require.Equal(t, 1, run(args, io.Discard, &errb))
			require.Contains(t, errb.String(), tc.want)
		})
	}
}

func TestRun_attrOverridesStrictAndReadErrors(t *testing.T) {
	t.Run("unknown top-level key", func(t *testing.T) {
		args, _ := floorFixture(t, []string{"v3_17"}, `resource_floor: { thing: "3.17.3" }`)
		var errb bytes.Buffer
		require.Equal(t, 1, run(args, io.Discard, &errb))
		require.Contains(t, errb.String(), "resource_floor")
	})
	t.Run("unreadable file is not treated as absent", func(t *testing.T) {
		args, _ := floorFixture(t, []string{"v3_17"}, "")
		for i, a := range args {
			if a == "-attr-overrides" {
				args[i+1] = filepath.Dir(args[i+1]) // a directory: read fails, not ErrNotExist
			}
		}
		require.Equal(t, 1, run(args, io.Discard, io.Discard))
	})
}

// TestRunMissingOverridesOK confirms a missing overrides file is non-fatal.
func TestRunMissingOverridesOK(t *testing.T) {
	root := t.TempDir()
	sdkDir := filepath.Join(root, "sdk")
	ouNote := op{method: "POST", path: "/v3/ou-note"}
	for _, v := range []string{"v3_14", "v3_15", "v3_16"} {
		writeClient(t, sdkDir, v, nil)
	}
	writeClient(t, sdkDir, "v3_17", []op{ouNote})

	cfgPath := filepath.Join(root, "generator_config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`
resources:
  ou_note:
    create:
      path: /v3/ou-note
      method: POST
`), 0o644))

	outPath := filepath.Join(root, "out.yml")

	var errb bytes.Buffer
	code := run([]string{
		"-sdk", sdkDir,
		"-config", cfgPath,
		"-overrides", filepath.Join(root, "does-not-exist.yaml"),
		"-out", outPath,
	}, io.Discard, &errb)
	require.Equal(t, 0, code)
	// The missing overrides file is reported as a non-fatal note.
	require.Contains(t, errb.String(), "no overrides file")

	b, err := os.ReadFile(outPath)
	require.NoError(t, err)
	require.Contains(t, string(b), `min: "3.17.0"`)
}

// TestRun_errorPaths covers the flag-parse and load failure exit codes.
func TestRun_errorPaths(t *testing.T) {
	t.Run("bad flag", func(t *testing.T) {
		code := run([]string{"-nope"}, io.Discard, io.Discard)
		require.Equal(t, 2, code)
	})

	t.Run("missing config", func(t *testing.T) {
		var errb bytes.Buffer
		code := run([]string{"-config", filepath.Join(t.TempDir(), "nope.yml")}, io.Discard, &errb)
		require.Equal(t, 1, code)
		require.Contains(t, errb.String(), "loading generator config")
	})

	t.Run("malformed config", func(t *testing.T) {
		cfg := filepath.Join(t.TempDir(), "bad.yml")
		require.NoError(t, os.WriteFile(cfg, []byte("resources: [this is not a map"), 0o644))
		var errb bytes.Buffer
		code := run([]string{"-config", cfg}, io.Discard, &errb)
		require.Equal(t, 1, code)
	})

	t.Run("missing sdk client", func(t *testing.T) {
		root := t.TempDir()
		cfg := filepath.Join(root, "config.yml")
		require.NoError(t, os.WriteFile(cfg, []byte("resources:\n  x:\n    read:\n      path: /v3/x\n      method: GET\n"), 0o644))
		var errb bytes.Buffer
		code := run([]string{
			"-config", cfg,
			"-sdk", filepath.Join(root, "no-sdk"),
			"-overrides", filepath.Join(root, "none.yaml"),
			"-out", filepath.Join(root, "out.yml"),
		}, io.Discard, &errb)
		require.Equal(t, 1, code)
		require.Contains(t, errb.String(), "reading client")
	})
}
