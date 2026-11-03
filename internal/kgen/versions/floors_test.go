package versions

import (
	"os"
	"path/filepath"
	"testing"

	"terraform-provider-kion/internal/kgen/kfs/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// generateWithOverrides runs the pipeline over one resource (and optionally a
// data source) whose defining op exists in the tracked versions named by
// presentIn, with attrOverrides as attr_version_overrides.yaml. It returns the
// files written.
func generateWithOverrides(t *testing.T, presentIn []string, withDataSource bool, attrOverrides string) (map[string]string, error) {
	t.Helper()
	cfg := "resources:\n  thing:\n    create:\n      path: /v3/thing\n      method: POST\n"
	if withDataSource {
		cfg += "data_sources:\n  thing:\n    read:\n      path: /v3/thing\n      method: POST\n"
	}
	m := mocks.NewMockFS(t)
	m.EXPECT().ReadFile("cfg/generator_config.yaml").Return([]byte(cfg), nil)
	m.EXPECT().ReadFile("cfg/config_overrides.yaml").Return(nil, os.ErrNotExist)
	m.EXPECT().ReadFile("cfg/attr_version_overrides.yaml").Return([]byte(attrOverrides), nil)

	in := map[string]bool{}
	for _, d := range presentIn {
		in[d] = true
	}
	thing := op{method: "POST", path: "/v3/thing"}
	for _, v := range trackedVersions {
		path := filepath.Join("sdk", "generated", v.dir, "oas_client_gen.go")
		if in[v.dir] {
			m.EXPECT().ReadFile(path).Return(clientSrc(thing), nil)
		} else {
			m.EXPECT().ReadFile(path).Return(clientSrc(), nil)
		}
	}
	m.EXPECT().ReadDir(mock.Anything).Return(nil, os.ErrNotExist).Maybe()
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(nil).Maybe()
	writes := map[string]string{}
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		Run(func(name string, data []byte, _ os.FileMode) { writes[name] = string(data) }).
		Return(nil).Maybe()

	// The stub has no attributes of interest: keep derivation to the op gate.
	g := &generator{fs: m, src: stubSource{newFieldFrom: "v9_99"}}
	_, err := g.generate(Options{
		SDKDir: "sdk", ServiceRoot: "svc",
		ConfigPath: "cfg/generator_config.yaml", Overrides: "cfg/config_overrides.yaml",
		AttrOverrides: "cfg/attr_version_overrides.yaml",
	})
	return writes, err
}

func TestResourceFloorRaisesDerivedMin(t *testing.T) {
	writes, err := generateWithOverrides(t, []string{"v3_17"}, true, `resource_floors: { thing: "3.17.3" }`)
	require.NoError(t, err)
	out := writes[filepath.Join("svc", "thing", "thing_version_gen.go")]
	assert.Contains(t, out, `minKionVersion = conns.MustParseKionVersion("3.17.3")`)
	assert.NotContains(t, out, `"3.17.0"`)
}

func TestResourceFloorBelowDerivedIsAnError(t *testing.T) {
	_, err := generateWithOverrides(t, []string{"v3_17"}, false, `resource_floors: { thing: "3.16.0" }`)
	require.ErrorContains(t, err, "below the derived minimum")
	require.ErrorContains(t, err, "thing")
	require.ErrorContains(t, err, "3.16.0")
	require.ErrorContains(t, err, "3.17.0")
}

func TestResourceFloorOnUngatedEntryStillGates(t *testing.T) {
	// Op present in every tracked version: no gate derived, yet a floor needs one.
	writes, err := generateWithOverrides(t, []string{"v3_14", "v3_15", "v3_16", "v3_17"}, false, `resource_floors: { thing: "3.17.3" }`)
	require.NoError(t, err)
	out := writes[filepath.Join("svc", "thing", "thing_version_gen.go")]
	require.NotEmpty(t, out, "a floor on an ungated entry must produce a gate")
	assert.Contains(t, out, `minKionVersion = conns.MustParseKionVersion("3.17.3")`)
}

func TestResourceFloorOnUnknownNameIsAnError(t *testing.T) {
	_, err := generateWithOverrides(t, []string{"v3_17"}, false, `resource_floors: { nothing: "3.17.3" }`)
	require.ErrorContains(t, err, "neither a resource nor a data source")
	require.ErrorContains(t, err, "nothing")
}

func TestResourceFloorMalformedIsAnError(t *testing.T) {
	_, err := generateWithOverrides(t, []string{"v3_17"}, false, `resource_floors: { thing: "3.17" }`)
	require.ErrorContains(t, err, "not a 3.N.P version")
}

func TestRaiseMin(t *testing.T) {
	got, err := RaiseMin("x", "3.17.0", "", "")
	require.NoError(t, err)
	assert.Equal(t, "3.17.0", got)

	got, err = RaiseMin("x", "3.17.0", "", "3.17.0")
	require.NoError(t, err)
	assert.Equal(t, "3.17.0", got, "equal to the derived minimum is allowed")

	_, err = RaiseMin("x", "3.14.0", "3.15.0", "3.15.2")
	require.ErrorContains(t, err, "not below the derived maximum")
}

func TestLoadFloors(t *testing.T) {
	read := func(b string, err error) func(string) ([]byte, error) {
		return func(string) ([]byte, error) { return []byte(b), err }
	}
	res, ds := []string{"thing"}, []string{"other"}

	t.Run("missing file is no floors", func(t *testing.T) {
		f, err := LoadFloors(read("", os.ErrNotExist), "p", res, ds)
		require.NoError(t, err)
		assert.Empty(t, f)
	})
	t.Run("other read error is returned", func(t *testing.T) {
		_, err := LoadFloors(read("", os.ErrPermission), "p", res, ds)
		require.ErrorIs(t, err, os.ErrPermission)
	})
	t.Run("unknown top-level key fails", func(t *testing.T) {
		_, err := LoadFloors(read(`resource_floor: { thing: "3.17.3" }`, nil), "p", res, ds)
		require.ErrorContains(t, err, "resource_floor")
	})
	t.Run("data source name is a valid target", func(t *testing.T) {
		f, err := LoadFloors(read(`resource_floors: { other: "3.17.3" }`, nil), "p", res, ds)
		require.NoError(t, err)
		assert.Equal(t, "3.17.3", f["other"])
	})
}

func TestApplyFloor(t *testing.T) {
	got, emit, err := ApplyFloor("x", "resources", "", FloorNoOp, "", "", "")
	require.NoError(t, err)
	assert.False(t, emit)
	assert.Equal(t, "", got)

	_, _, err = ApplyFloor("x", "resources", "3.17.3", FloorNoOp, "", "", "")
	require.ErrorContains(t, err, "no create or read operation")
	_, _, err = ApplyFloor("x", "resources", "3.17.3", FloorUnresolved, "POST /v3/x", "", "")
	require.ErrorContains(t, err, "no tracked SDK version")

	got, emit, err = ApplyFloor("x", "resources", "3.17.3", FloorResolved, "", "3.17.0", "")
	require.NoError(t, err)
	assert.True(t, emit)
	assert.Equal(t, "3.17.3", got)
}

func TestGenerate_unknownAttrOverridesKeyFails(t *testing.T) {
	_, err := generateWithOverrides(t, []string{"v3_17"}, false, `resource_floor: { thing: "3.17.3" }`)
	require.ErrorContains(t, err, "resource_floor")
}
