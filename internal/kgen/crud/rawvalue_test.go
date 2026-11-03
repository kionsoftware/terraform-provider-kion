package crud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func model(attrs ...string) map[string]ModelField {
	out := map[string]ModelField{}
	for _, a := range attrs {
		out[a] = ModelField{TFSDK: a, GoName: pascalCase(a)}
	}
	return out
}

func TestResolveRawValues_collapsesTypedVariants(t *testing.T) {
	t.Parallel()
	body := &Struct{Fields: []Field{
		{GoName: "DefaultValue", JSONName: "default_value", Type: "Raw"},
		{GoName: "Name", JSONName: "name", Type: "OptString"},
	}}
	byTF := model("default_value_string", "default_value_list", "default_value_map", "name")

	got := resolveRawValues(body, byTF)
	require.Len(t, got, 1)
	assert.Equal(t, "DefaultValue", got[0].SDKField)
	assert.Equal(t, "buildDefaultValue", got[0].Func)
	assert.Equal(t, "DefaultValueString", got[0].StringGo)
	assert.Equal(t, "DefaultValueList", got[0].ListGo)
	assert.Equal(t, "DefaultValueMap", got[0].MapGo)
	assert.Equal(t, "default_value_string, default_value_list, or default_value_map", got[0].Attrs)
	// A bare jx.Raw is a required property, so sending none of the variants is
	// an error rather than an omitted field.
	assert.True(t, got[0].Required)
}

// A jx.Raw field the model carries directly is a normalized JSON document
// (scope.criteria), not a polymorphic value split into variants. Treating it as
// one would replace a working bind with a builder for attributes that do not
// exist.
func TestResolveRawValues_skipsDirectlyModeledRaw(t *testing.T) {
	t.Parallel()
	body := &Struct{Fields: []Field{{GoName: "Criteria", JSONName: "criteria", Type: "Raw"}}}
	assert.Empty(t, resolveRawValues(body, model("criteria")))
}

// Nothing stands in for it: leave it to bodyBinds, which refuses loudly rather
// than emitting a builder with no source.
func TestResolveRawValues_skipsWhenNoVariantsExist(t *testing.T) {
	t.Parallel()
	body := &Struct{Fields: []Field{{GoName: "Blob", JSONName: "blob", Type: "Raw"}}}
	assert.Empty(t, resolveRawValues(body, model("name")))
}

func TestResolveRawValues_partialVariantSet(t *testing.T) {
	t.Parallel()
	body := &Struct{Fields: []Field{{GoName: "Value", JSONName: "value", Type: "Raw"}}}

	got := resolveRawValues(body, model("value_string"))
	require.Len(t, got, 1)
	assert.Equal(t, "ValueString", got[0].StringGo)
	assert.Empty(t, got[0].ListGo)
	assert.Empty(t, got[0].MapGo)
	assert.Equal(t, "value_string", got[0].Attrs)
}

// The read half must name the same attributes as the write half; it is derived
// from the binds rather than re-matched so the two cannot disagree.
func TestResolveRawValueFlats_pairsWithBinds(t *testing.T) {
	t.Parallel()
	body := &Struct{Fields: []Field{{GoName: "DefaultValue", JSONName: "default_value", Type: "Raw"}}}
	byTF := model("default_value_string", "default_value_list")
	binds := resolveRawValues(body, byTF)

	resp := []Field{{GoName: "DefaultValue", JSONName: "default_value", Type: "Raw"}}
	got := resolveRawValueFlats(resp, byTF, "v.Data.Value.", binds)
	require.Len(t, got, 1)
	assert.Equal(t, "flattenDefaultValue", got[0].Func)
	assert.Equal(t, "v.Data.Value.DefaultValue", got[0].SDKPath)
	assert.Equal(t, "DefaultValueString", got[0].StringGo)
	assert.Empty(t, got[0].MapGo)
}

func TestResolveRawValueFlats_withoutABindIsNotFlattened(t *testing.T) {
	t.Parallel()
	resp := []Field{{GoName: "DefaultValue", JSONName: "default_value", Type: "Raw"}}
	assert.Empty(t, resolveRawValueFlats(resp, model("default_value_string"), "v.Data.Value.", nil))
}

func TestJoinWithOr(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", joinWithOr(nil))
	assert.Equal(t, "a", joinWithOr([]string{"a"}))
	assert.Equal(t, "a or b", joinWithOr([]string{"a", "b"}))
	assert.Equal(t, "a, b, or c", joinWithOr([]string{"a", "b", "c"}))
}

func TestMergeRawValues_emitsEachHelperOnce(t *testing.T) {
	t.Parallel()
	a := []rawValueBind{{Func: "buildDefaultValue"}, {Func: "buildOther"}}
	b := []rawValueBind{{Func: "buildDefaultValue"}}

	got := mergeRawValues(a, b)
	require.Len(t, got, 2)
	assert.Equal(t, "buildDefaultValue", got[0].Func)
	assert.Equal(t, "buildOther", got[1].Func)
}
