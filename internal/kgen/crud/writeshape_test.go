package crud

import (
	"strings"
	"testing"
)

// A struct inside an already-nested object used to be skipped with a warning,
// so kion_billing_source_gcp's big_query_export never reached the create body.
func TestObjectSubExpr(t *testing.T) {
	got, ok := objectSubExpr(Field{GoName: "BigQueryExport", JSONName: "big_query_export", Type: "GCPBigQueryExport"},
		ModelField{GoName: "BigQueryExport", TFSDK: "big_query_export", Type: "basetypes.ObjectValue"}, "plan.Gcp")
	if !ok || got != "flex.ObjectAsSDK[generated.GCPBigQueryExport](plan.Gcp.BigQueryExport)" {
		t.Fatalf("got %q, %v", got, ok)
	}
	if _, ok := objectSubExpr(Field{Type: "OptGCPBigQueryExport"}, ModelField{Type: "basetypes.ObjectValue"}, "p"); ok {
		t.Error("an Opt-wrapped struct is not handled")
	}
	if _, ok := objectSubExpr(Field{Type: "GCPBigQueryExport"}, ModelField{Type: "basetypes.StringValue"}, "p"); ok {
		t.Error("a non-object model sub is not handled")
	}
}

func writeLikeModel() map[string]ModelField {
	return map[string]ModelField{
		"id":                   {GoName: "Id", TFSDK: "id", Type: "types.String"},
		"name":                 {GoName: "Name", TFSDK: "name", Type: "types.String"},
		"billing_report_type":  {GoName: "BillingReportType", TFSDK: "billing_report_type", Type: "types.String"},
		"gcp_billing_account":  {GoName: "GcpBillingAccount", TFSDK: "gcp_billing_account", Type: "GcpBillingAccountValue"},
		"skip_validation":      {GoName: "SkipValidation", TFSDK: "skip_validation", Type: "types.Bool"},
		"billing_start_date":   {GoName: "BillingStartDate", TFSDK: "billing_start_date", Type: "types.String"},
		"service_account_id":   {GoName: "ServiceAccountId", TFSDK: "service_account_id", Type: "types.Int64"},
		"unrelated_attribute":  {GoName: "UnrelatedAttribute", TFSDK: "unrelated_attribute", Type: "types.String"},
		"another_unrelated_id": {GoName: "AnotherUnrelatedId", TFSDK: "another_unrelated_id", Type: "types.Int64"},
	}
}

func TestBuildMergeFields(t *testing.T) {
	got, err := buildMergeFields([]writeShapeField{
		{TF: "name", To: "aws_payer.name", Kind: "string"},
		{TF: "billing_start_date", To: "aws_payer.billing_start_date", Kind: "datecode"},
		{TF: "skip_validation", To: "skip_validation", Kind: "bool"},
		{TF: "gcp_billing_account.big_query_export.table_name", To: "u.big_query_export.table_name", Kind: "string"},
		{TF: "billing_report_type", To: "aws_payer.billing_report_type_id", Kind: "enum", Values: map[string]any{"none": 0, "cur": 1}},
		{TF: "billing_report_type", To: "use_proprietary_reports", Kind: "enum", Values: map[string]any{"none": false, "cur": true}},
	}, writeLikeModel())
	if err != nil {
		t.Fatalf("buildMergeFields: %v", err)
	}
	for _, want := range []string{
		`{Value: plan.Name, To: "aws_payer.name", Kind: flex.MergeString},`,
		`{Value: plan.BillingStartDate, To: "aws_payer.billing_start_date", Kind: flex.MergeDatecode},`,
		`{Value: plan.SkipValidation, To: "skip_validation", Kind: flex.MergeBool},`,
		`{Value: plan.GcpBillingAccount, Sub: []string{"big_query_export", "table_name"}, To: "u.big_query_export.table_name", Kind: flex.MergeString},`,
		// Keys sorted so regeneration is stable.
		`Kind: flex.MergeEnum, Enum: map[string]any{"cur": 1, "none": 0}},`,
		`Kind: flex.MergeEnum, Enum: map[string]any{"cur": true, "none": false}},`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s\n%s", want, got)
		}
	}
}

func TestBuildMergeFields_rejects(t *testing.T) {
	cases := map[string]writeShapeField{
		"attribute not in model": {TF: "nope", To: "a", Kind: "string"},
		"unknown kind":           {TF: "name", To: "a", Kind: "float"},
		"enum without values":    {TF: "name", To: "a", Kind: "enum"},
		"no target":              {TF: "name", Kind: "string"},
		"the id":                 {TF: "id", To: "id", Kind: "string"},
		"sub of a scalar":        {TF: "name.x", To: "a", Kind: "string"},
	}
	for name, f := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := buildMergeFields([]writeShapeField{f}, writeLikeModel()); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
	if _, err := buildMergeFields([]writeShapeField{
		{TF: "name", To: "a", Kind: "string"},
		{TF: "skip_validation", To: "a", Kind: "bool"},
	}, writeLikeModel()); err == nil {
		t.Fatal("two fields writing one path must fail")
	}
}

// A read_shape kind that flattens to a different framework type than the model
// attribute fails at runtime on every read; it has to fail generation instead.
func TestCheckReadShapeKinds(t *testing.T) {
	shape := readShape{
		Scalars: []readShapeSub{{TF: "name", From: "x.name", Kind: "string"}},
		Objects: []readShapeObject{{
			TF: "gcp_billing_account", ValueType: "GcpBillingAccountValue", From: "x.ba",
			Subs: []readShapeSub{
				{TF: "billing_start_date", From: "billing_start_date", Kind: "datecode"},
				{TF: "is_reseller", From: "is_reseller", Kind: "bool"},
				{TF: "secret", From: "secret", Kind: "int", WriteOnly: true},
			},
		}},
	}
	subs := map[string]map[string]string{
		"GcpBillingAccountValue": {"billing_start_date": "basetypes.StringValue", "is_reseller": "basetypes.BoolValue", "secret": "basetypes.StringValue"},
	}
	if err := checkReadShapeKinds(shape, writeLikeModel(), subs); err != nil {
		t.Fatalf("valid shape rejected: %v", err)
	}

	shape.Objects[0].Subs[0].Kind = "int"
	err := checkReadShapeKinds(shape, writeLikeModel(), subs)
	if err == nil || !strings.Contains(err.Error(), "billing_start_date") {
		t.Fatalf("int into a string sub-attribute must fail, got %v", err)
	}

	bad := readShape{Scalars: []readShapeSub{{TF: "name", From: "x", Kind: "bool"}}}
	if err := checkReadShapeKinds(bad, writeLikeModel(), subs); err == nil {
		t.Fatal("bool into a string attribute must fail")
	}
}
