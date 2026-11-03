package flex

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ResolveUnknowns nulls every attr.Value field of model still unknown.
//
// An Optional+Computed attribute is unknown at plan time when the config omits
// it, and the read-back after create only assigns what the API returns. A
// write-only input (cloud_rule_id) or a field the API never echoes therefore
// stays unknown into state, and Terraform rejects that with "Provider returned
// invalid result object after apply". Null is the honest answer: the API does
// not have the value. With UseStateForUnknown the null then persists, so later
// plans read it from state and show no diff.
func ResolveUnknowns(ctx context.Context, model any) diag.Diagnostics {
	var diags diag.Diagnostics

	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return diags
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return diags
	}

	valueType := reflect.TypeOf((*attr.Value)(nil)).Elem()

	for i := range v.NumField() {
		f := v.Field(i)
		if !f.CanSet() || !f.Type().Implements(valueType) {
			continue
		}
		cur, ok := f.Interface().(attr.Value)
		if !ok || cur.IsNull() || !cur.IsUnknown() {
			continue
		}

		// Built through the value's own type rather than a types.XNull()
		// switch, so custom types (jsontypes.Normalized) and collections keep
		// their element type instead of decaying to a base type.
		t := cur.Type(ctx)
		null, err := t.ValueFromTerraform(ctx, tftypes.NewValue(t.TerraformType(ctx), nil))
		if err != nil {
			diags.AddError(
				"Resolving unknown attribute",
				"Could not build a null "+t.String()+" for field "+v.Type().Field(i).Name+": "+err.Error(),
			)
			continue
		}
		f.Set(reflect.ValueOf(null))
	}

	return diags
}
