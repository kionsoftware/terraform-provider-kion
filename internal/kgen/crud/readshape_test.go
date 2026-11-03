package crud

import (
	"strings"
	"testing"
)

// billingLikeShape mirrors the real custom-billing-source declaration: two
// nested objects under one private read, one of which carries a write-only
// secret. It is the shape that motivated both features, so the tests below
// exercise them together rather than in isolation.
func billingLikeShape() readShape {
	return readShape{
		Scalars: []readShapeSub{
			{TF: "id", From: "custom_billing_source.id", Kind: "id"},
			{TF: "name", From: "custom_billing_source.name", Kind: "string"},
		},
		Objects: []readShapeObject{
			{
				TF: "aws_connection", ValueType: "AwsConnectionValue",
				From: "custom_billing_source.aws_connection",
				Subs: []readShapeSub{
					{TF: "account_number", From: "account_number", Kind: "string"},
					{TF: "bucket_access_role", From: "role_name", Kind: "string"},
				},
			},
			{
				TF: "azure_connection", ValueType: "AzureConnectionValue",
				From: "custom_billing_source.azure_connection",
				Subs: []readShapeSub{
					{TF: "storage_container", From: "storage_container", Kind: "string"},
					{TF: "tenant_client_secret", From: "tenant_client_secret", Kind: "string", WriteOnly: true},
				},
			},
		},
	}
}

func billingLikeModel() map[string]ModelField {
	return map[string]ModelField{
		"id":               {GoName: "Id", TFSDK: "id", Type: "types.String"},
		"name":             {GoName: "Name", TFSDK: "name", Type: "types.String"},
		"aws_connection":   {GoName: "AwsConnection", TFSDK: "aws_connection", Type: "AwsConnectionValue"},
		"azure_connection": {GoName: "AzureConnection", TFSDK: "azure_connection", Type: "AzureConnectionValue"},
	}
}

// A second object must reach the wire struct. Before Objects was a slice only
// the first could, which is what blocked azure_connection.
func TestBuildWireStruct_multipleObjects(t *testing.T) {
	got, err := buildWireStruct("billing_source", billingLikeShape())
	if err != nil {
		t.Fatalf("buildWireStruct: %v", err)
	}
	for _, want := range []string{
		`json:"aws_connection"`,
		`json:"azure_connection"`,
		`json:"account_number"`,
		`json:"storage_container"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("wire struct missing %s\n%s", want, got)
		}
	}
}

// A write-only sub is never returned truthfully (the payer read sanitizes it to
// "REDACTED"), so it must not appear in the wire struct at all.
func TestBuildWireStruct_omitsWriteOnly(t *testing.T) {
	got, err := buildWireStruct("billing_source", billingLikeShape())
	if err != nil {
		t.Fatalf("buildWireStruct: %v", err)
	}
	if strings.Contains(got, "tenant_client_secret") {
		t.Errorf("write-only sub leaked into the wire struct:\n%s", got)
	}
}

// Flatten must still supply every sub-attribute. The Value constructor requires
// the full set. But a write-only one carries the prior model value rather than
// a wire value, so state keeps what was configured.
func TestBuildNestedFlatten_writeOnlyCarriesPriorState(t *testing.T) {
	got, err := buildNestedFlatten(billingLikeShape(), billingLikeModel())
	if err != nil {
		t.Fatalf("buildNestedFlatten: %v", err)
	}
	if !strings.Contains(got, `"tenant_client_secret": flex.CarryOrNull(ctx, m.AzureConnection.TenantClientSecret, AzureConnectionValue{}.AttributeTypes(ctx)["tenant_client_secret"]),`) {
		t.Errorf("write-only sub should carry the prior model value\n%s", got)
	}
	if strings.Contains(got, "w.Data.CustomBillingSource.AzureConnection.TenantClientSecret") {
		t.Errorf("write-only sub must not be sourced from the wire\n%s", got)
	}
	// The non-write-only siblings still come from the wire.
	if !strings.Contains(got, "w.Data.CustomBillingSource.AzureConnection.StorageContainer") {
		t.Errorf("ordinary sub should be sourced from the wire\n%s", got)
	}
}

// Both objects must be flattened, not just the first.
func TestBuildNestedFlatten_multipleObjects(t *testing.T) {
	got, err := buildNestedFlatten(billingLikeShape(), billingLikeModel())
	if err != nil {
		t.Fatalf("buildNestedFlatten: %v", err)
	}
	for _, want := range []string{
		"m.AwsConnection = AwsConnectionVal",
		"m.AzureConnection = AzureConnectionVal",
		// the aws sub renames role_name -> bucket_access_role
		"w.Data.CustomBillingSource.AwsConnection.RoleName",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("flatten missing %q\n%s", want, got)
		}
	}
}

// A scalar can be write-only too, in which case it is simply not assigned and
// the model keeps its prior value.
func TestBuildNestedFlatten_writeOnlyScalarSkipped(t *testing.T) {
	s := readShape{Scalars: []readShapeSub{
		{TF: "id", From: "custom_billing_source.id", Kind: "id"},
		{TF: "name", From: "custom_billing_source.name", Kind: "string", WriteOnly: true},
	}}
	got, err := buildNestedFlatten(s, billingLikeModel())
	if err != nil {
		t.Fatalf("buildNestedFlatten: %v", err)
	}
	if strings.Contains(got, "m.Name =") {
		t.Errorf("write-only scalar should not be assigned\n%s", got)
	}
	if !strings.Contains(got, "m.Id =") {
		t.Errorf("ordinary scalar should still be assigned\n%s", got)
	}
}

// An object naming a model attribute that does not exist is a config error, not
// a silently dropped attribute.
func TestBuildNestedFlatten_unknownObjectErrors(t *testing.T) {
	s := readShape{Objects: []readShapeObject{{
		TF: "nope", ValueType: "NopeValue", From: "custom_billing_source.nope",
		Subs: []readShapeSub{{TF: "x", From: "x", Kind: "string"}},
	}}}
	if _, err := buildNestedFlatten(s, billingLikeModel()); err == nil {
		t.Fatal("expected an error for an object not present in the model")
	}
}

// A sub declared kind: object reads its own leaves one level deeper, as GCP's
// big_query_export does inside gcp_billing_account.billing_account. Carrying it
// from prior state left a required attribute empty on import.
func TestReadShape_nestedObjectSub(t *testing.T) {
	s := readShape{Objects: []readShapeObject{{
		TF: "gcp_billing_account_create", ValueType: "GcpBillingAccountCreateValue",
		From: "gcp_billing_account.billing_account",
		Subs: []readShapeSub{
			{TF: "name", From: "name", Kind: "string"},
			{TF: "big_query_export", From: "big_query_export", Kind: "object", Subs: []readShapeSub{
				{TF: "dataset_name", From: "dataset_name", Kind: "string"},
			}},
		},
	}}}
	model := map[string]ModelField{"gcp_billing_account_create": {GoName: "GcpBillingAccountCreate"}}

	got, err := buildNestedFlatten(s, model)
	if err != nil {
		t.Fatalf("buildNestedFlatten: %v", err)
	}
	want := `"big_query_export": flex.ObjectFromAttrs(ctx, &diags, GcpBillingAccountCreateValue{}.AttributeTypes(ctx)["big_query_export"], map[string]attr.Value{
"dataset_name": types.StringValue(w.Data.GcpBillingAccount.BillingAccount.BigQueryExport.DatasetName),
}),`
	if !strings.Contains(got, want) {
		t.Errorf("nested object sub not flattened from the wire\n%s", got)
	}

	tree, err := buildWireTree(s)
	if err != nil {
		t.Fatalf("buildWireTree: %v", err)
	}
	if !strings.Contains(emitStruct(tree), "DatasetName string `json:\"dataset_name\"`") {
		t.Errorf("nested leaf missing from the wire struct\n%s", emitStruct(tree))
	}
}

// A plain JSON null (Go *uint / *string on the API side) must read as a null
// attribute: an int64 wire field turns it into 0, which diffs against an unset
// config for ever. Azure's azure_csp_id is the case.
func TestReadShape_optKinds(t *testing.T) {
	for kind, want := range map[string][2]string{
		"opt_int":    {"*int64", "types.Int64PointerValue(x)"},
		"opt_string": {"*string", "types.StringPointerValue(x)"},
	} {
		w, err := kindWire(kind)
		if err != nil || w != want[0] {
			t.Errorf("kindWire(%s) = %q, %v; want %q", kind, w, err, want[0])
		}
		c, err := kindConv(kind, "x")
		if err != nil || c != want[1] {
			t.Errorf("kindConv(%s) = %q, %v; want %q", kind, c, err, want[1])
		}
	}
}
