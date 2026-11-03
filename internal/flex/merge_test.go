package flex

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
)

func asMap(t *testing.T, v any) map[string]any {
	t.Helper()
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("%v is not an object", v)
	}
	return m
}

func TestObjectAsSDK(t *testing.T) {
	v := types.ObjectValueMust(
		map[string]attr.Type{"dataset_name": types.StringType, "table_name": types.StringType},
		map[string]attr.Value{"dataset_name": types.StringValue("ds"), "table_name": types.StringNull()},
	)
	got := ObjectAsSDK[generated.GCPBigQueryExport](v)
	if got.DatasetName.Value != "ds" || !got.DatasetName.Set {
		t.Errorf("dataset_name = %+v", got.DatasetName)
	}
	if got.TableName.Set {
		t.Errorf("a null attribute must stay unset, got %+v", got.TableName)
	}
	if z := ObjectAsSDK[generated.GCPBigQueryExport](types.ObjectNull(v.AttributeTypes(context.Background()))); z.DatasetName.Set {
		t.Error("a null object converts to the zero value")
	}
}

func mergeOut(t *testing.T, current string, fields []MergeField) map[string]any {
	t.Helper()
	body, diags := MergeJSON(context.Background(), []byte(current), fields)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	var got map[string]any
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("decoding merged body: %v", err)
	}
	return got
}

func TestMergeJSON_OverlaysOntoTheReadRecord(t *testing.T) {
	current := `{"status":200,"data":{"id":7,"account_type_id":1,"aws_payer":{"id":3,"name":"old","org_id":"o-1","key_id":"REDACTED","billing_start_date":202401}}}`
	got := mergeOut(t, current, []MergeField{
		{Value: types.StringValue("new"), To: "aws_payer.name", Kind: MergeString},
		{Value: types.StringValue("2025-03"), To: "aws_payer.billing_start_date", Kind: MergeDatecode},
		{Value: types.BoolValue(true), To: "skip_validation", Kind: MergeBool},
		{Value: types.StringNull(), To: "aws_payer.key_id", Kind: MergeString},
	})
	aws := asMap(t, got["aws_payer"])
	if aws["name"] != "new" {
		t.Errorf("name = %v, want new", aws["name"])
	}
	if aws["billing_start_date"] != json.Number("202503") {
		t.Errorf("billing_start_date = %v, want 202503", aws["billing_start_date"])
	}
	// Fields the plan does not model survive, which is what makes a full-overwrite
	// PUT safe; a null plan value leaves the read value alone.
	if aws["org_id"] != "o-1" || aws["id"] != json.Number("3") || aws["key_id"] != "REDACTED" {
		t.Errorf("unmodelled fields not preserved: %v", aws)
	}
	if got["id"] != json.Number("7") || got["skip_validation"] != true {
		t.Errorf("top level = %v", got)
	}
}

func TestMergeJSON_CreatesMissingObjects(t *testing.T) {
	got := mergeOut(t, `{"data":{"id":1}}`, []MergeField{
		{Value: types.Int64Value(9), To: "gcp_billing_account_update.service_account_id", Kind: MergeInt},
	})
	upd := asMap(t, got["gcp_billing_account_update"])
	if upd["service_account_id"] != json.Number("9") {
		t.Errorf("service_account_id = %v", upd["service_account_id"])
	}
}

func TestMergeJSON_Enum(t *testing.T) {
	values := map[string]any{"none": 0, "cur": 1, "dbrrt": 2}
	got := mergeOut(t, `{"data":{}}`, []MergeField{
		{Value: types.StringValue("DBRRT"), To: "t", Kind: MergeEnum, Enum: values},
	})
	if got["t"] != json.Number("2") {
		t.Errorf("t = %v, want 2", got["t"])
	}

	_, diags := MergeJSON(context.Background(), []byte(`{"data":{}}`), []MergeField{
		{Value: types.StringValue("bogus"), To: "t", Kind: MergeEnum, Enum: values},
	})
	if !diags.HasError() {
		t.Fatal("an unknown enum value must fail rather than be dropped")
	}
}

func TestMergeJSON_NestedSubAttribute(t *testing.T) {
	ctx := context.Background()
	bq := types.ObjectValueMust(
		map[string]attr.Type{"table_name": types.StringType, "dataset_name": types.StringType},
		map[string]attr.Value{"table_name": types.StringValue("t1"), "dataset_name": types.StringNull()},
	)
	outer := types.ObjectValueMust(
		map[string]attr.Type{"name": types.StringType, "big_query_export": bq.Type(ctx)},
		map[string]attr.Value{"name": types.StringValue("n"), "big_query_export": bq},
	)
	got := mergeOut(t, `{"data":{}}`, []MergeField{
		{Value: outer, Sub: []string{"name"}, To: "u.name", Kind: MergeString},
		{Value: outer, Sub: []string{"big_query_export", "table_name"}, To: "u.bq.table_name", Kind: MergeString},
		{Value: outer, Sub: []string{"big_query_export", "dataset_name"}, To: "u.bq.dataset_name", Kind: MergeString},
	})
	u := asMap(t, got["u"])
	if u["name"] != "n" {
		t.Errorf("name = %v", u["name"])
	}
	bqOut := asMap(t, u["bq"])
	if bqOut["table_name"] != "t1" {
		t.Errorf("table_name = %v", bqOut["table_name"])
	}
	if _, ok := bqOut["dataset_name"]; ok {
		t.Error("a null sub-attribute must not be written")
	}
}

func TestMergeJSON_RejectsBadInput(t *testing.T) {
	cases := map[string]struct {
		current string
		field   MergeField
	}{
		"no data envelope":      {`{"status":200}`, MergeField{Value: types.StringValue("x"), To: "a", Kind: MergeString}},
		"bad datecode":          {`{"data":{}}`, MergeField{Value: types.StringValue("March"), To: "a", Kind: MergeDatecode}},
		"missing sub":           {`{"data":{}}`, MergeField{Value: types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}), Sub: []string{"x"}, To: "a", Kind: MergeString}},
		"path through a scalar": {`{"data":{"a":1}}`, MergeField{Value: types.StringValue("x"), To: "a.b", Kind: MergeString}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, diags := MergeJSON(context.Background(), []byte(tc.current), []MergeField{tc.field})
			if !diags.HasError() {
				t.Fatal("expected an error")
			}
		})
	}
}
