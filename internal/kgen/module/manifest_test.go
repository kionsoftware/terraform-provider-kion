package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManifestIsCurrent fails when modules/module_manifest.json does not match
// what BuildManifest would produce now -- i.e. someone changed a resource's
// schema (adding, renaming or removing a settable attribute) without running
// `make modules`.
//
// Unlike Generate, BuildManifest needs no build tag: it reads the compiled
// resource schemas the same way module_test.go's helpers do, so this test runs
// in the default `go test ./...` build.
func TestManifestIsCurrent(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)

	onDisk, err := os.ReadFile(filepath.Join(root, "modules", ManifestFileName))
	require.NoError(t, err, "run: make modules")

	fresh, err := MarshalManifest(BuildManifest())
	require.NoError(t, err)

	assert.Equal(t, string(onDisk), string(fresh),
		"modules/module_manifest.json is stale -- run: make modules")
}

// TestBuildManifest_knownCollisions locks in the attribute name that collides
// with a `module` block meta-argument, so a rewriter reading this manifest never
// emits `version = ...` as a plain attribute inside a module block, which
// Terraform would misread as the module's own version constraint.
func TestBuildManifest_knownCollisions(t *testing.T) {
	t.Parallel()
	m := BuildManifest()

	cases := []struct {
		tfType, attr, wantVar string
	}{
		{"kion_compliance_program", "version", "compliance_program_version"},
	}
	for _, c := range cases {
		mod := findModule(t, m, c.tfType)
		v := findVariable(t, mod, c.attr)
		assert.Equal(t, c.wantVar, v.Var, "%s.%s", c.tfType, c.attr)
	}
}

// TestBuildManifest_dropsLastUpdated locks in that provider bookkeeping never
// leaks into the manifest as something a caller could try to set.
func TestBuildManifest_dropsLastUpdated(t *testing.T) {
	t.Parallel()
	m := BuildManifest()
	for _, mod := range m.Modules {
		for _, v := range mod.Variables {
			assert.NotEqual(t, "last_updated", v.Attr, "module %s", mod.TFType)
		}
	}
}

func findModule(t *testing.T, m *Manifest, tfType string) Entry {
	t.Helper()
	for _, mod := range m.Modules {
		if mod.TFType == tfType {
			return mod
		}
	}
	t.Fatalf("manifest has no module for %s", tfType)
	return Entry{}
}

func findVariable(t *testing.T, mod Entry, attr string) VariableManifest {
	t.Helper()
	for _, v := range mod.Variables {
		if v.Attr == attr {
			return v
		}
	}
	t.Fatalf("module %s has no variable for attribute %s", mod.TFType, attr)
	return VariableManifest{}
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
