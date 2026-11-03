package importmodules_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/importmodules"
)

// TestParseManifest_roundTrip covers the JSON shape module.Manifest actually
// writes (module.MarshalManifest), independent of the compiled schemas: this
// package must decode that exact shape without importing internal/provider.
func TestParseManifest_roundTrip(t *testing.T) {
	t.Parallel()
	data := []byte(`{
  "version": 1,
  "modules": [
    {
      "tf_type": "kion_ou",
      "path": "terraform-kion-ou",
      "variables": [
        {"attr": "description", "var": "description", "block": false},
        {"attr": "name", "var": "name", "block": false}
      ]
    }
  ]
}
`)
	m, err := importmodules.ParseManifest(data)
	require.NoError(t, err)
	require.Len(t, m.Modules, 1)
	assert.Equal(t, "kion_ou", m.Modules[0].TFType)
	assert.Equal(t, "terraform-kion-ou", m.Modules[0].Path)
	require.Len(t, m.Modules[0].Variables, 2)
	assert.Equal(t, importmodules.Variable{Attr: "name", Var: "name"}, m.Modules[0].Variables[1])

	byType := m.ByType()
	assert.Contains(t, byType, "kion_ou")
}

// TestLoadManifest_missingFile: a clear error, not a nil-pointer panic later,
// when --manifest points at a path that does not exist.
func TestLoadManifest_missingFile(t *testing.T) {
	t.Parallel()
	_, err := importmodules.LoadManifest(filepath.Join(t.TempDir(), "does-not-exist.json"))
	assert.Error(t, err)
}
