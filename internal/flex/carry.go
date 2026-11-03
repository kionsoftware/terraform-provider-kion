package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// CarryOrNull returns prior when it already has type t, else a null of type t.
// A read that carries a write-only value from prior state has none on import,
// where the zero types.Object has no attribute types and an object built from
// it is rejected as an invalid attribute type.
func CarryOrNull(ctx context.Context, prior attr.Value, t attr.Type) attr.Value {
	if prior != nil && prior.Type(ctx).Equal(t) {
		return prior
	}
	null, err := t.ValueFromTerraform(ctx, tftypes.NewValue(t.TerraformType(ctx), nil))
	if err != nil {
		return prior
	}
	return null
}

// ObjectFromAttrs builds an object of type t (a types.ObjectType) from vals,
// appending any construction error to diags; a read shape uses it for an
// object nested inside another object's attribute map.
func ObjectFromAttrs(ctx context.Context, diags *diag.Diagnostics, t attr.Type, vals map[string]attr.Value) attr.Value {
	ot, ok := t.(basetypes.ObjectType)
	if !ok {
		diags.AddError("building nested object", "attribute type is not an object type")
		return CarryOrNull(ctx, nil, t)
	}
	v, d := types.ObjectValue(ot.AttrTypes, vals)
	diags.Append(d...)
	return v
}
