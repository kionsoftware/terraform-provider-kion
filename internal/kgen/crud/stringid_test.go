package crud

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func thingSpec() fixtureSpec {
	return fixtureSpec{
		Name: "thing",
		Dir:  filepath.Join("testdata", "stringid"),
		Ops: resOps{
			Create: &opRef{Method: "POST", Path: "/v3/thing"},
			Read:   &opRef{Method: "GET", Path: "/v3/thing/{id}"},
			Update: &opRef{Method: "PUT", Path: "/v3/thing/{id}"},
			Delete: &opRef{Method: "DELETE", Path: "/v3/thing/{id}"},
		},
		DS:         dsOps{Read: &opRef{Method: "GET", Path: "/v3/thing"}},
		TestValues: filepath.Join("testdata", "stringid", "test_values.yaml"),
	}
}

func generateThing(t *testing.T) map[string]string {
	t.Helper()
	files, err := generateFixture(t, thingSpec())
	require.NoError(t, err)
	for name, src := range files {
		_, err := parser.ParseFile(token.NewFileSet(), name, src, parser.ParseComments)
		require.NoError(t, err, "%s does not parse:\n%s", name, src)
	}
	return files
}

// A string API id stays a string end to end: nothing parses or formats it as
// an integer. Refresh the golden with UPDATE_GOLDEN=1.
func TestEntityStringID(t *testing.T) {
	src := generateThing(t)["thing.go"]

	require.NotContains(t, src, "strconv.ParseInt(state.Id.ValueString()")
	require.Contains(t, src, "id := state.Id.ValueString()")
	require.Contains(t, src, `fmt.Sprintf("reading %s (ID: %s)", ResNameThing, id)`)
	require.Contains(t, src, "plan.Id = types.StringValue(id)")
	require.Contains(t, src, "errs.RawCreatedStringID(")

	require.NotContains(t, src, "strconv")
	require.NotContains(t, src, "%d")
	require.Contains(t, src, "generated.GetThingParams{ID: id}")
	require.Contains(t, src, "generated.PutThingParams{ID: id}")
	require.Contains(t, src, "generated.DeleteThingParams{ID: id}")
	require.Contains(t, src, "model.Id = types.StringValue(v.Data.Value.ID.Value)")
	assertGolden(t, "stringid.go.golden", src)
}

func TestEntityStringID_dataSource(t *testing.T) {
	src := generateThing(t)["thing_data_source.go"]

	require.Contains(t, src, `"id": schema.StringAttribute{`)
	require.Contains(t, src, "generated.GetThingParams{ID: data.Id.ValueString()}")
	require.Regexp(t, "Id +types.String +`tfsdk:\"id\"`", src)
	require.Contains(t, src, "types.StringValue(lbl.ID.Value)")
	require.NotContains(t, src, "ValueInt64()")
	require.NotContains(t, src, "int64(lbl.ID.Value)")
}

func TestEntityStringID_sweeper(t *testing.T) {
	src := generateThing(t)["sweep.go"]

	require.Contains(t, src, "resource.AddTestSweepers(")
	require.Contains(t, src, "var ids []string")
	require.Contains(t, src, "ids = append(ids, item.ID.Value)")
	require.Contains(t, src, "generated.DeleteThingParams{ID: id}")
	require.Contains(t, src, `"deleting kion_thing (%s): %w"`)
	require.NotContains(t, src, "int64(item.ID.Value)")
}

func TestEntityStringID_acceptanceTest(t *testing.T) {
	files := generateThing(t)
	src := files["thing_test.go"]

	require.NotContains(t, src, "strconv")
	require.Contains(t, src, "id := rs.Primary.ID")
	require.Contains(t, src, `"reading kion_thing (%s): %w"`)
	require.Contains(t, src, "generated.GetThingParams{ID: id}")
	require.Contains(t, files, "thing_data_source_test.go")
}

// Only the entity archetype carries string ids; every other one parses ids as
// integers, so it must refuse rather than emit code that cannot compile.
func TestStringIDRefusedOutsideEntity(t *testing.T) {
	for _, kind := range []string{noReadKind, assocKind, blendedKind, parentListKind, compoundKindParentRead} {
		t.Run(kind, func(t *testing.T) {
			spec := thingSpec()
			spec.Arch = &archetype{Kind: kind}
			_, err := generateFixture(t, spec)
			require.Error(t, err)
			require.Contains(t, err.Error(), "thing")
			require.Contains(t, err.Error(), "string id")
		})
	}
}

// Entity features that address the record by an integer id are refused for a
// string id instead of rendered.
func TestEntityStringID_refusesIntOnlyFeatures(t *testing.T) {
	rm := thingResourceModel(t)
	rm.Labels = &labelSyncBind{ModelGo: "Name"}
	_, err := renderEntity(rm)
	require.ErrorContains(t, err, "thing")
	require.ErrorContains(t, err, "string id")

	rm = thingResourceModel(t)
	rm.ReadCompanion = "readThingExtras"
	_, err = renderEntity(rm)
	require.ErrorContains(t, err, "string id")
}

func thingResourceModel(t *testing.T) ResourceModel {
	t.Helper()
	spec := thingSpec()
	model, err := NewFileSource().ModelFields(filepath.Join(spec.Dir, "thing_schema_gen.go"), "ThingModel")
	require.NoError(t, err)
	rm, err := resolveResource("thing", spec.Ops, spec.DS, fixtureIndexFrom(t, spec.Dir), model, false)
	require.NoError(t, err)
	return rm
}

// A string id cannot be cast to or from an integer one, so mixing them across
// ops is refused; same-typed and int-to-int cases are unchanged.
func TestIDExpr_stringID(t *testing.T) {
	for _, tc := range []struct{ op, parse, want string }{
		{"", "string", "id"},
		{"string", "string", "id"},
		{"", "int64", "idInt"},
		{"uint64", "int64", "uint64(idInt)"},
	} {
		got, err := idExpr(tc.op, tc.parse)
		require.NoError(t, err)
		require.Equal(t, tc.want, got)
	}
	_, err := idExpr("int64", "string")
	require.Error(t, err)
	_, err = idExpr("string", "uint64")
	require.Error(t, err)
}

// An update body that wraps the record in an object carrying its id must send
// the string id, not leave the sub-field zero.
func TestEntityStringID_wrappedUpdateID(t *testing.T) {
	spec := thingSpec()
	spec.Dir = filepath.Join("testdata", "stringidwrap")
	spec.TestValues = ""
	files, err := generateFixture(t, spec)
	require.NoError(t, err)
	src := files["thing.go"]
	require.Contains(t, src, "ID:          generated.OptString{Value: id, Set: true},")
	require.NotContains(t, src, "idInt")
}

// A wrapped id sub-field whose type does not match the record id cannot be
// sourced; generation fails rather than sending a zero id.
func TestWrappedIDSubFieldMismatchIsAnError(t *testing.T) {
	idx := sdkIndex{structs: map[string]Struct{
		"Ref": {Name: "Ref", Fields: []Field{
			{GoName: "ID", JSONName: "id", Type: "OptInt64"},
			{GoName: "Name", JSONName: "name", Type: "OptString"},
		}},
	}}
	body := &Struct{Name: "Body", Fields: []Field{{GoName: "Ref", JSONName: "ref", Type: "OptRef"}}}
	byTF := map[string]ModelField{
		"id":   {GoName: "Id", TFSDK: "id", Type: "types.String"},
		"name": {GoName: "Name", TFSDK: "name", Type: "types.String"},
	}
	schemaGen := filepath.Join("testdata", "stringid", "thing_schema_gen.go")

	_, err := resolveNested(NewFileSource(), schemaGen, body, byTF, idx, nestedOpts{IDAttr: "id", IDVar: "id", IDType: "string"})
	require.ErrorContains(t, err, `"ref"`)
	require.ErrorContains(t, err, "OptInt64")

	idx.structs["Ref"].Fields[0].Type = "OptString"
	_, err = resolveNested(NewFileSource(), schemaGen, body, byTF, idx, nestedOpts{IDAttr: "id", IDVar: "idInt", IDType: "int64"})
	require.ErrorContains(t, err, "OptString")
}
