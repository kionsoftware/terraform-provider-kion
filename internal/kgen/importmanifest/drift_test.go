package importmanifest_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/importmanifest"
	"terraform-provider-kion/internal/kgen/kfs"
)

// TestManifestIsCurrent fails when codegen/import_manifest.json does not match
// what Generate would produce now -- i.e. someone changed an archetype or a read
// path without running `make import-manifest`.
func TestManifestIsCurrent(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)

	onDisk, err := os.ReadFile(filepath.Join(root, importmanifest.OutputPath))
	require.NoError(t, err, "run: make import-manifest")

	tmp := t.TempDir()
	mirrorCodegen(t, root, tmp)
	_, err = importmanifest.Generate(kfs.OS{}, tmp)
	require.NoError(t, err)

	fresh, err := os.ReadFile(filepath.Join(tmp, importmanifest.OutputPath))
	require.NoError(t, err)

	assert.Equal(t, string(onDisk), string(fresh),
		"import_manifest.json is stale -- run: make import-manifest")
}

// TestEveryProviderResourceHasARow is the scope contract: a resource added to
// the provider must appear in the manifest, readable or explicitly not.
func TestEveryProviderResourceHasARow(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)

	raw, err := os.ReadFile(filepath.Join(root, "codegen", "schema_snapshots", "new.json"))
	require.NoError(t, err)
	var snap struct {
		ProviderSchemas map[string]struct {
			ResourceSchemas map[string]json.RawMessage `json:"resource_schemas"`
		} `json:"provider_schemas"`
	}
	require.NoError(t, json.Unmarshal(raw, &snap))

	raw, err = os.ReadFile(filepath.Join(root, importmanifest.OutputPath))
	require.NoError(t, err)
	var m importmanifest.Manifest
	require.NoError(t, json.Unmarshal(raw, &m))

	got := map[string]bool{}
	for _, r := range m.Resources {
		got[r.TFType] = true
	}
	for _, provider := range snap.ProviderSchemas {
		for tfType := range provider.ResourceSchemas {
			assert.True(t, got[tfType], "%s missing from the manifest", tfType)
		}
	}
}

// TestUnreadableResourcesStateWhy: a gap must be explained, never blank.
func TestUnreadableResourcesStateWhy(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(projectRoot(t), importmanifest.OutputPath))
	require.NoError(t, err)
	var m importmanifest.Manifest
	require.NoError(t, json.Unmarshal(raw, &m))

	for _, r := range m.Resources {
		if !r.Readable {
			assert.NotEmpty(t, r.Reason, r.TFType)
		}
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for range 6 {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("could not locate project root")
	return ""
}

func mirrorCodegen(t *testing.T, root, dst string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dst, "codegen", "schema_snapshots"), 0o755))
	for _, rel := range []string{
		"codegen/generator_config.yaml",
		"codegen/crud_archetypes.yaml",
		"codegen/private_endpoints.yaml",
		"codegen/schema_snapshots/new.json",
	} {
		data, err := os.ReadFile(filepath.Join(root, rel))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dst, rel), data, 0o644))
	}
}
