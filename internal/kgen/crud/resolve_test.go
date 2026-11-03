package crud

import (
	"errors"
	"maps"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureIndex builds an sdkIndex from the Label testdata fixtures via the real
// fileSource, exercising the parsers end-to-end, not a mock.
func fixtureIndex(t *testing.T) sdkIndex {
	t.Helper()
	return fixtureIndexFrom(t, "testdata")
}

// fixtureIndexFrom builds an sdkIndex from the oas_*_gen.go fixtures in dir.
func fixtureIndexFrom(t *testing.T, dir string) sdkIndex {
	t.Helper()
	src := NewFileSource()
	cms, err := src.ClientMethods(filepath.Join(dir, "oas_client_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	methods := map[string]ClientMethod{}
	for _, m := range cms {
		methods[m.HTTPMethod+" "+m.Path] = m
	}
	schemas, err := src.Structs(filepath.Join(dir, "oas_schemas_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	params, err := src.Structs(filepath.Join(dir, "oas_parameters_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	structs := maps.Clone(schemas)
	maps.Copy(structs, params)
	markers, err := src.MarkerImpls(filepath.Join(dir, "oas_schemas_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	return sdkIndex{methods: methods, structs: structs, markerImpls: markers}
}

func labelOps() resOps {
	return resOps{
		Create: &opRef{Method: "POST", Path: "/v3/label"},
		Read:   &opRef{Method: "GET", Path: "/v3/label/{id}"},
		Update: &opRef{Method: "PATCH", Path: "/v3/label/{id}"},
		Delete: &opRef{Method: "DELETE", Path: "/v3/label/{id}"},
	}
}

func labelDS() dsOps {
	return dsOps{Read: &opRef{Method: "GET", Path: "/v3/label"}}
}

func TestResolveResource_label(t *testing.T) {
	idx := fixtureIndex(t)
	model, err := NewFileSource().ModelFields("testdata/label_schema_gen.go", "LabelModel")
	if err != nil {
		t.Fatal(err)
	}

	rm, err := resolveResource("label", labelOps(), labelDS(), idx, model, false)
	if err != nil {
		t.Fatalf("resolveResource: %v", err)
	}

	if rm.List == nil || rm.List.Method.Name != "GetLabelIndex" {
		t.Errorf("rm.List = %+v (want resolved GetLabelIndex)", rm.List)
	}
	if rm.Pascal != "Label" || rm.Model != "LabelModel" {
		t.Errorf("rm pascal/model = %q/%q", rm.Pascal, rm.Model)
	}
	if rm.Create.Method.Name != "PostLabel" || rm.Create.Body == nil || rm.Create.Body.Name != "CreateLabel" {
		t.Errorf("create = %+v body=%+v", rm.Create.Method, rm.Create.Body)
	}
	if rm.Read.Method.Name != "GetLabel" || rm.Read.Params == nil || rm.Read.Params.Name != "GetLabelParams" {
		t.Errorf("read = %+v params=%+v", rm.Read.Method, rm.Read.Params)
	}
	if rm.Read.RespType != "LabelResponse" || len(rm.Read.RespFields) != 4 {
		t.Errorf("read envelope = %q fields=%d", rm.Read.RespType, len(rm.Read.RespFields))
	}
	if rm.Update == nil || rm.Update.Method.Name != "PatchLabel" {
		t.Errorf("update = %+v", rm.Update)
	}
	if rm.Delete == nil || rm.Delete.Method.Name != "DeleteLabel" {
		t.Errorf("delete = %+v", rm.Delete)
	}
	if rm.IDField.GoName != "Id" || len(rm.Fields) != 3 {
		t.Errorf("id=%+v fields=%d (want Id + 3 non-id)", rm.IDField, len(rm.Fields))
	}
}

func TestResolveOp_privateEndpointRefused(t *testing.T) {
	idx := fixtureIndex(t)
	_, err := resolveOp("read", &opRef{Method: "GET", Path: "/v1/dashboard/{id}"}, idx)
	var target noSDKOpError
	if !errors.As(err, &target) {
		t.Fatalf("want noSDKOpError for a route absent from the SDK, got %v", err)
	}
}

// labelRead is the resolved single-read op the list resolution is matched
// against (its payload type must equal the list element).
func labelRead(t *testing.T, idx sdkIndex) OpModel {
	t.Helper()
	op, err := resolveOp("read", &opRef{Method: "GET", Path: "/v3/label/{id}"}, idx)
	if err != nil {
		t.Fatalf("resolveOp read: %v", err)
	}
	return *op
}

// Shape A: Data wraps a struct that holds the items slice plus Total, and the
// op paginates.
func TestResolveList_shapeA_paginatedEnvelope(t *testing.T) {
	idx := fixtureIndex(t)
	lm, err := resolveList(&opRef{Method: "GET", Path: "/v3/label"}, idx, labelRead(t, idx))
	if err != nil {
		t.Fatalf("resolveList: %v", err)
	}
	if lm.Method.Name != "GetLabelIndex" {
		t.Errorf("method = %s", lm.Method.Name)
	}
	if lm.RespType != "LabelListPaginatedResponse" || lm.DataInner != "LabelListPaginated" {
		t.Errorf("resp=%s inner=%s", lm.RespType, lm.DataInner)
	}
	if lm.DataDirect || !lm.DataOpt || lm.ElemIsWrapper {
		t.Errorf("direct=%v opt=%v elemIsWrapper=%v (want false/true/false)", lm.DataDirect, lm.DataOpt, lm.ElemIsWrapper)
	}
	if lm.ItemsGo != "Items" || lm.ElemType != "Label" || !lm.ItemsNil || !lm.ItemsOpt {
		t.Errorf("items=%s elem=%s nil=%v opt=%v", lm.ItemsGo, lm.ElemType, lm.ItemsNil, lm.ItemsOpt)
	}
	if lm.TotalGo != "Total" || !lm.Paginated || lm.PageParam != "Page" || lm.CountParam != "Count" {
		t.Errorf("total=%s paginated=%v page=%s count=%s", lm.TotalGo, lm.Paginated, lm.PageParam, lm.CountParam)
	}
	a := lm.access()
	if a.EnvelopeGuard != "resp.Data.Set" || a.ItemsExpr != "resp.Data.Value.Items" || a.TotalExpr != "resp.Data.Value.Total" {
		t.Errorf("access = %+v", a)
	}
	if a.ItemsGuard != "items.Set && !items.Null" || a.ItemsSlice != "items.Value" {
		t.Errorf("access items = %+v", a)
	}
}

// Shape B: Data IS the items slice, and the op takes no pagination params. This
// is the regression: it used to abort list resolution entirely (the inner type
// was looked up as the literal string "[]Label"), silently downgrading the data
// source to id-only.
func TestResolveList_shapeB_directSliceNotPaginated(t *testing.T) {
	idx := fixtureIndex(t)
	lm, err := resolveList(&opRef{Method: "GET", Path: "/v3/label-flat"}, idx, labelRead(t, idx))
	if err != nil {
		t.Fatalf("resolveList: %v", err)
	}
	if !lm.DataDirect || lm.DataInner != "" {
		t.Errorf("direct=%v inner=%q (want true/empty)", lm.DataDirect, lm.DataInner)
	}
	if lm.ItemsGo != "Data" || lm.ElemType != "Label" || lm.ItemsOpt || lm.ItemsNil {
		t.Errorf("items=%s elem=%s opt=%v nil=%v", lm.ItemsGo, lm.ElemType, lm.ItemsOpt, lm.ItemsNil)
	}
	if lm.Paginated || lm.PageParam != "" || lm.TotalGo != "" {
		t.Errorf("paginated=%v page=%q total=%q (endpoint has neither)", lm.Paginated, lm.PageParam, lm.TotalGo)
	}
	a := lm.access()
	if a.EnvelopeGuard != "" || a.ItemsExpr != "resp.Data" || a.ItemsSlice != "items" || a.ItemsGuard != "" {
		t.Errorf("access = %+v", a)
	}
	if a.ItemsBreak != "len(items) == 0" || a.TotalExpr != "" || a.HasParams {
		t.Errorf("access = %+v", a)
	}
}

// A collection op whose params carry a required value the data source cannot
// supply (the parent id of a nested collection) is refused rather than called
// with a zero id.
func TestResolveList_requiredParamRefused(t *testing.T) {
	idx := fixtureIndex(t)
	_, err := resolveList(&opRef{Method: "GET", Path: "/v3/label/{id}/child"}, idx, labelRead(t, idx))
	if err == nil {
		t.Fatal("want refusal for a list op with a required param")
	}
	if !strings.Contains(err.Error(), "requires param") {
		t.Errorf("error should name the unsupplyable param, got %v", err)
	}
}

// An items slice of some unrelated type is not the collection for this resource.
func TestResolveList_elementTypeMismatchRefused(t *testing.T) {
	idx := fixtureIndex(t)
	_, err := resolveList(&opRef{Method: "GET", Path: "/v3/label-foreign"}, idx, labelRead(t, idx))
	if err == nil {
		t.Fatal("want refusal when the list element is not a type the read yields")
	}
}

func TestResolveList_noListOpRefused(t *testing.T) {
	idx := fixtureIndex(t)
	if _, err := resolveList(&opRef{Method: "GET", Path: "/v3/nope"}, idx, labelRead(t, idx)); err == nil {
		t.Fatal("want error for unresolved list op")
	}
}

// nestedRecordIndex models the azure_policy shape: the read payload
// (AzurePolicyAugmented) has NO top-level id and nests the record under the
// azure_policy field, which the generated model exposes as a nested object
// attribute (AzurePolicyValue).
func nestedRecordIndex() (OpModel, map[string]ModelField, sdkIndex) {
	idx := sdkIndex{structs: map[string]Struct{
		"AzurePolicyDefinition": {Name: "AzurePolicyDefinition", Fields: []Field{
			{GoName: "ID", JSONName: "id", Type: "OptUint64", Optional: true},
			{GoName: "Name", JSONName: "name", Type: "OptString", Optional: true},
		}},
		"OptNilUserArray": {Name: "OptNilUserArray", Fields: []Field{
			{GoName: "Value", Type: "[]User"}, {GoName: "Set", Type: "bool"}, {GoName: "Null", Type: "bool"},
		}},
	}}
	read := OpModel{Verb: "read", RespPayload: "AzurePolicyAugmented", RespFields: []Field{
		{GoName: "AzurePolicy", JSONName: "azure_policy", Type: "OptAzurePolicyDefinition", Optional: true},
		{GoName: "OwnerUsers", JSONName: "owner_users", Type: "OptNilUserArray", Optional: true},
	}}
	byTF := map[string]ModelField{
		"id":           {GoName: "Id", TFSDK: "id", Type: "types.String"},
		"azure_policy": {GoName: "AzurePolicy", TFSDK: "azure_policy", Type: "AzurePolicyValue"},
		"owner_users":  {GoName: "OwnerUsers", TFSDK: "owner_users", Type: "types.List"},
	}
	return read, byTF, idx
}

func TestDetectRecordWrapper_nestedRecordUnderModelObjectAttr(t *testing.T) {
	read, byTF, idx := nestedRecordIndex()
	var fields []ModelField
	for _, mf := range byTF {
		if mf.TFSDK != "id" {
			fields = append(fields, mf)
		}
	}
	detectRecordWrapper(&read, byTF["id"], fields, idx)
	if read.RespWrapperGo != "AzurePolicy" {
		t.Fatalf("RespWrapperGo = %q, want AzurePolicy", read.RespWrapperGo)
	}
	if !read.RespWrapperOpt {
		t.Error("RespWrapperOpt = false, want true for an Opt-wrapped record")
	}
	// The wrapper's fields must be visible so the record view can reach the id.
	view := buildRecordView(ResourceModel{Read: read})
	if _, ok := view.Fields["id"]; !ok {
		t.Errorf("record view has no id field; paths = %v", view.Paths)
	}
	if got := view.Paths["id"]; got != "AzurePolicy.Value.ID" {
		t.Errorf("id path = %q, want AzurePolicy.Value.ID", got)
	}
}

func TestNestedRecordWrapper_skippedWhenPayloadHasTopLevelID(t *testing.T) {
	read, byTF, idx := nestedRecordIndex()
	read.RespFields = append(read.RespFields, Field{GoName: "ID", JSONName: "id", Type: "OptUint64", Optional: true})
	if _, _, _, ok := nestedRecordWrapper(&read, byTF, idx); ok {
		t.Error("must not descend into a sub-object when the payload already yields the id")
	}
}

func TestNestedRecordWrapper_skippedWhenAmbiguous(t *testing.T) {
	read, byTF, idx := nestedRecordIndex()
	idx.structs["OtherDefinition"] = Struct{Name: "OtherDefinition", Fields: []Field{
		{GoName: "ID", JSONName: "id", Type: "OptUint64", Optional: true},
	}}
	read.RespFields = append(read.RespFields, Field{GoName: "Other", JSONName: "other", Type: "OptOtherDefinition", Optional: true})
	byTF["other"] = ModelField{GoName: "Other", TFSDK: "other", Type: "OtherValue"}
	if _, _, _, ok := nestedRecordWrapper(&read, byTF, idx); ok {
		t.Error("two candidate records must be refused rather than guessed at")
	}
}

func TestNestedRecordWrapper_skippedWhenNotAModelObjectAttr(t *testing.T) {
	read, byTF, idx := nestedRecordIndex()
	// A plain scalar model attribute over the same SDK struct is not a record
	// wrapper. The record would have nowhere to land.
	byTF["azure_policy"] = ModelField{GoName: "AzurePolicy", TFSDK: "azure_policy", Type: "types.String"}
	if _, _, _, ok := nestedRecordWrapper(&read, byTF, idx); ok {
		t.Error("a non-object model attribute must not be treated as a record wrapper")
	}
}

func TestLoadConfig_skipsFlaggedAndParsesOps(t *testing.T) {
	resources, dataSources, flagged, err := loadConfig("testdata/generator_config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := resources["account"]; ok {
		t.Error("commented-out INCOMPLETE resource must not be loaded")
	}
	if !flagged["guessed"] {
		t.Error("heuristic-annotated resource must be flagged")
	}
	lbl := resources["label"]
	if lbl.Create == nil || lbl.Create.Method != "POST" || lbl.Create.Path != "/v3/label" {
		t.Errorf("label create = %+v", lbl.Create)
	}
	if lbl.Delete == nil || lbl.Delete.Method != "DELETE" {
		t.Errorf("label delete = %+v", lbl.Delete)
	}
	if ds := dataSources["label"]; ds.Read == nil || ds.Read.Path != "/v3/label" {
		t.Errorf("label ds read = %+v", ds.Read)
	}
}
