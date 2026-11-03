package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// OrderLikePrior returns read with its object elements in prior's order,
// matched on the key attribute; elements prior lacks follow in read's order.
// It is for a list the API returns in an order of its own (session tags come
// back in creation order), where taking that order verbatim fails the apply
// with "Provider produced inconsistent result". A null or unknown prior, or an
// element without the key, leaves read as it is.
func OrderLikePrior(ctx context.Context, prior, read types.List, key string) (types.List, diag.Diagnostics) {
	if prior.IsNull() || prior.IsUnknown() || read.IsNull() || read.IsUnknown() {
		return read, nil
	}
	var diags diag.Diagnostics
	keyOf := func(v attr.Value) (string, bool) {
		ov, ok := v.(basetypes.ObjectValuable)
		if !ok {
			return "", false
		}
		obj, d := ov.ToObjectValue(ctx)
		diags.Append(d...)
		k, ok := obj.Attributes()[key]
		if !ok || k.IsNull() || k.IsUnknown() {
			return "", false
		}
		return k.String(), true
	}

	elems := read.Elements()
	used := make([]bool, len(elems))
	ordered := make([]attr.Value, 0, len(elems))
	for _, p := range prior.Elements() {
		pk, ok := keyOf(p)
		if !ok {
			continue
		}
		for i, e := range elems {
			if ek, ok := keyOf(e); !used[i] && ok && ek == pk {
				used[i] = true
				ordered = append(ordered, e)
				break
			}
		}
	}
	for i, e := range elems {
		if !used[i] {
			ordered = append(ordered, e)
		}
	}
	if diags.HasError() {
		return read, diags
	}
	out, d := types.ListValue(read.ElementType(ctx), ordered)
	diags.Append(d...)
	return out, diags
}
