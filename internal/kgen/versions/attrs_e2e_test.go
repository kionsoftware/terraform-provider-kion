package versions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"terraform-provider-kion/internal/kgen/crud"
	"terraform-provider-kion/internal/kgen/kfs/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// stubSource serves one create op whose body gained a field in v3_16, so the
// attribute path can be exercised without real SDK files on disk.
type stubSource struct {
	newFieldFrom string
	// newFieldUntil, when set, is the LAST version carrying the field. Later
	// versions drop it, which is what a backport ahead of the newest line looks
	// like from the SDK's side.
	newFieldUntil string
}

func (s stubSource) ClientMethods(file string) ([]crud.ClientMethod, error) {
	return []crud.ClientMethod{{
		Name: "PostThing", HTTPMethod: "POST", Path: "/v3/thing", BodyType: "ThingCreate",
	}}, nil
}

func (s stubSource) Structs(file string) (map[string]crud.Struct, error) {
	fields := []crud.Field{{GoName: "Name", JSONName: "name", Type: "string"}}
	// The named version AND LATER carry the newer field. Compared numerically,
	// not by substring: a substring match means "only this version", which is
	// the same thing only while the named version happens to be the newest --
	// it stopped being true the moment v3_17 was tracked.
	present := versionAtLeast(file, s.newFieldFrom)
	if present && s.newFieldUntil != "" && !versionAtLeast(s.newFieldUntil, file) {
		present = false // dropped by a release newer than the window
	}
	if present {
		fields = append(fields, crud.Field{GoName: "Shiny", JSONName: "shiny", Type: "OptString"})
	}
	return map[string]crud.Struct{"ThingCreate": {Name: "ThingCreate", Fields: fields}}, nil
}

// versionAtLeast reports whether the v3_NN in file is >= the v3_NN in want.
func versionAtLeast(file, want string) bool {
	minorOf := func(s string) int {
		i := strings.Index(s, "v3_")
		if i < 0 {
			return -1
		}
		n := 0
		for _, c := range s[i+3:] {
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + int(c-'0')
		}
		return n
	}
	f, w := minorOf(file), minorOf(want)
	return f >= 0 && w >= 0 && f >= w
}

func (s stubSource) MarkerImpls(string) (map[string][]string, error) { return nil, nil }

func (s stubSource) ModelFields(string, string) ([]crud.ModelField, error) {
	return []crud.ModelField{
		{GoName: "Name", TFSDK: "name", Type: "types.String"},
		{GoName: "Shiny", TFSDK: "shiny", Type: "types.String"},
	}, nil
}

// A resource present in every version still needs a file when one of its
// attributes is newer, the case the whole change exists for. Without the
// injected source this path was never reached end to end: the real filesystem
// had no SDK at the test paths, so deriveAttrMins returned empty and the test
// passed while proving nothing.
func TestGenerate_attributeOnlyResourceGetsGate(t *testing.T) {
	cfg := `resources:
  thing:
    create:
      path: /v3/thing
      method: POST
`
	m := mocks.NewMockFS(t)
	m.EXPECT().ReadFile("cfg/generator_config.yaml").Return([]byte(cfg), nil)
	m.EXPECT().ReadFile("cfg/config_overrides.yaml").Return(nil, os.ErrNotExist)
	m.EXPECT().ReadFile("cfg/attr_version_overrides.yaml").Return(nil, os.ErrNotExist)

	// The op exists in every tracked version, so the resource itself is ungated.
	full := op{method: "POST", path: "/v3/thing"}
	for _, v := range trackedVersions {
		path := filepath.Join("sdk", "generated", v.dir, "oas_client_gen.go")
		m.EXPECT().ReadFile(path).Return(clientSrc(full), nil)
	}
	m.EXPECT().ReadDir(mock.Anything).Return(nil, os.ErrNotExist)
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(nil)
	writes := map[string]string{}
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		Run(func(name string, data []byte, _ os.FileMode) { writes[name] = string(data) }).
		Return(nil)

	g := &generator{fs: m, src: stubSource{newFieldFrom: "v3_16"}}
	n, err := g.generate(Options{
		SDKDir: "sdk", ServiceRoot: "svc",
		ConfigPath: "cfg/generator_config.yaml", Overrides: "cfg/config_overrides.yaml",
		AttrOverrides: "cfg/attr_version_overrides.yaml",
	})
	require.NoError(t, err)
	require.Equal(t, 1, n, "an attribute-only resource must still get a file")

	path := filepath.Join("svc", "thing", "thing_version_gen.go")
	require.Contains(t, writes, path)
	out := writes[path]

	// Resource gate unbounded (a no-op), attribute gate doing the work.
	assert.Contains(t, out, "minKionVersion = conns.KionVersion{}")
	assert.Contains(t, out, `"shiny": {{Min: conns.MustParseKionVersion("3.16.0")}}`)
	assert.NotContains(t, out, `"name":`, "a field present since the oldest version needs no gate")
	assert.Contains(t, out, "framework.RequireAttrKionVersions(")
	assert.Contains(t, out, "attrKionVersions")
}

// A field present in the middle of the tracked range and dropped by the newest
// release: it needs both a min and a max, and it has to survive being absent
// from the newest version's field list.
func TestGenerate_attributeDroppedByNewestVersionGetsAnUpperBound(t *testing.T) {
	cfg := `resources:
  thing:
    create:
      path: /v3/thing
      method: POST
`
	m := mocks.NewMockFS(t)
	m.EXPECT().ReadFile("cfg/generator_config.yaml").Return([]byte(cfg), nil)
	m.EXPECT().ReadFile("cfg/config_overrides.yaml").Return(nil, os.ErrNotExist)
	m.EXPECT().ReadFile("cfg/attr_version_overrides.yaml").Return(nil, os.ErrNotExist)

	full := op{method: "POST", path: "/v3/thing"}
	for _, v := range trackedVersions {
		path := filepath.Join("sdk", "generated", v.dir, "oas_client_gen.go")
		m.EXPECT().ReadFile(path).Return(clientSrc(full), nil)
	}
	m.EXPECT().ReadDir(mock.Anything).Return(nil, os.ErrNotExist)
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(nil)
	writes := map[string]string{}
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		Run(func(name string, data []byte, _ os.FileMode) { writes[name] = string(data) }).
		Return(nil)

	// Present from v3_15 through v3_16, gone in v3_17.
	g := &generator{fs: m, src: stubSource{newFieldFrom: "v3_15", newFieldUntil: "v3_16"}}
	n, err := g.generate(Options{
		SDKDir: "sdk", ServiceRoot: "svc",
		ConfigPath: "cfg/generator_config.yaml", Overrides: "cfg/config_overrides.yaml",
		AttrOverrides: "cfg/attr_version_overrides.yaml",
	})
	require.NoError(t, err)
	require.Equal(t, 1, n)

	out := writes[filepath.Join("svc", "thing", "thing_version_gen.go")]
	assert.Contains(t, out, `"shiny":`, "a field the newest version lacks must still be gated")
	assert.Contains(t, out, `Min: conns.MustParseKionVersion("3.15.0")`)
	assert.Contains(t, out, `Before: conns.MustParseKionVersion("3.17.0")`,
		"carried through 3.16, so the exclusive bound is the next tracked line")
}
