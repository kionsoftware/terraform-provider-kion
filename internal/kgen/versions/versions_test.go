package versions

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	fsmocks "terraform-provider-kion/internal/kgen/kfs/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fakeDirEntry is a minimal os.DirEntry for driving the mocked ReadDir.
type fakeDirEntry struct {
	name string
	dir  bool
}

func (f fakeDirEntry) Name() string             { return f.name }
func (f fakeDirEntry) IsDir() bool              { return f.dir }
func (fakeDirEntry) Type() fs.FileMode          { return 0 }
func (fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// --- parseClientOps -------------------------------------------------------

func TestParseClientOps(t *testing.T) {
	src := `package foo

// OUNoteCreate invokes createOUNote operation.
//
// POST /v3/ou-note
func (c *Client) OUNoteCreate(ctx context.Context) error { return nil }

// OUNoteRead invokes readOUNote operation.
//
// GET /v3/ou-note/{id}
func (c *Client) OUNoteRead(ctx context.Context) error { return nil }

// helper is not a top-level client method preceded by an op line.
// GET /v3/orphan
func helper() {}
`
	ops := parseClientOps(src)
	assert.Contains(t, ops, op{method: "POST", path: "/v3/ou-note"})
	assert.Contains(t, ops, op{method: "GET", path: "/v3/ou-note/{id}"})
	// The orphan op is not immediately followed by a *Client method.
	assert.NotContains(t, ops, op{method: "GET", path: "/v3/orphan"})
	assert.Len(t, ops, 2)
}

func TestParseClientOps_lowercaseMethodNormalized(t *testing.T) {
	// opLineRE only matches upper-case methods, so lowercase doc lines are
	// ignored, confirm they don't register.
	src := "// post /v3/thing\nfunc (c *Client) Thing() {}\n"
	ops := parseClientOps(src)
	assert.Empty(t, ops)
}

// --- definingOp -----------------------------------------------------------

func TestDefiningOp(t *testing.T) {
	tests := []struct {
		name   string
		e      entry
		wantOK bool
		want   op
	}{
		{
			name:   "create wins over read",
			e:      entry{Create: &opRef{Path: "/v3/x", Method: "post"}, Read: &opRef{Path: "/v3/x/{id}", Method: "get"}},
			wantOK: true,
			want:   op{method: "POST", path: "/v3/x"},
		},
		{
			name:   "read when no create",
			e:      entry{Read: &opRef{Path: "/v3/x/{id}", Method: "get"}},
			wantOK: true,
			want:   op{method: "GET", path: "/v3/x/{id}"},
		},
		{
			name:   "neither",
			e:      entry{Update: &opRef{Path: "/v3/x", Method: "put"}},
			wantOK: false,
		},
		{
			name:   "create present but empty fields",
			e:      entry{Create: &opRef{Path: "", Method: "post"}},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := definingOp(tt.e)
			require.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// --- computeRange ---------------------------------------------------------

func TestComputeRange(t *testing.T) {
	// trackedVersions = [v3_14, v3_15, v3_16, v3_17] (indices 0..3).
	tests := []struct {
		name       string
		present    []bool
		resolved   bool
		emit       bool
		contiguous bool
		min, max   string
	}{
		{
			name:     "unresolved (in no version)",
			present:  []bool{false, false, false, false},
			resolved: false,
		},
		{
			name:       "fully supported oldest..open => no emit",
			present:    []bool{true, true, true, true},
			resolved:   true,
			emit:       false,
			contiguous: true,
		},
		{
			name:       "min-only: introduced in 3.17, still current",
			present:    []bool{false, false, false, true},
			resolved:   true,
			emit:       true,
			contiguous: true,
			min:        "3.17.0",
			max:        "", // open (newest)
		},
		{
			name:       "min introduced mid, still current",
			present:    []bool{false, false, true, true},
			resolved:   true,
			emit:       true,
			contiguous: true,
			min:        "3.16.0",
			max:        "",
		},
		{
			name:       "min+max bounded: dropped after 3.15",
			present:    []bool{true, true, false, false},
			resolved:   true,
			emit:       true,
			contiguous: true,
			min:        "3.14.0",
			max:        "3.15.0",
		},
		{
			name:       "bounded window in the middle",
			present:    []bool{false, true, true, false},
			resolved:   true,
			emit:       true,
			contiguous: true,
			min:        "3.15.0",
			max:        "3.16.0",
		},
		{
			name:       "non-contiguous uses overall min/max",
			present:    []bool{true, false, true, false},
			resolved:   true,
			emit:       true,
			contiguous: false,
			min:        "3.14.0",
			max:        "3.16.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := computeRange(tt.present)
			require.Equal(t, tt.resolved, r.resolved)
			if !tt.resolved {
				return
			}
			assert.Equal(t, tt.emit, r.emit)
			assert.Equal(t, tt.contiguous, r.contiguous)
			if tt.emit {
				// support fields are only meaningful for gated (emitted) entries.
				assert.Equal(t, tt.min, r.support.Min)
				assert.Equal(t, tt.max, r.support.Max)
			}
		})
	}
}

// --- mergeOverrides -------------------------------------------------------

func TestMergeOverrides(t *testing.T) {
	base := &config{
		Resources: map[string]entry{
			"a": {Create: &opRef{Path: "/v3/a", Method: "post"}, Read: &opRef{Path: "/v3/a/{id}", Method: "get"}},
		},
		DataSources: map[string]entry{},
	}
	override := &config{
		Resources: map[string]entry{
			// Override only Create; Read must be preserved from base.
			"a": {Create: &opRef{Path: "/v3/a-new", Method: "post"}},
			// New entry not in base.
			"b": {Read: &opRef{Path: "/v3/b", Method: "get"}},
		},
		DataSources: map[string]entry{},
	}
	merged := mergeOverrides(base, override)
	assert.Equal(t, "/v3/a-new", merged.Resources["a"].Create.Path)
	require.NotNil(t, merged.Resources["a"].Read)
	assert.Equal(t, "/v3/a/{id}", merged.Resources["a"].Read.Path)
	assert.Equal(t, "/v3/b", merged.Resources["b"].Read.Path)
}

// --- deriveSection --------------------------------------------------------

func versionOpsFrom(present map[string][]bool) []map[op]struct{} {
	ops := make([]map[op]struct{}, len(trackedVersions))
	for i := range ops {
		ops[i] = map[op]struct{}{}
	}
	for opStr, pres := range present {
		// opStr is "METHOD path".
		var o op
		for j := 0; j < len(opStr); j++ {
			if opStr[j] == ' ' {
				o = op{method: opStr[:j], path: opStr[j+1:]}
				break
			}
		}
		for i, p := range pres {
			if p {
				ops[i][o] = struct{}{}
			}
		}
	}
	return ops
}

func TestDeriveSection(t *testing.T) {
	entries := map[string]entry{
		// bounded: only in 3.12,3.13 => emit min+max
		"dropped": {Create: &opRef{Path: "/v3/dropped", Method: "post"}},
		// fully supported => skipped
		"full": {Create: &opRef{Path: "/v3/full", Method: "post"}},
		// min-only => emit
		"new": {Create: &opRef{Path: "/v3/new", Method: "post"}},
		// no op => skipped
		"noop": {Update: &opRef{Path: "/v3/noop", Method: "put"}},
		// op in no version => skipped
		"missing": {Create: &opRef{Path: "/v3/missing", Method: "post"}},
	}
	versionOps := versionOpsFrom(map[string][]bool{
		"POST /v3/dropped": {true, true, false, false},
		"POST /v3/full":    {true, true, true, true},
		"POST /v3/new":     {false, false, false, true},
	})

	var log bytes.Buffer
	out, err := deriveSection("resources", entries, versionOps, nil, &log)
	require.NoError(t, err)

	// Sorted by name; only "dropped" and "new" emit.
	require.Len(t, out, 2)
	assert.Equal(t, "dropped", out[0].name)
	assert.Equal(t, "3.14.0", out[0].support.Min)
	assert.Equal(t, "3.15.0", out[0].support.Max)
	assert.Equal(t, "new", out[1].name)
	assert.Equal(t, "3.17.0", out[1].support.Min)
	assert.Equal(t, "", out[1].support.Max)

	logStr := log.String()
	assert.Contains(t, logStr, "noop")
	assert.Contains(t, logStr, "missing")
}

// --- renderVersionGen -----------------------------------------------------

func TestRenderVersionGen_minOnly(t *testing.T) {
	out, err := renderVersionGen("ou_note", support{Min: "3.16.0"}, false, nil)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "// Code generated by `kgen versions`; DO NOT EDIT.")
	assert.Contains(t, s, "package ou_note")
	assert.Contains(t, s, `import "terraform-provider-kion/internal/conns"`)
	assert.Contains(t, s, `minKionVersion = conns.MustParseKionVersion("3.16.0")`)
	assert.Contains(t, s, "maxKionVersion = conns.KionVersion{} // unbounded")
}

func TestRenderVersionGen_minAndMax(t *testing.T) {
	out, err := renderVersionGen("thing", support{Min: "3.12.0", Max: "3.13.0"}, false, nil)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, `minKionVersion = conns.MustParseKionVersion("3.12.0")`)
	assert.Contains(t, s, `maxKionVersion = conns.MustParseKionVersion("3.13.0")`)
	assert.NotContains(t, s, "unbounded")
}

func TestRenderVersionGen_unboundedMin(t *testing.T) {
	// Shouldn't happen for gated entries, but the code path is specified.
	out, err := renderVersionGen("thing", support{Min: "", Max: "3.13.0"}, false, nil)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "minKionVersion = conns.KionVersion{}")
	assert.Contains(t, s, `maxKionVersion = conns.MustParseKionVersion("3.13.0")`)
}

// --- generate end-to-end against a mocked kfs.FS --------------------------

// clientSrc builds a minimal oas_client_gen.go exposing the given ops.
func clientSrc(ops ...op) []byte {
	var b bytes.Buffer
	b.WriteString("package client\n\n")
	for i, o := range ops {
		b.WriteString("// Op" + string(rune('A'+i)) + " invokes op.\n//\n")
		b.WriteString("// " + o.method + " " + o.path + "\n")
		b.WriteString("func (c *Client) Op" + string(rune('A'+i)) + "() {}\n\n")
	}
	return b.Bytes()
}

func TestGenerate_endToEnd(t *testing.T) {
	m := fsmocks.NewMockFS(t)
	m.EXPECT().ReadFile("codegen/attr_version_overrides.yaml").Return(nil, os.ErrNotExist).Maybe()

	configYAML := `
resources:
  ou_note:
    create:
      path: /v3/ou-note
      method: POST
  full_thing:
    create:
      path: /v3/full
      method: POST
data_sources:
  legacy:
    read:
      path: /v3/legacy/{id}
      method: GET
`
	m.EXPECT().ReadFile("cfg/generator_config.yaml").Return([]byte(configYAML), nil)
	// Overrides missing => no-op.
	m.EXPECT().ReadFile("cfg/config_overrides.yaml").Return(nil, os.ErrNotExist)

	// Per-version clients:
	//  ou-note: only in v3_17 (min-only, gated)
	//  full:    in every version (fully supported, skipped)
	//  legacy:  only in v3_14,v3_15 (bounded, gated)
	ouNote := op{method: "POST", path: "/v3/ou-note"}
	full := op{method: "POST", path: "/v3/full"}
	legacy := op{method: "GET", path: "/v3/legacy/{id}"}

	presence := map[string][]op{
		"v3_14": {full, legacy},
		"v3_15": {full, legacy},
		"v3_16": {full},
		"v3_17": {full, ouNote},
	}
	for _, v := range trackedVersions {
		path := filepath.Join("sdk", "generated", v.dir, "oas_client_gen.go")
		m.EXPECT().ReadFile(path).Return(clientSrc(presence[v.dir]...), nil)
	}

	// Collision scan: service dirs don't yet exist => no inline var.
	m.EXPECT().ReadDir(mock.Anything).Return(nil, os.ErrNotExist)
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(nil)
	writes := map[string]string{}
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		Run(func(name string, data []byte, _ os.FileMode) { writes[name] = string(data) }).
		Return(nil)

	g := &generator{fs: m}
	n, err := g.generate(Options{
		SDKDir:      "sdk",
		ServiceRoot: "svc",
		ConfigPath:  "cfg/generator_config.yaml",
		Overrides:   "cfg/config_overrides.yaml",
	})
	require.NoError(t, err)
	require.Equal(t, 2, n)

	// ou_note: min-only gate.
	ouPath := filepath.Join("svc", "ou_note", "ou_note_version_gen.go")
	require.Contains(t, writes, ouPath)
	assert.Contains(t, writes[ouPath], "package ou_note")
	assert.Contains(t, writes[ouPath], `minKionVersion = conns.MustParseKionVersion("3.17.0")`)
	assert.Contains(t, writes[ouPath], "maxKionVersion = conns.KionVersion{} // unbounded")

	// legacy: bounded min+max gate.
	legacyPath := filepath.Join("svc", "legacy", "legacy_version_gen.go")
	require.Contains(t, writes, legacyPath)
	assert.Contains(t, writes[legacyPath], `minKionVersion = conns.MustParseKionVersion("3.14.0")`)
	assert.Contains(t, writes[legacyPath], `maxKionVersion = conns.MustParseKionVersion("3.15.0")`)

	// isResource must survive the whole pipeline, not just renderVersionGen:
	// ou_note is a resource and gets a plan-time gate; legacy is a data source
	// and must not, or the generated file would not compile.
	assert.Contains(t, writes[ouPath], "func (r *ou_noteResource) ModifyPlan(")
	assert.Contains(t, writes[ouPath], "resource.ResourceWithModifyPlan")
	assert.NotContains(t, writes[legacyPath], "ModifyPlan")

	// full_thing: fully supported, no file written.
	assert.NotContains(t, writes, filepath.Join("svc", "full_thing", "full_thing_version_gen.go"))
}

// TestGenerate_skipsInlineVarCollision verifies a service that already declares
// an inline var minKionVersion is skipped (not overwritten) so the build stays
// green, while a clean service still gets its generated file.
func TestGenerate_skipsInlineVarCollision(t *testing.T) {
	m := fsmocks.NewMockFS(t)
	m.EXPECT().ReadFile("codegen/attr_version_overrides.yaml").Return(nil, os.ErrNotExist).Maybe()

	configYAML := `
resources:
  ou_note:
    create:
      path: /v3/ou-note
      method: POST
  billing_rule:
    create:
      path: /v3/billing-rule
      method: POST
`
	m.EXPECT().ReadFile("cfg.yml").Return([]byte(configYAML), nil)
	m.EXPECT().ReadFile("ov.yml").Return(nil, os.ErrNotExist)

	ouNote := op{method: "POST", path: "/v3/ou-note"}
	billing := op{method: "POST", path: "/v3/billing-rule"}
	presence := map[string][]op{"v3_17": {ouNote, billing}} // both min-only, gated
	for _, v := range trackedVersions {
		path := filepath.Join("sdk", "generated", v.dir, "oas_client_gen.go")
		m.EXPECT().ReadFile(path).Return(clientSrc(presence[v.dir]...), nil)
	}

	// ou_note package already declares the inline var => collision, skip.
	ouDir := filepath.Join("svc", "ou_note")
	m.EXPECT().ReadDir(ouDir).Return([]os.DirEntry{fakeDirEntry{name: "ou_note.go"}}, nil)
	m.EXPECT().ReadFile(filepath.Join(ouDir, "ou_note.go")).
		Return([]byte("package ou_note\n\nvar minKionVersion = conns.MustParseKionVersion(\"3.16.0\")\n"), nil)

	// billing_rule is clean => write.
	brDir := filepath.Join("svc", "billing_rule")
	m.EXPECT().ReadDir(brDir).Return([]os.DirEntry{fakeDirEntry{name: "billing_rule.go"}}, nil)
	m.EXPECT().ReadFile(filepath.Join(brDir, "billing_rule.go")).
		Return([]byte("package billing_rule\n\nfunc x() {}\n"), nil)
	m.EXPECT().MkdirAll(brDir, mock.Anything).Return(nil)

	var wrotePath string
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		Run(func(name string, _ []byte, _ os.FileMode) { wrotePath = name }).Return(nil)

	g := &generator{fs: m}
	n, err := g.generate(Options{ConfigPath: "cfg.yml", Overrides: "ov.yml", SDKDir: "sdk", ServiceRoot: "svc"})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	assert.Equal(t, filepath.Join(brDir, "billing_rule_version_gen.go"), wrotePath)
}

// TestInlineVersionVarFile covers the collision-scan helper directly.
func TestInlineVersionVarFile(t *testing.T) {
	t.Run("missing dir => no collision", func(t *testing.T) {
		m := fsmocks.NewMockFS(t)
		m.EXPECT().ReadFile("codegen/attr_version_overrides.yaml").Return(nil, os.ErrNotExist).Maybe()
		m.EXPECT().ReadDir("d").Return(nil, os.ErrNotExist)
		g := &generator{fs: m}
		who, err := g.inlineVersionVarFile("d", "x")
		require.NoError(t, err)
		assert.Empty(t, who)
	})
	t.Run("ignores generated + test files, finds inline var", func(t *testing.T) {
		m := fsmocks.NewMockFS(t)
		m.EXPECT().ReadFile("codegen/attr_version_overrides.yaml").Return(nil, os.ErrNotExist).Maybe()
		m.EXPECT().ReadDir("d").Return([]os.DirEntry{
			fakeDirEntry{name: "x_version_gen.go"},
			fakeDirEntry{name: "x_test.go"},
			fakeDirEntry{name: "sub", dir: true},
			fakeDirEntry{name: "x.go"},
		}, nil)
		m.EXPECT().ReadFile(filepath.Join("d", "x.go")).
			Return([]byte("package x\nvar maxKionVersion = conns.KionVersion{}\n"), nil)
		g := &generator{fs: m}
		who, err := g.inlineVersionVarFile("d", "x")
		require.NoError(t, err)
		assert.Equal(t, "x.go", who)
	})
	t.Run("no inline var => empty", func(t *testing.T) {
		m := fsmocks.NewMockFS(t)
		m.EXPECT().ReadFile("codegen/attr_version_overrides.yaml").Return(nil, os.ErrNotExist).Maybe()
		m.EXPECT().ReadDir("d").Return([]os.DirEntry{fakeDirEntry{name: "x.go"}}, nil)
		m.EXPECT().ReadFile(filepath.Join("d", "x.go")).Return([]byte("package x\nfunc y(){}\n"), nil)
		g := &generator{fs: m}
		who, err := g.inlineVersionVarFile("d", "x")
		require.NoError(t, err)
		assert.Empty(t, who)
	})
}

func TestGenerate_configReadError(t *testing.T) {
	m := fsmocks.NewMockFS(t)
	m.EXPECT().ReadFile("codegen/attr_version_overrides.yaml").Return(nil, os.ErrNotExist).Maybe()
	m.EXPECT().ReadFile("cfg.yml").Return(nil, os.ErrPermission)

	g := &generator{fs: m}
	_, err := g.generate(Options{ConfigPath: "cfg.yml", Overrides: "ov.yml", SDKDir: "sdk", ServiceRoot: "svc"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading generator config")
}

func TestGenerate_sdkReadError(t *testing.T) {
	m := fsmocks.NewMockFS(t)
	m.EXPECT().ReadFile("codegen/attr_version_overrides.yaml").Return(nil, os.ErrNotExist).Maybe()
	m.EXPECT().ReadFile("cfg.yml").Return([]byte("resources: {}\n"), nil)
	m.EXPECT().ReadFile("ov.yml").Return(nil, os.ErrNotExist)
	// First version client read fails.
	m.EXPECT().ReadFile(filepath.Join("sdk", "generated", "v3_14", "oas_client_gen.go")).
		Return(nil, os.ErrNotExist)

	g := &generator{fs: m}
	_, err := g.generate(Options{ConfigPath: "cfg.yml", Overrides: "ov.yml", SDKDir: "sdk", ServiceRoot: "svc"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading client for v3_14")
}

func TestGenerate_overridesMerged(t *testing.T) {
	m := fsmocks.NewMockFS(t)
	m.EXPECT().ReadFile("codegen/attr_version_overrides.yaml").Return(nil, os.ErrNotExist).Maybe()
	base := `
resources:
  thing:
    create:
      path: /v3/old
      method: POST
`
	overrides := `
resources:
  thing:
    create:
      path: /v3/new
      method: POST
`
	m.EXPECT().ReadFile("cfg.yml").Return([]byte(base), nil)
	m.EXPECT().ReadFile("ov.yml").Return([]byte(overrides), nil)

	newOp := op{method: "POST", path: "/v3/new"}
	presence := map[string][]op{
		"v3_17": {newOp}, // only newest => min-only gate at 3.17
	}
	for _, v := range trackedVersions {
		path := filepath.Join("sdk", "generated", v.dir, "oas_client_gen.go")
		m.EXPECT().ReadFile(path).Return(clientSrc(presence[v.dir]...), nil)
	}

	m.EXPECT().ReadDir(mock.Anything).Return(nil, os.ErrNotExist)
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(nil)
	var written string
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ os.FileMode) { written = string(data) }).
		Return(nil)

	g := &generator{fs: m}
	n, err := g.generate(Options{ConfigPath: "cfg.yml", Overrides: "ov.yml", SDKDir: "sdk", ServiceRoot: "svc"})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	// Confirms the override's /v3/new op (present only in v3_16) drove the range.
	assert.Contains(t, written, `minKionVersion = conns.MustParseKionVersion("3.17.0")`)
}

// A resource gets a plan-time gate; a data source must not, both because it has
// no plan and because its struct is not named <name>DataSource (gcpRegionsDataSource).
func TestRenderVersionGen_resourceGetsModifyPlan(t *testing.T) {
	out, err := renderVersionGen("ou_note", support{Min: "3.16.0"}, true, nil)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "var _ resource.ResourceWithModifyPlan = &ou_noteResource{}")
	assert.Contains(t, s, "func (r *ou_noteResource) ModifyPlan(")
	assert.Contains(t, s, `"kion_ou_note"`)
	// Destroy plans must stay allowed, or an unsupported instance strands state.
	assert.Contains(t, s, "if req.Plan.Raw.IsNull() {")
	// Both halves of the gate, or dropping one from the template goes unnoticed.
	assert.Contains(t, s, "framework.RequireKionVersionInRange(")
	assert.Contains(t, s, "framework.RequireAttrKionVersions(")
	assert.Contains(t, s, "attrKionVersions")
}

func TestRenderVersionGen_dataSourceHasNoModifyPlan(t *testing.T) {
	out, err := renderVersionGen("gcp_regions", support{Min: "3.16.0"}, false, nil)
	require.NoError(t, err)
	s := string(out)
	assert.NotContains(t, s, "ModifyPlan")
	assert.NotContains(t, s, "gcp_regionsResource")
}
