package config_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"terraform-provider-kion/internal/kgen/config"
	"terraform-provider-kion/internal/kgen/config/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// standardOps/standardSvcs are a small fixture reused across the render/check
// tests: a full-CRUD resource (label), a create-only resource (noread) that
// triggers the "no read" warning, a heuristic resource (widget), and a data
// source (thing).
func standardOps() map[string]config.Op {
	return map[string]config.Op{
		"PostLabel":     {Method: "POST", Path: "/v3/label"},
		"GetLabel":      {Method: "GET", Path: "/v3/label/{id}"},
		"PatchLabel":    {Method: "PATCH", Path: "/v3/label/{id}"},
		"DeleteLabel":   {Method: "DELETE", Path: "/v3/label/{id}"},
		"PostNoread":    {Method: "POST", Path: "/v3/noread"},
		"PostWidget":    {Method: "POST", Path: "/v3/widget"},
		"GetWidget":     {Method: "GET", Path: "/v3/widget/{id}"},
		"GetThingIndex": {Method: "GET", Path: "/v3/thing"},
	}
}

func standardSvcs() []config.ServiceOps {
	return []config.ServiceOps{
		{Name: "label", Create: []string{"PostLabel"}, Read: []string{"GetLabel"}, Update: []string{"PatchLabel"}, Delete: []string{"DeleteLabel"}},
		{Name: "noread", Create: []string{"PostNoread"}},
		{Name: "widget"}, // stub -> heuristic
		{Name: "thing", DataSourceRead: []string{"GetThingIndex"}},
	}
}

// newSource returns a mock wired with the standard fixture for the default spec
// and service-root paths (so orDefault's default branch is exercised).
func newSource(t *testing.T) *mocks.MockSource {
	t.Helper()
	m := mocks.NewMockSource(t)
	m.EXPECT().Operations("spec/openapi3.json").Return(standardOps(), nil)
	m.EXPECT().ServiceOps("internal/service").Return(standardSvcs(), nil)
	return m
}

// failWriter fails every write, to exercise the io.Writer error paths.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestGen(t *testing.T) {
	m := newSource(t)

	var buf bytes.Buffer
	// Empty Options exercises orDefault's default branch; the default overrides
	// path does not exist under the package dir, so withOverrides is a no-op.
	require.NoError(t, config.Gen(m, config.Options{}, &buf))
	out := buf.String()

	assert.Contains(t, out, "provider:\n  name: kion")
	assert.Contains(t, out, "resources:")
	assert.Contains(t, out, "  label:")
	assert.Contains(t, out, "path: /v3/label")
	assert.Contains(t, out, "method: DELETE")
	// heuristic resources are flagged for review.
	assert.Contains(t, out, "# heuristic")
	// create-only resource (noread) is listed as incomplete, missing its read.
	assert.Contains(t, out, "# noread: INCOMPLETE, missing read")
	// data source section + its narrower ignores.
	assert.Contains(t, out, "data_sources:")
	assert.Contains(t, out, "ignores: [status, record_id]")
	assert.Contains(t, out, "# summary:")
}

// TestGenIncompleteMissingCreate covers render's create-missing branch: a
// read-only resource (no create op) is listed as incomplete.
func TestGenIncompleteMissingCreate(t *testing.T) {
	m := mocks.NewMockSource(t)
	ops := map[string]config.Op{"GetReadonly": {Method: "GET", Path: "/v3/readonly/{id}"}}
	svcs := []config.ServiceOps{{Name: "readonly", Read: []string{"GetReadonly"}}}
	m.EXPECT().Operations("spec/openapi3.json").Return(ops, nil)
	m.EXPECT().ServiceOps("internal/service").Return(svcs, nil)

	var buf bytes.Buffer
	require.NoError(t, config.Gen(m, config.Options{}, &buf))
	assert.Contains(t, buf.String(), "# readonly: INCOMPLETE, missing create")
}

func TestGenWithOverrides(t *testing.T) {
	m := newSource(t)

	ovPath := filepath.Join(t.TempDir(), "overrides.yaml")
	require.NoError(t, os.WriteFile(ovPath, []byte(`
resources:
  brandnew:
    create:
      path: /v3/brand-new
      method: POST
    read:
      path: /v3/brand-new/{id}
      method: GET
    update:
      path: /v3/brand-new/{id}
      method: PATCH
    delete:
      path: /v3/brand-new/{id}
      method: DELETE
data_sources:
  extrads:
    read:
      path: /v3/extra-ds
      method: GET
`), 0o600))

	var buf bytes.Buffer
	require.NoError(t, config.Gen(m, config.Options{Overrides: ovPath}, &buf))
	out := buf.String()

	assert.Contains(t, out, "  brandnew:")
	assert.Contains(t, out, "path: /v3/brand-new")
	assert.Contains(t, out, "  extrads:")
	assert.Contains(t, out, "path: /v3/extra-ds")
}

// TestGenOverrideExtraIgnores verifies an override's `ignores` list is appended
// to the resource's baked-in schema ignores. Used to drop a polymorphic OAS
// field (e.g. default_value) that would otherwise make tfplugingen skip the
// whole resource.
func TestGenOverrideExtraIgnores(t *testing.T) {
	m := newSource(t)

	ovPath := filepath.Join(t.TempDir(), "overrides.yaml")
	require.NoError(t, os.WriteFile(ovPath, []byte(`
resources:
  brandnew:
    create:
      path: /v3/brand-new
      method: POST
    read:
      path: /v3/brand-new/{id}
      method: GET
    ignores: [default_value]
`), 0o600))

	var buf bytes.Buffer
	require.NoError(t, config.Gen(m, config.Options{Overrides: ovPath}, &buf))
	out := buf.String()

	assert.Contains(t, out, "ignores: [status, record_id, data, default_value]")
}

// TestGenDataSourceOverrideIgnores verifies a data-source override's `ignores`
// list is appended to its baked-in schema ignores (used to drop a polymorphic
// field so the data source generates).
func TestGenDataSourceOverrideIgnores(t *testing.T) {
	m := newSource(t)

	ovPath := filepath.Join(t.TempDir(), "overrides.yaml")
	require.NoError(t, os.WriteFile(ovPath, []byte(`
data_sources:
  thing:
    read:
      path: /v3/thing
      method: GET
    ignores: [default_value]
`), 0o600))

	var buf bytes.Buffer
	require.NoError(t, config.Gen(m, config.Options{Overrides: ovPath}, &buf))

	assert.Contains(t, buf.String(), "ignores: [status, record_id, default_value]")
}

// A data source read's `match` carries through to the rendered config, where
// the import manifest reads it as the parent filter.
func TestGenDataSourceReadMatch(t *testing.T) {
	m := newSource(t)

	ovPath := filepath.Join(t.TempDir(), "overrides.yaml")
	require.NoError(t, os.WriteFile(ovPath, []byte(`
data_sources:
  thing:
    read:
      path: /v3/owner/{id}/thing
      method: GET
      match: { owner_type_id: "3", region: east }
`), 0o600))

	var buf bytes.Buffer
	require.NoError(t, config.Gen(m, config.Options{Overrides: ovPath}, &buf))
	assert.Contains(t, buf.String(), `  thing:
    read:
      path: /v3/owner/{id}/thing
      method: GET
      match:
        owner_type_id: "3"
        region: "east"
`)
}

func TestGenRejectsInvalidReadMatch(t *testing.T) {
	cases := map[string]string{
		"empty field": `
data_sources:
  thing:
    read: { path: "/v3/owner/{id}/thing", method: GET, match: { "": "3" } }
`,
		"empty value": `
data_sources:
  thing:
    read: { path: "/v3/owner/{id}/thing", method: GET, match: { owner_type_id: "" } }
`,
		"empty map": `
data_sources:
  thing:
    read: { path: "/v3/owner/{id}/thing", method: GET, match: {} }
`,
		"no parent in path": `
data_sources:
  thing:
    read: { path: /v3/thing, method: GET, match: { owner_type_id: "3" } }
`,
		"on a resource": `
resources:
  label:
    read: { path: "/v3/owner/{id}/label", method: GET, match: { owner_type_id: "3" } }
`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			m := newSource(t)
			ovPath := filepath.Join(t.TempDir(), "overrides.yaml")
			require.NoError(t, os.WriteFile(ovPath, []byte(body), 0o600))
			var buf bytes.Buffer
			err := config.Gen(m, config.Options{Overrides: ovPath}, &buf)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "match")
		})
	}
}

func TestGenMalformedOverrides(t *testing.T) {
	m := newSource(t)

	ovPath := filepath.Join(t.TempDir(), "bad.yaml")
	require.NoError(t, os.WriteFile(ovPath, []byte("resources: [not a map"), 0o600))

	var buf bytes.Buffer
	err := config.Gen(m, config.Options{Overrides: ovPath}, &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing overrides")
}

func TestGenWriteError(t *testing.T) {
	m := newSource(t)
	err := config.Gen(m, config.Options{}, failWriter{})
	require.Error(t, err)
}

func TestGenDeriveError(t *testing.T) {
	m := mocks.NewMockSource(t)
	m.EXPECT().Operations("spec.json").Return(nil, errors.New("boom"))
	var buf bytes.Buffer
	err := config.Gen(m, config.Options{Spec: "spec.json", ServiceRoot: "svc"}, &buf)
	require.Error(t, err)
}

func TestGenOverridesReadError(t *testing.T) {
	m := newSource(t)
	// A directory is not IsNotExist, so withOverrides surfaces a read error.
	var buf bytes.Buffer
	err := config.Gen(m, config.Options{Overrides: t.TempDir()}, &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading overrides")
}

// TestDerivePickVerbAndHeuristic exercises pickVerb's candidate-missing and
// method-mismatch branches, and heuristic's patch/delete fallbacks.
func TestDerivePickVerbAndHeuristic(t *testing.T) {
	m := mocks.NewMockSource(t)
	ops := map[string]config.Op{
		"PostX":   {Method: "POST", Path: "/v3/x"},
		"GetX":    {Method: "GET", Path: "/v3/x/{id}"},
		"PatchX":  {Method: "PATCH", Path: "/v3/x/{id}"},
		"DeleteX": {Method: "DELETE", Path: "/v3/x/{id}"},
	}
	svcs := []config.ServiceOps{
		// Create candidates: "NoSuchOp" is absent (pickVerb skips it), "GetX" is
		// present but is a GET (method mismatch for a create), so the code-derived
		// create resolves to nil and the resource falls back to the heuristic,
		// which finds Post/Get/Patch/DeleteX by name.
		{Name: "x", Create: []string{"NoSuchOp", "GetX"}},
	}
	m.EXPECT().Operations("s").Return(ops, nil)
	m.EXPECT().ServiceOps("r").Return(svcs, nil)

	got, err := config.Derive(m, config.Options{Spec: "s", ServiceRoot: "r"})
	require.NoError(t, err)
	require.Len(t, got, 1)
	x := got[0]
	assert.True(t, x.ResourceHeuristic)
	require.NotNil(t, x.Resource)
	assert.Equal(t, "/v3/x", x.Resource.Create.Path)
	assert.Equal(t, "/v3/x/{id}", x.Resource.Update.Path)
	assert.Equal(t, "/v3/x/{id}", x.Resource.Delete.Path)
}

func TestDeriveErrors(t *testing.T) {
	t.Run("operations error", func(t *testing.T) {
		m := mocks.NewMockSource(t)
		m.EXPECT().Operations("spec.json").Return(nil, errors.New("boom"))
		_, err := config.Derive(m, config.Options{Spec: "spec.json", ServiceRoot: "svc"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading spec operations")
	})

	t.Run("service ops error", func(t *testing.T) {
		m := mocks.NewMockSource(t)
		m.EXPECT().Operations("spec.json").Return(map[string]config.Op{}, nil)
		m.EXPECT().ServiceOps("svc").Return(nil, errors.New("boom"))
		_, err := config.Derive(m, config.Options{Spec: "spec.json", ServiceRoot: "svc"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading service ops")
	})
}

// checkSource wires a fresh mock with the standard fixture bound to explicit
// spec/service-root paths, for the Check drift scenarios.
func checkSource(t *testing.T) *mocks.MockSource {
	t.Helper()
	m := mocks.NewMockSource(t)
	m.EXPECT().Operations("spec.json").Return(standardOps(), nil)
	m.EXPECT().ServiceOps("svc").Return(standardSvcs(), nil)
	return m
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "generator_config.yaml")
	require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	return p
}

func runCheck(t *testing.T, cfgPath string) (int, string) {
	t.Helper()
	m := checkSource(t)
	var buf bytes.Buffer
	n, err := config.Check(m, config.Options{Spec: "spec.json", ServiceRoot: "svc", ConfigPath: cfgPath}, &buf)
	require.NoError(t, err)
	return n, buf.String()
}

func TestCheck(t *testing.T) {
	t.Run("in sync", func(t *testing.T) {
		cfg := writeConfig(t, `
resources:
  label:
    create:
      path: /v3/label
      method: POST
  noread:
    create:
      path: /v3/noread
      method: POST
`)
		n, out := runCheck(t, cfg)
		assert.Equal(t, 0, n)
		assert.Contains(t, out, "config in sync")
	})

	t.Run("missing implemented resource", func(t *testing.T) {
		cfg := writeConfig(t, "resources: {}\n")
		n, out := runCheck(t, cfg)
		assert.Positive(t, n)
		assert.Contains(t, out, "missing: resource \"label\"")
	})

	t.Run("stale create path", func(t *testing.T) {
		cfg := writeConfig(t, `
resources:
  label:
    create:
      path: /v3/OLD-label
      method: POST
  noread:
    create:
      path: /v3/noread
      method: POST
`)
		n, out := runCheck(t, cfg)
		assert.Positive(t, n)
		assert.Contains(t, out, "stale: resource \"label\"")
	})

	t.Run("extra config resource", func(t *testing.T) {
		cfg := writeConfig(t, `
resources:
  label:
    create:
      path: /v3/label
      method: POST
  noread:
    create:
      path: /v3/noread
      method: POST
  ghost:
    create:
      path: /v3/ghost
      method: POST
`)
		n, out := runCheck(t, cfg)
		assert.Positive(t, n)
		assert.Contains(t, out, "extra: config resource \"ghost\"")
	})
}

func TestCheckReadErrors(t *testing.T) {
	t.Run("derive error", func(t *testing.T) {
		m := mocks.NewMockSource(t)
		m.EXPECT().Operations("spec.json").Return(nil, errors.New("boom"))
		var buf bytes.Buffer
		_, err := config.Check(m, config.Options{Spec: "spec.json", ServiceRoot: "svc", ConfigPath: "unused"}, &buf)
		require.Error(t, err)
	})

	t.Run("overrides read error", func(t *testing.T) {
		m := checkSource(t)
		var buf bytes.Buffer
		_, err := config.Check(m, config.Options{Spec: "spec.json", ServiceRoot: "svc", Overrides: t.TempDir(), ConfigPath: "unused"}, &buf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading overrides")
	})

	t.Run("missing config file", func(t *testing.T) {
		m := checkSource(t)
		var buf bytes.Buffer
		_, err := config.Check(m, config.Options{Spec: "spec.json", ServiceRoot: "svc", ConfigPath: filepath.Join(t.TempDir(), "nope.yml")}, &buf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading")
	})

	t.Run("malformed config file", func(t *testing.T) {
		cfg := writeConfig(t, "resources: [not a map")
		m := checkSource(t)
		var buf bytes.Buffer
		_, err := config.Check(m, config.Options{Spec: "spec.json", ServiceRoot: "svc", ConfigPath: cfg}, &buf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing")
	})

	t.Run("report write error", func(t *testing.T) {
		// An empty config makes both implemented resources "missing" -> two emit
		// calls. The first sets the write error; the second exercises the guard
		// that short-circuits once a write has failed.
		cfg := writeConfig(t, "resources: {}\n")
		m := checkSource(t)
		_, err := config.Check(m, config.Options{Spec: "spec.json", ServiceRoot: "svc", ConfigPath: cfg}, failWriter{})
		require.Error(t, err)
	})
}
