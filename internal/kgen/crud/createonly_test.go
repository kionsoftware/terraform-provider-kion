package crud

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// createOnlyThing is the string-id fixture with secret in the model; the read
// payload (Thing) does not carry it, the create payload (ThingCreated) does.
func createOnlyThing(createOnly ...string) fixtureSpec {
	spec := thingSpec()
	spec.Model = filepath.Join("testdata", "createonly", "thing_schema_gen.go")
	spec.Arch = &archetype{Kind: entityKind, CreateOnly: createOnly}
	return spec
}

// createOnlyWidget is the int-id fixture: create answers a *WidgetCreated
// envelope, and the read payload (Widget) carries secret too.
func createOnlyWidget(arch archetype) fixtureSpec {
	arch.Kind = entityKind
	return fixtureSpec{
		Name: "widget",
		Dir:  filepath.Join("testdata", "createonlyint"),
		Ops: resOps{
			Create: &opRef{Method: "POST", Path: "/v3/widget"},
			Read:   &opRef{Method: "GET", Path: "/v3/widget/{id}"},
			Update: &opRef{Method: "PUT", Path: "/v3/widget/{id}"},
			Delete: &opRef{Method: "DELETE", Path: "/v3/widget/{id}"},
		},
		Arch: &arch,
	}
}

func generateParsed(t *testing.T, spec fixtureSpec) map[string]string {
	t.Helper()
	files, err := generateFixture(t, spec)
	require.NoError(t, err)
	for name, src := range files {
		_, err := parser.ParseFile(token.NewFileSet(), name, src, parser.ParseComments)
		require.NoError(t, err, "%s does not parse:\n%s", name, src)
	}
	return files
}

// funcBody returns the source of the named top-level function or method.
func funcBody(t *testing.T, src, name string) string {
	t.Helper()
	m := regexp.MustCompile(`(?s)\nfunc (\([^)]*\) )?` + regexp.QuoteMeta(name) + `\(.*?\n}\n`).FindString(src)
	require.NotEmpty(t, m, "func %s not found", name)
	return m
}

func TestEntityCreateOnlyAttribute(t *testing.T) {
	src := generateParsed(t, createOnlyThing("secret"))["thing.go"]

	// captured from the POST response, before the read-back can null it
	require.Regexp(t, `configuredSecret := flex\.OptStringToFrameworkNullIfEmpty\(created\.Data\.Value\.Secret\)[\s\S]*readOut, readErr :=`, src)
	require.Contains(t, src, "plan.Secret = configuredSecret")
	// an empty string from a public client stores null, not ""
	require.Contains(t, src, "flex.OptStringToFrameworkNullIfEmpty(")

	create := funcBody(t, src, "Create")
	require.Regexp(t, `flattenThing\(readOut, &plan\)[\s\S]*plan\.Secret = configuredSecret`, create)

	// Read and flatten never assign it, so state keeps the captured value.
	require.NotContains(t, funcBody(t, src, "flattenThing"), "Secret")
	require.NotContains(t, funcBody(t, src, "Read"), "Secret")

	// Update has no response to take it from, so it carries state's value.
	update := funcBody(t, src, "Update")
	require.Contains(t, update, "req.State.Get(ctx, &state)")
	require.Regexp(t, `flattenThing\(readOut, &plan\)[\s\S]*plan\.Secret = state\.Secret`, update)
	assertGolden(t, "createonly.go.golden", src)
}

// An integer id with a pointer create envelope, an OptNilString source, and a
// read payload that carries the field too: flatten must still leave it alone.
func TestEntityCreateOnly_intIDPointerData(t *testing.T) {
	src := generateParsed(t, createOnlyWidget(archetype{CreateOnly: []string{"secret"}}))["widget.go"]

	require.Regexp(t, `configuredSecret := flex\.OptStringToFrameworkNullIfEmpty\(generated\.OptString\{Value: created\.Data\.Secret\.Value, Set: created\.Data\.Secret\.Set && !created\.Data\.Secret\.Null\}\)[\s\S]*readOut, readErr :=`, src)
	require.Contains(t, funcBody(t, src, "Create"), "plan.Secret = configuredSecret")
	require.NotContains(t, funcBody(t, src, "flattenWidget"), "Secret")
	require.Contains(t, funcBody(t, src, "Update"), "plan.Secret = state.Secret")
}

// Without create_only the same read-payload field is flattened as usual.
func TestEntityCreateOnly_absentLeavesFlattenAlone(t *testing.T) {
	src := generateParsed(t, createOnlyWidget(archetype{}))["widget.go"]
	require.Contains(t, funcBody(t, src, "flattenWidget"), "Secret")
	require.NotContains(t, src, "configuredSecret")
}

func TestCreateOnlyCapture_sourceTypes(t *testing.T) {
	for _, tc := range []struct{ typ, want string }{
		{"OptString", "flex.OptStringToFrameworkNullIfEmpty(created.Data.S)"},
		{"string", "flex.OptStringToFrameworkNullIfEmpty(generated.NewOptString(created.Data.S))"},
		{"OptNilString", "flex.OptStringToFrameworkNullIfEmpty(generated.OptString{Value: created.Data.S.Value, Set: created.Data.S.Set && !created.Data.S.Null})"},
	} {
		got, ok := createOnlyCapture(tc.typ, "created.Data.S", "generated")
		require.True(t, ok, tc.typ)
		require.Equal(t, tc.want, got)
	}
	for _, typ := range []string{"OptInt64", "int64", "[]string", "OptBool"} {
		_, ok := createOnlyCapture(typ, "x", "generated")
		require.False(t, ok, typ)
	}
}

// A create_only entry that cannot be honored fails generation by name; a
// dropped attribute is never silent.
func TestEntityCreateOnly_refused(t *testing.T) {
	for _, tc := range []struct {
		name string
		spec fixtureSpec
		want []string
	}{
		{"not in model", createOnlyWidget(archetype{CreateOnly: []string{"nope"}}), []string{"widget", "nope", "not in the model"}},
		{"not on create response", createOnlyWidget(archetype{CreateOnly: []string{"note"}}), []string{"widget", "note", "create response"}},
		{"non-string sdk field", createOnlyWidget(archetype{CreateOnly: []string{"count"}}), []string{"widget", "count", "OptInt64"}},
		{"non-string attribute", createOnlyWidget(archetype{CreateOnly: []string{"size"}}), []string{"widget", "size", "types.Int64"}},
		{"also rewritten", createOnlyWidget(archetype{CreateOnly: []string{"secret"}, Rewritten: []string{"secret"}}), []string{"widget", "secret", "rewritten"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := generateFixture(t, tc.spec)
			require.Error(t, err)
			for _, w := range tc.want {
				require.ErrorContains(t, err, w)
			}
		})
	}
}

func TestCreateOnlyRefusedOutsideEntity(t *testing.T) {
	spec := createOnlyWidget(archetype{CreateOnly: []string{"secret"}})
	spec.Arch.Kind = noReadKind
	_, err := generateFixture(t, spec)
	require.ErrorContains(t, err, "widget")
	require.ErrorContains(t, err, "create_only")
}

// A create answering a bare CreatedResponse has no payload to capture from.
func TestEntityCreateOnly_refusesWithoutCreatePayload(t *testing.T) {
	spec := createOnlyWidget(archetype{})
	model, err := NewFileSource().ModelFields(filepath.Join(spec.Dir, "widget_schema_gen.go"), "WidgetModel")
	require.NoError(t, err)
	rm, err := resolveResource("widget", spec.Ops, spec.DS, fixtureIndexFrom(t, spec.Dir), model, false)
	require.NoError(t, err)
	rm.CreateOnly = []createOnlyAttr{{TF: "secret", ModelGo: "Secret"}}
	rm.Create.RespType, rm.Create.RespFields = "", nil
	_, err = renderEntity(rm)
	require.ErrorContains(t, err, "widget")
	require.ErrorContains(t, err, "secret")
	require.ErrorContains(t, err, "create response")
}
