package crud

import (
	"fmt"
	"testing"
)

// noValueTypes is a Source with no tfplugingen Value types at all, which is the
// enforcement/budget case: their payload objects and arrays are modeled as flat
// scalars, not nested attributes.
type noValueTypes struct{}

func (noValueTypes) ClientMethods(string) ([]ClientMethod, error) { return nil, nil }
func (noValueTypes) Structs(string) (map[string]Struct, error)    { return nil, nil }
func (noValueTypes) MarkerImpls(string) (map[string][]string, error) {
	return nil, nil
}
func (noValueTypes) ModelFields(_, modelType string) ([]ModelField, error) {
	return nil, fmt.Errorf("no such type %q", modelType)
}

// enforcementIndex mirrors the enforcement records: a cloud rule and a service
// carried as whole objects, each with an id.
func enforcementIndex() sdkIndex {
	return sdkIndex{structs: map[string]Struct{
		"CloudRule": {Name: "CloudRule", Fields: []Field{
			{GoName: "ID", JSONName: "id", Type: "OptUint64"},
			{GoName: "Name", JSONName: "name", Type: "OptString"},
		}},
		"Service": {Name: "Service", Fields: []Field{
			{GoName: "ID", JSONName: "id", Type: "OptUint64"},
		}},
	}}
}

// A record that carries the cloud rule as `cloud_rule` must still reach the
// model's `cloud_rule_id`. OUEnforcement and FundingSourceEnforcmeent name it
// that way while ProjectEnforcement names the identical field `cloud_rule_id`,
// and the two that did not match dropped the attribute entirely on read (#70).
func TestResolveNestedFlatten_objectNameToIDAttribute(t *testing.T) {
	fields := []Field{
		{GoName: "CloudRule", JSONName: "cloud_rule", Type: "OptCloudRule"},
		{GoName: "Service", JSONName: "service", Type: "OptService"},
	}
	byTF := map[string]ModelField{
		"cloud_rule_id": {GoName: "CloudRuleId", TFSDK: "cloud_rule_id", Type: "types.Int64"},
		"service_id":    {GoName: "ServiceId", TFSDK: "service_id", Type: "types.Int64"},
	}

	res, err := resolveNestedFlatten(noValueTypes{}, "", fields, byTF, enforcementIndex(), "rec.")
	if err != nil {
		t.Fatalf("resolveNestedFlatten: %v", err)
	}
	got := map[string]objIDProjFlat{}
	for _, p := range res.ObjIDProjs {
		got[p.ModelGo] = p
	}
	for _, want := range []struct{ model, src, expr string }{
		{"CloudRuleId", "rec.CloudRule", "int64(rec.CloudRule.Value.ID.Value)"},
		{"ServiceId", "rec.Service", "int64(rec.Service.Value.ID.Value)"},
	} {
		p, ok := got[want.model]
		if !ok {
			t.Fatalf("no id projection for %s; got %v", want.model, res.ObjIDProjs)
		}
		if p.SrcPath != want.src || p.IDExpr != want.expr || !p.Opt {
			t.Errorf("%s = %+v; want src %q expr %q opt true", want.model, p, want.src, want.expr)
		}
	}
}

// A payload carrying BOTH the object and a flat id of the same name must leave
// the attribute to the ordinary scalar path, or two binds would write it.
func TestResolveNestedFlatten_flatIDWinsOverObject(t *testing.T) {
	fields := []Field{
		{GoName: "CloudRule", JSONName: "cloud_rule", Type: "OptCloudRule"},
		{GoName: "CloudRuleID", JSONName: "cloud_rule_id", Type: "OptUint64"},
	}
	byTF := map[string]ModelField{
		"cloud_rule_id": {GoName: "CloudRuleId", TFSDK: "cloud_rule_id", Type: "types.Int64"},
	}

	res, err := resolveNestedFlatten(noValueTypes{}, "", fields, byTF, enforcementIndex(), "rec.")
	if err != nil {
		t.Fatalf("resolveNestedFlatten: %v", err)
	}
	if len(res.ObjIDProjs) != 0 {
		t.Errorf("object projected over a flat id of the same name: %+v", res.ObjIDProjs)
	}
}

// budgetPayload mirrors the budget read: a config object and the per-month rows
// the single create amount became.
func budgetPayload() ([]Field, sdkIndex) {
	fields := []Field{
		{GoName: "Config", JSONName: "config", Type: "OptBudgetConfig"},
		{GoName: "Data", JSONName: "data", Type: "OptNilBudgetDataArray"},
	}
	idx := sdkIndex{structs: map[string]Struct{
		"OptNilBudgetDataArray": {Name: "OptNilBudgetDataArray", Fields: []Field{
			{GoName: "Value", Type: "[]BudgetData"},
			{GoName: "Null", Type: "bool"},
		}},
		"BudgetData": {Name: "BudgetData", Fields: []Field{
			{GoName: "Amount", JSONName: "amount", Type: "OptFloat64"},
			{GoName: "Datecode", JSONName: "datecode", Type: "OptString"},
		}},
	}}
	return fields, idx
}

func TestResolveSumFlats_budgetAmount(t *testing.T) {
	fields, idx := budgetPayload()
	byTF := map[string]ModelField{
		"amount": {GoName: "Amount", TFSDK: "amount", Type: "types.Int64"},
	}

	got, err := resolveSumFlats(
		[]sumFromDecl{{TF: "amount", From: "data", Field: "amount"}},
		fields, byTF, idx, "v.Data.Value.",
	)
	if err != nil {
		t.Fatalf("resolveSumFlats: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d sums, want 1", len(got))
	}
	want := sumFlat{
		ModelGo: "Amount", Var: "amountSum", SrcPath: "v.Data.Value.Data.Value",
		ElemExpr: "elem.Amount.Value", Round: true,
	}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
}

// A declaration that no longer resolves must fail the generate rather than
// quietly leave the attribute dropped, which is the bug it exists to fix.
func TestResolveSumFlats_rejectsStaleDeclaration(t *testing.T) {
	fields, idx := budgetPayload()
	byTF := map[string]ModelField{
		"amount": {GoName: "Amount", TFSDK: "amount", Type: "types.Int64"},
	}

	for name, decl := range map[string]sumFromDecl{
		"unknown attribute":  {TF: "total", From: "data", Field: "amount"},
		"unknown array":      {TF: "amount", From: "rows", Field: "amount"},
		"unknown elem field": {TF: "amount", From: "data", Field: "total"},
		"not an array":       {TF: "amount", From: "config", Field: "amount"},
	} {
		if _, err := resolveSumFlats([]sumFromDecl{decl}, fields, byTF, idx, "v.Data.Value."); err == nil {
			t.Errorf("%s: resolveSumFlats accepted %+v", name, decl)
		}
	}
}
