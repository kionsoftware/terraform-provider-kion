package schemas

import (
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	fsmocks "terraform-provider-kion/internal/kgen/kfs/mocks"
	runnermocks "terraform-provider-kion/internal/kgen/schemas/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestApplyRenames verifies a body-attribute rename is applied to the provider
// code spec (the thing tfplugingen itself cannot do).
func TestApplyRenames(t *testing.T) {
	m := fsmocks.NewMockFS(t)

	renamesYAML := "gcp_regions:\n  data: regions\n"
	specJSON := `{"datasources":[{"name":"gcp_regions","schema":{"attributes":[` +
		`{"name":"data","list":{"computed_optional_required":"computed"}}]}}]}`

	m.EXPECT().ReadFile("/p/renames.yaml").Return([]byte(renamesYAML), nil)
	m.EXPECT().ReadFile("/p/spec.json").Return([]byte(specJSON), nil)

	var written []byte
	m.EXPECT().WriteFile("/p/spec.json", mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ fs.FileMode) { written = data }).
		Return(nil)

	g := &generator{fs: m}
	require.NoError(t, g.applyRenames("/p/spec.json", "/p/renames.yaml"))

	assert.Contains(t, string(written), `"name": "regions"`)
	assert.NotContains(t, string(written), `"name": "data"`)
}

// TestApplyRenames_noSidecar verifies a missing renames file is a no-op: it
// never reads or writes the spec. NewMockFS asserts no unexpected calls occur.
func TestApplyRenames_noSidecar(t *testing.T) {
	m := fsmocks.NewMockFS(t)
	m.EXPECT().ReadFile("/p/renames.yaml").Return(nil, os.ErrNotExist)

	g := &generator{fs: m}
	require.NoError(t, g.applyRenames("/p/spec.json", "/p/renames.yaml"))
}

// TestApplySchemaOverrides verifies the sidecar retypes an attribute, injects a
// plan modifier with its import, and sets the schema description, the things
// tfplugingen can't derive from the OpenAPI spec.
func TestApplySchemaOverrides(t *testing.T) {
	m := fsmocks.NewMockFS(t)

	overYAML := "resources:\n" +
		"  label:\n" +
		"    description: Manages a Kion Label.\n" +
		"    attributes:\n" +
		"      id:\n" +
		"        type: string\n" +
		"        computed_optional_required: computed\n" +
		"        plan_modifiers:\n" +
		"          - stringplanmodifier.UseStateForUnknown()\n"
	specJSON := `{"resources":[{"name":"label","schema":{"attributes":[` +
		`{"name":"id","int64":{"computed_optional_required":"computed_optional","description":"ID"}}]}}]}`

	m.EXPECT().ReadFile("/p/over.yaml").Return([]byte(overYAML), nil)
	m.EXPECT().ReadFile("/p/spec.json").Return([]byte(specJSON), nil)

	var written []byte
	m.EXPECT().WriteFile("/p/spec.json", mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ fs.FileMode) { written = data }).
		Return(nil)

	g := &generator{fs: m}
	require.NoError(t, g.applySchemaOverrides("/p/spec.json", "/p/over.yaml"))

	s := string(written)
	assert.Contains(t, s, `"string"`, "id retyped to string")
	assert.NotContains(t, s, `"int64"`, "int64 type node removed")
	assert.Contains(t, s, `"computed_optional_required": "computed"`)
	assert.Contains(t, s, "stringplanmodifier.UseStateForUnknown()")
	assert.Contains(t, s, "resource/schema/stringplanmodifier", "plan-modifier import derived")
	assert.Contains(t, s, "Manages a Kion Label.", "schema description set")
}

// TestApplySchemaOverrides_noSidecar: with no sidecar there are no per-type
// overrides, but the global string-id default still runs. So the spec is read,
// (potentially) mutated, and written back.
// TestApplySchemaOverrides_retypeNestedToCustomScalar covers retyping a
// single_nested attribute (with its generated custom_type + empty attributes)
// into a scalar carrying a framework custom type, the scope_criteria `criteria`
// case (object → jsontypes.Normalized string). The stale nested keys must be
// dropped and the new custom_type emitted.
func TestApplySchemaOverrides_retypeNestedToCustomScalar(t *testing.T) {
	m := fsmocks.NewMockFS(t)

	overYAML := "resources:\n" +
		"  scope_criteria:\n" +
		"    attributes:\n" +
		"      criteria:\n" +
		"        type: string\n" +
		"        computed_optional_required: required\n" +
		"        custom_type:\n" +
		"          import: github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes\n" +
		"          type: jsontypes.NormalizedType{}\n" +
		"          value_type: jsontypes.Normalized\n"
	specJSON := `{"resources":[{"name":"scope_criteria","schema":{"attributes":[` +
		`{"name":"criteria","single_nested":{"computed_optional_required":"optional",` +
		`"custom_type":{"type":"CriteriaType"},"attributes":[]}}]}}]}`

	m.EXPECT().ReadFile("/p/over.yaml").Return([]byte(overYAML), nil)
	m.EXPECT().ReadFile("/p/spec.json").Return([]byte(specJSON), nil)

	var written []byte
	m.EXPECT().WriteFile("/p/spec.json", mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ fs.FileMode) { written = data }).
		Return(nil)

	g := &generator{fs: m}
	require.NoError(t, g.applySchemaOverrides("/p/spec.json", "/p/over.yaml"))

	s := string(written)
	assert.Contains(t, s, `"string"`, "criteria retyped to string")
	assert.NotContains(t, s, `"single_nested"`, "single_nested node removed")
	assert.NotContains(t, s, "CriteriaType", "stale generated custom_type dropped")
	assert.Contains(t, s, "jsontypes.NormalizedType{}", "custom_type type emitted")
	assert.Contains(t, s, "jsontypes.Normalized", "custom_type value_type emitted")
	assert.Contains(t, s, "terraform-plugin-framework-jsontypes/jsontypes", "custom_type import emitted")
	assert.Contains(t, s, `"computed_optional_required": "required"`)
}

func TestApplySchemaOverrides_noSidecar(t *testing.T) {
	m := fsmocks.NewMockFS(t)
	m.EXPECT().ReadFile("/p/over.yaml").Return(nil, os.ErrNotExist)
	m.EXPECT().ReadFile("/p/spec.json").Return([]byte(`{"resources":[]}`), nil)
	m.EXPECT().WriteFile("/p/spec.json", mock.Anything, os.FileMode(0o600)).Return(nil)

	g := &generator{fs: m}
	require.NoError(t, g.applySchemaOverrides("/p/spec.json", "/p/over.yaml"))
}

// TestApplySchemaOverrides_addsMissingAttribute verifies an override naming an
// attribute absent from the spec, because tfplugingen dropped it (e.g. a
// polymorphic value split into typed sub-attributes). Is *inserted* as a new
// attribute node rather than silently ignored.
func TestApplySchemaOverrides_addsMissingAttribute(t *testing.T) {
	m := fsmocks.NewMockFS(t)

	overYAML := "resources:\n" +
		"  custom_variable:\n" +
		"    attributes:\n" +
		"      default_value_string:\n" +
		"        type: string\n" +
		"        computed_optional_required: optional\n" +
		"        description: String default value.\n"
	// The spec has only id; the polymorphic default_value was dropped upstream.
	specJSON := `{"resources":[{"name":"custom_variable","schema":{"attributes":[` +
		`{"name":"id","string":{"computed_optional_required":"computed"}}]}}]}`

	m.EXPECT().ReadFile("/p/over.yaml").Return([]byte(overYAML), nil)
	m.EXPECT().ReadFile("/p/spec.json").Return([]byte(specJSON), nil)

	var written []byte
	m.EXPECT().WriteFile("/p/spec.json", mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ fs.FileMode) { written = data }).
		Return(nil)

	g := &generator{fs: m}
	require.NoError(t, g.applySchemaOverrides("/p/spec.json", "/p/over.yaml"))

	s := string(written)
	assert.Contains(t, s, `"name": "default_value_string"`, "missing attribute inserted")
	assert.Contains(t, s, `"string"`, "inserted attr carries its type")
	assert.Contains(t, s, "String default value.", "inserted attr carries description")
	assert.Contains(t, s, `"computed_optional_required": "optional"`)
}

// TestApplySchemaOverrides_addsCollectionElementType verifies an inserted list
// (or map) attribute carries the element_type node tfplugingen-framework
// requires, e.g. a list of strings becomes {"list": {"element_type": {"string": {}}}}.
func TestApplySchemaOverrides_addsCollectionElementType(t *testing.T) {
	m := fsmocks.NewMockFS(t)

	overYAML := "resources:\n" +
		"  custom_variable:\n" +
		"    attributes:\n" +
		"      default_value_list:\n" +
		"        type: list\n" +
		"        element_type: string\n" +
		"        computed_optional_required: optional\n"
	specJSON := `{"resources":[{"name":"custom_variable","schema":{"attributes":[` +
		`{"name":"id","string":{"computed_optional_required":"computed"}}]}}]}`

	m.EXPECT().ReadFile("/p/over.yaml").Return([]byte(overYAML), nil)
	m.EXPECT().ReadFile("/p/spec.json").Return([]byte(specJSON), nil)

	var written []byte
	m.EXPECT().WriteFile("/p/spec.json", mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ fs.FileMode) { written = data }).
		Return(nil)

	g := &generator{fs: m}
	require.NoError(t, g.applySchemaOverrides("/p/spec.json", "/p/over.yaml"))

	// Parse back to assert the element_type structure precisely.
	var spec map[string]any
	require.NoError(t, json.Unmarshal(written, &spec))
	resources, ok := spec["resources"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, resources)
	res0, ok := resources[0].(map[string]any)
	require.True(t, ok)
	schemaObj, ok := res0["schema"].(map[string]any)
	require.True(t, ok)
	attrs, ok := schemaObj["attributes"].([]any)
	require.True(t, ok)
	var listNode map[string]any
	for _, a := range attrs {
		am, ok := a.(map[string]any)
		if !ok {
			continue
		}
		if am["name"] == "default_value_list" {
			listNode = am
		}
	}
	require.NotNil(t, listNode, "list attribute inserted")
	listType, ok := listNode["list"].(map[string]any)
	require.True(t, ok, "typed as list")
	elem, ok := listType["element_type"].(map[string]any)
	require.True(t, ok, "list carries element_type")
	_, ok = elem["string"]
	assert.True(t, ok, "element_type is string")
}

// TestApplySchemaOverrides_addsToEmptySchema verifies overrides can add
// attributes to a resource whose schema has no attributes array at all, e.g. an
// association whose entire body was dropped (tfplugingen emits an empty schema,
// which it then rejects). Reconstruction must repopulate it.
func TestApplySchemaOverrides_addsToEmptySchema(t *testing.T) {
	m := fsmocks.NewMockFS(t)

	overYAML := "resources:\n" +
		"  global_permission_mapping:\n" +
		"    attributes:\n" +
		"      app_role_id:\n" +
		"        type: int64\n" +
		"        computed_optional_required: required\n"
	specJSON := `{"resources":[{"name":"global_permission_mapping","schema":{}}]}`

	m.EXPECT().ReadFile("/p/over.yaml").Return([]byte(overYAML), nil)
	m.EXPECT().ReadFile("/p/spec.json").Return([]byte(specJSON), nil)

	var written []byte
	m.EXPECT().WriteFile("/p/spec.json", mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ fs.FileMode) { written = data }).
		Return(nil)

	g := &generator{fs: m}
	require.NoError(t, g.applySchemaOverrides("/p/spec.json", "/p/over.yaml"))

	node := findWrittenAttr(t, written, "app_role_id")
	require.NotNil(t, node, "attribute added to a previously-empty schema")
	_, ok := node["int64"].(map[string]any)
	assert.True(t, ok, "carries its type")
}

// TestApplySchemaOverrides_addsSingleNested verifies an inserted single_nested
// attribute carries its nested `attributes` array (recursively built from the
// override), as tfplugingen-framework requires for a SingleNestedAttribute.
func TestApplySchemaOverrides_addsSingleNested(t *testing.T) {
	m := fsmocks.NewMockFS(t)

	overYAML := "resources:\n" +
		"  aws_account:\n" +
		"    attributes:\n" +
		"      aws_organizational_unit:\n" +
		"        type: single_nested\n" +
		"        computed_optional_required: optional\n" +
		"        attributes:\n" +
		"          name:\n" +
		"            type: string\n" +
		"            computed_optional_required: optional\n" +
		"          org_unit_id:\n" +
		"            type: string\n" +
		"            computed_optional_required: optional\n"
	specJSON := `{"resources":[{"name":"aws_account","schema":{"attributes":[` +
		`{"name":"id","string":{"computed_optional_required":"computed"}}]}}]}`

	m.EXPECT().ReadFile("/p/over.yaml").Return([]byte(overYAML), nil)
	m.EXPECT().ReadFile("/p/spec.json").Return([]byte(specJSON), nil)

	var written []byte
	m.EXPECT().WriteFile("/p/spec.json", mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ fs.FileMode) { written = data }).
		Return(nil)

	g := &generator{fs: m}
	require.NoError(t, g.applySchemaOverrides("/p/spec.json", "/p/over.yaml"))

	node := findWrittenAttr(t, written, "aws_organizational_unit")
	require.NotNil(t, node, "single_nested attribute inserted")
	sn, ok := node["single_nested"].(map[string]any)
	require.True(t, ok, "typed as single_nested")
	nested, ok := sn["attributes"].([]any)
	require.True(t, ok, "single_nested carries nested attributes")
	names := map[string]bool{}
	for _, na := range nested {
		child, ok := na.(map[string]any)
		require.True(t, ok)
		cn, ok := child["name"].(string)
		require.True(t, ok)
		names[cn] = true
	}
	assert.True(t, names["name"] && names["org_unit_id"], "nested attrs built recursively")
}

// findWrittenAttr parses the written provider code spec and returns the first
// resource's top-level attribute node with the given name (nil if absent),
// using checked type assertions throughout.
func findWrittenAttr(t *testing.T, written []byte, name string) map[string]any {
	t.Helper()
	var spec map[string]any
	require.NoError(t, json.Unmarshal(written, &spec))
	resources, ok := spec["resources"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, resources)
	res0, ok := resources[0].(map[string]any)
	require.True(t, ok)
	schemaObj, ok := res0["schema"].(map[string]any)
	require.True(t, ok)
	attrs, ok := schemaObj["attributes"].([]any)
	require.True(t, ok)
	for _, a := range attrs {
		am, ok := a.(map[string]any)
		if ok && am["name"] == name {
			return am
		}
	}
	return nil
}

// TestApplySchemaOverrides_addsSetNested verifies an inserted set_nested attribute
// wraps its nested attributes in a nested_object, as the code spec requires.
func TestApplySchemaOverrides_addsSetNested(t *testing.T) {
	m := fsmocks.NewMockFS(t)

	overYAML := "resources:\n" +
		"  project:\n" +
		"    attributes:\n" +
		"      budget:\n" +
		"        type: set_nested\n" +
		"        computed_optional_required: optional\n" +
		"        attributes:\n" +
		"          amount:\n" +
		"            type: string\n" +
		"            computed_optional_required: optional\n"
	specJSON := `{"resources":[{"name":"project","schema":{"attributes":[` +
		`{"name":"id","string":{"computed_optional_required":"computed"}}]}}]}`

	m.EXPECT().ReadFile("/p/over.yaml").Return([]byte(overYAML), nil)
	m.EXPECT().ReadFile("/p/spec.json").Return([]byte(specJSON), nil)

	var written []byte
	m.EXPECT().WriteFile("/p/spec.json", mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ fs.FileMode) { written = data }).
		Return(nil)

	g := &generator{fs: m}
	require.NoError(t, g.applySchemaOverrides("/p/spec.json", "/p/over.yaml"))

	node := findWrittenAttr(t, written, "budget")
	require.NotNil(t, node, "set_nested attribute inserted")
	sn, ok := node["set_nested"].(map[string]any)
	require.True(t, ok, "typed as set_nested")
	no, ok := sn["nested_object"].(map[string]any)
	require.True(t, ok, "set_nested wraps in nested_object")
	nested, ok := no["attributes"].([]any)
	require.True(t, ok, "nested_object carries attributes")
	require.NotEmpty(t, nested)
	first, ok := nested[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "amount", first["name"])
}

// TestRenameSharedDecls verifies that generated top-level types/funcs a data
// source shares with its resource (nested-object CustomTypes) are renamed in the
// data-source file so the two don't redeclare each other in the same package.
func TestRenameSharedDecls(t *testing.T) {
	resourceSrc := "package ami\n" +
		"type ExpiresAtType struct{}\n" +
		"func (t ExpiresAtType) Foo() {}\n" +
		"func NewExpiresAtValueNull() ExpiresAtValue { return ExpiresAtValue{} }\n" +
		"type ExpiresAtValue struct{}\n"
	dsSrc := "package ami\n" +
		"type ExpiresAtType struct{}\n" +
		"func NewExpiresAtValueNull() ExpiresAtValue { return ExpiresAtValue{} }\n" +
		"type ExpiresAtValue struct{}\n" +
		"type Unrelated struct{}\n"

	got := renameSharedDecls(dsSrc, topLevelDecls(resourceSrc))

	// Shared decls (and their references) are suffixed.
	assert.Contains(t, got, "type ExpiresAtTypeDataSource struct{}")
	assert.Contains(t, got, "type ExpiresAtValueDataSource struct{}")
	assert.Contains(t, got, "func NewExpiresAtValueNullDataSource() ExpiresAtValueDataSource")
	assert.Contains(t, got, "return ExpiresAtValueDataSource{}")
	// Non-shared decls are untouched.
	assert.Contains(t, got, "type Unrelated struct{}")
	// The original colliding identifiers no longer appear as declarations.
	assert.NotContains(t, got, "type ExpiresAtType struct")
}

// TestDedupTopLevelDecls verifies duplicate top-level declarations (which
// tfplugingen emits when the same nested object appears at multiple paths in one
// schema) are collapsed to a single copy so the file compiles.
func TestDedupTopLevelDecls(t *testing.T) {
	src := "package p\n\n" +
		"type CloudRuleType struct{ x int }\n\n" +
		"func (t CloudRuleType) Foo() {}\n\n" +
		"type Keep struct{}\n\n" +
		"type CloudRuleType struct{ x int }\n\n" +
		"func (t CloudRuleType) Foo() {}\n"

	got := dedupTopLevelDecls(src)

	assert.Equal(t, 1, strings.Count(got, "type CloudRuleType struct"), "duplicate type collapsed")
	assert.Equal(t, 1, strings.Count(got, "func (t CloudRuleType) Foo()"), "duplicate method collapsed")
	assert.Contains(t, got, "type Keep struct{}", "unique decls preserved")
	// Result must still parse.
	_, perr := parser.ParseFile(token.NewFileSet(), "x.go", got, 0)
	assert.NoError(t, perr, "deduped source still parses")
}

// unguardedValueFromObject is tfplugingen's output for a nested-object
// CustomType: it reads in.Attributes() with no null/unknown check, even though
// the inverse ToObjectValue and the sibling ValueFromTerraform both have one.
const unguardedValueFromObject = `package p

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (t AzurePolicyType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	descriptionAttribute, ok := attributes["description"]

	if !ok {
		diags.AddError("Attribute Missing", ` + "`description is missing from object`" + `)

		return nil, diags
	}

	_ = descriptionAttribute

	return NewAzurePolicyValueMust(nil, attributes), diags
}
`

// TestGuardValueFromObject verifies the null/unknown guard is injected into a
// generated ValueFromObject. Without it, an Optional+Computed single-nested
// attribute carrying objectplanmodifier.UseStateForUnknown fails every plan:
// the framework round-trips the plan value through ToObjectValue -> plan
// modifiers -> ValueFromObject, and an unknown ObjectValue has no attributes,
// so the generated code reports `Attribute Missing`.
func TestGuardValueFromObject(t *testing.T) {
	got := guardValueFromObject(unguardedValueFromObject)

	assert.Contains(t, got, "if in.IsNull()", "null guard injected")
	assert.Contains(t, got, "return NewAzurePolicyValueNull(), diags", "null returns the typed null value")
	assert.Contains(t, got, "if in.IsUnknown()", "unknown guard injected")
	assert.Contains(t, got, "return NewAzurePolicyValueUnknown(), diags", "unknown returns the typed unknown value")

	// The guards must precede the attribute lookup that would otherwise fail.
	assert.Less(t, strings.Index(got, "if in.IsUnknown()"), strings.Index(got, "in.Attributes()"),
		"guards precede the attribute read")

	_, perr := parser.ParseFile(token.NewFileSet(), "x.go", got, 0)
	assert.NoError(t, perr, "guarded source still parses")
}

// TestGuardValueFromObject_idempotent verifies a second pass is a no-op, so
// regenerating over already-guarded output does not stack duplicate guards.
func TestGuardValueFromObject_idempotent(t *testing.T) {
	once := guardValueFromObject(unguardedValueFromObject)
	twice := guardValueFromObject(once)

	assert.Equal(t, once, twice, "already-guarded source is unchanged")
	assert.Equal(t, 1, strings.Count(twice, "if in.IsNull()"), "guard not duplicated")
}

// TestGuardValueFromObject_leavesOtherFuncsAlone verifies the transform touches
// only ValueFromObject methods, and returns unparseable source unchanged.
func TestGuardValueFromObject_leavesOtherFuncsAlone(t *testing.T) {
	src := "package p\n\nfunc (t FooType) ValueFromTerraform(in int) int { return in }\n"

	got := guardValueFromObject(src)

	assert.NotContains(t, got, "IsNull()", "unrelated methods untouched")
	assert.Equal(t, "not go source", guardValueFromObject("not go source"), "unparseable source returned as-is")
}

// TestMergeSpecAdditions verifies OpenAPI fragments (paths + component schemas)
// from the additions sidecar are merged into the spec, used to supply endpoints
// the public spec omits (e.g. dashboard's private read) so tfplugingen can
// generate them.
func TestMergeSpecAdditions(t *testing.T) {
	spec := `{"paths":{"/beta/dashboard":{"post":{}}},"components":{"schemas":{"Existing":{"type":"object"}}}}`
	additions := "" +
		"paths:\n" +
		"  /beta/dashboard/{id}:\n" +
		"    get:\n" +
		"      responses:\n" +
		"        '200':\n" +
		"          description: ok\n" +
		"components:\n" +
		"  schemas:\n" +
		"    Dashboard:\n" +
		"      type: object\n" +
		"      properties:\n" +
		"        name: { type: string }\n"

	merged, err := mergeSpecAdditions([]byte(spec), []byte(additions))
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(merged, &out))
	paths, ok := out["paths"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, paths, "/beta/dashboard", "existing path preserved")
	assert.Contains(t, paths, "/beta/dashboard/{id}", "added path merged")
	comps, ok := out["components"].(map[string]any)
	require.True(t, ok)
	schemas, ok := comps["schemas"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, schemas, "Existing", "existing schema preserved")
	assert.Contains(t, schemas, "Dashboard", "added schema merged")
}

// TestMergeSpecAdditions_verbLevel verifies adding a verb to an EXISTING path
// preserves the path's other verbs (e.g. add GET to a path that has PATCH/DELETE).
func TestMergeSpecAdditions_verbLevel(t *testing.T) {
	spec := `{"paths":{"/beta/dashboard/{id}":{"patch":{"x":1},"delete":{"y":2}}}}`
	additions := "" +
		"paths:\n" +
		"  /beta/dashboard/{id}:\n" +
		"    get:\n" +
		"      responses:\n" +
		"        '200':\n" +
		"          description: ok\n"

	merged, err := mergeSpecAdditions([]byte(spec), []byte(additions))
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(merged, &out))
	mergedPaths, ok := out["paths"].(map[string]any)
	require.True(t, ok)
	verbs, ok := mergedPaths["/beta/dashboard/{id}"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, verbs, "get", "added verb present")
	assert.Contains(t, verbs, "patch", "existing verb preserved")
	assert.Contains(t, verbs, "delete", "existing verb preserved")
}

// TestGenerate_toolMissing verifies a missing codegen tool is reported with a
// pointer to the install target, using the mocked command runner.
func TestGenerate_toolMissing(t *testing.T) {
	r := runnermocks.NewMockRunner(t)
	r.EXPECT().LookPath("tfplugingen-openapi").Return(errors.New("not found"))

	g := &generator{run: r}
	_, err := g.generate(Options{ProjectRoot: t.TempDir()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "make install-codegen-tools")
}

// TestDistribute verifies a generated schema file is repackaged into its
// service package and a schema unit test is emitted alongside it.
func TestDistribute(t *testing.T) {
	m := fsmocks.NewMockFS(t)

	out := "/w/out"
	m.EXPECT().ReadDir(out).Return([]os.DirEntry{fakeDirEntry{name: "resource_foo", dir: true}}, nil)
	m.EXPECT().Stat("/proj/internal/service/foo").Return(fakeFileInfo{dir: true}, nil)
	m.EXPECT().ReadFile("/w/out/resource_foo/foo_resource_gen.go").
		Return([]byte("package resource_foo\n\nfunc FooResourceSchema() {}\n"), nil)

	// Constructor discovery: scan the package's hand-written source.
	m.EXPECT().ReadDir("/proj/internal/service/foo").
		Return([]os.DirEntry{fakeDirEntry{name: "foo.go"}}, nil)
	m.EXPECT().ReadFile("/proj/internal/service/foo/foo.go").
		Return([]byte("package foo\n\nfunc NewFooResource() resource.Resource { return nil }\n"), nil)

	var schemaBytes, testBytes []byte
	m.EXPECT().WriteFile("/proj/internal/service/foo/foo_schema_gen.go", mock.Anything, mock.Anything).
		Run(func(_ string, d []byte, _ fs.FileMode) { schemaBytes = d }).Return(nil)
	m.EXPECT().WriteFile("/proj/internal/service/foo/foo_schema_gen_test.go", mock.Anything, mock.Anything).
		Run(func(_ string, d []byte, _ fs.FileMode) { testBytes = d }).Return(nil)

	k := kind{
		outDir:       out,
		dirPrefix:    "resource_",
		srcSuffix:    "_resource_gen.go",
		dstSuffix:    "_schema_gen.go",
		schemaFuncRe: regexp.MustCompile(`func ([A-Za-z0-9]+)ResourceSchema`),
		ctorRe:       regexp.MustCompile(`func (New[A-Za-z0-9_]+Resource)\(\)\s+resource\.Resource`),
		ctorKind:     "resource",
		testSuffix:   "_schema_gen_test.go",
		testTmpl:     resourceTestTmpl,
	}

	g := &generator{fs: m}
	written, err := g.distribute("/proj", k)
	require.NoError(t, err)

	assert.Len(t, written, 2)
	assert.Contains(t, string(schemaBytes), "package foo")
	assert.NotContains(t, string(schemaBytes), "package resource_foo")
	assert.Contains(t, string(testBytes), "func TestFooResourceSchema")
	assert.Contains(t, string(testBytes), "NewFooResource()")
}

// --- minimal fakes for os.DirEntry / os.FileInfo used by the mock returns ---

type fakeDirEntry struct {
	name string
	dir  bool
}

func (f fakeDirEntry) Name() string             { return f.name }
func (f fakeDirEntry) IsDir() bool              { return f.dir }
func (fakeDirEntry) Type() fs.FileMode          { return 0 }
func (fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

type fakeFileInfo struct{ dir bool }

func (fakeFileInfo) Name() string       { return "" }
func (fakeFileInfo) Size() int64        { return 0 }
func (fakeFileInfo) Mode() fs.FileMode  { return 0 }
func (fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool      { return f.dir }
func (fakeFileInfo) Sys() any           { return nil }

func TestApplyStringIDDefault(t *testing.T) {
	mkRes := func(name, idType string) map[string]any {
		return map[string]any{
			"name": name,
			"schema": map[string]any{
				"attributes": []any{
					map[string]any{"name": "id", idType: map[string]any{"computed_optional_required": "computed"}},
				},
			},
		}
	}
	resources := []any{mkRes("plain", "int64"), mkRes("explicit", "int64")}
	explicit := map[string]bool{"explicit": true}

	if err := applyStringIDDefault(resources, explicit); err != nil {
		t.Fatalf("applyStringIDDefault: %v", err)
	}

	idKey := func(t *testing.T, res any) string {
		t.Helper()
		obj, ok := res.(map[string]any)
		require.True(t, ok)
		schemaObj, ok := obj["schema"].(map[string]any)
		require.True(t, ok)
		attrs, ok := schemaObj["attributes"].([]any)
		require.True(t, ok)
		am, ok := attrs[0].(map[string]any)
		require.True(t, ok)
		for k := range am {
			if k != "name" {
				return k
			}
		}
		return ""
	}
	if got := idKey(t, resources[0]); got != "string" {
		t.Errorf("plain resource id type = %q, want string (global default)", got)
	}
	if got := idKey(t, resources[1]); got != "int64" {
		t.Errorf("explicit-override resource id type = %q, want int64 (skipped)", got)
	}
}

// TestCheckNothingSkipped covers the case tfplugingen-openapi reports by logging
// and then exiting 0: a configured type it could not map is absent from the code
// spec, and without this check the run looks successful while the stale generated
// file stays on disk.
func TestCheckNothingSkipped(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "generator_config.yaml")
	require.NoError(t, os.WriteFile(cfg, []byte(
		"resources:\n  kept:\n  dropped:\ndata_sources:\n  kept_ds:\n"), 0o600))

	spec := filepath.Join(dir, "spec.json")
	write := func(body string) { require.NoError(t, os.WriteFile(spec, []byte(body), 0o600)) }
	g := &generator{}

	write(`{"resources":[{"name":"kept"},{"name":"dropped"}],"datasources":[{"name":"kept_ds"}]}`)
	require.NoError(t, g.checkNothingSkipped(spec, cfg), "nothing missing should pass")

	write(`{"resources":[{"name":"kept"}],"datasources":[]}`)
	err := g.checkNothingSkipped(spec, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resource dropped")
	assert.Contains(t, err.Error(), "data source kept_ds")
	assert.Contains(t, err.Error(), "property_types")
}

// TestUseStateForUnknownDefault_keptWithRequiresReplace: declaring
// RequiresReplace on a Computed attribute must not drop UseStateForUnknown, or
// a change to any other attribute leaves the unset one unknown and replaces the
// record (kion_project_cloud_access_role.cloud_access_role_type_id did).
func TestUseStateForUnknownDefault_keptWithRequiresReplace(t *testing.T) {
	mod := func(def string) any {
		return map[string]any{"custom": map[string]any{
			"imports":           []any{map[string]any{"path": planModifierBase + "int64planmodifier"}},
			"schema_definition": def,
		}}
	}
	attr := func(name string, mods any) map[string]any {
		typeObj := map[string]any{"computed_optional_required": "computed_optional"}
		if mods != nil {
			typeObj["plan_modifiers"] = mods
		}
		return map[string]any{"name": name, "int64": typeObj}
	}
	replaceOnly := attr("type_id", []any{mod("int64planmodifier.RequiresReplace()")})
	both := attr("both", []any{mod("int64planmodifier.UseStateForUnknown()"), mod("int64planmodifier.RequiresReplace()")})
	optOut := attr("recomputed", []any{})
	spec := []any{map[string]any{"name": "r", "schema": map[string]any{
		"attributes": []any{replaceOnly, both, optOut},
	}}}

	require.NoError(t, applyUseStateForUnknownDefault(spec))

	defs := func(a map[string]any) []string {
		typeObj, ok := a["int64"].(map[string]any)
		require.True(t, ok)
		mods, ok := typeObj["plan_modifiers"].([]any)
		require.True(t, ok)
		var out []string
		for _, m := range mods {
			mm, ok := m.(map[string]any)
			require.True(t, ok)
			c, ok := mm["custom"].(map[string]any)
			require.True(t, ok)
			def, ok := c["schema_definition"].(string)
			require.True(t, ok)
			out = append(out, def)
		}
		return out
	}
	assert.Equal(t, []string{"int64planmodifier.UseStateForUnknown()", "int64planmodifier.RequiresReplace()"}, defs(replaceOnly),
		"UseStateForUnknown is prepended so it runs before RequiresReplace")
	assert.Equal(t, []string{"int64planmodifier.UseStateForUnknown()", "int64planmodifier.RequiresReplace()"}, defs(both),
		"an attribute that already keeps state is not given a second one")
	assert.Empty(t, defs(optOut), "an explicit [] still opts out")
}
