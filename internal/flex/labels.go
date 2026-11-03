package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
)

// Labels are a map in the provider and a list of {key, value} on the wire, and
// they are not part of any resource's own request body: Kion carries them on a
// per-resource sub-resource (PUT /v3/<type>/{id}/labels). So a resource that
// exposes `labels` has to sync them separately, and until it does the attribute
// is accepted and discarded.

// AssociateLabelsFromFramework converts a labels map to the wire array.
// A null or unknown map yields nil, which the caller skips rather than sending
// as an empty array -- an empty array clears every label on the record.
func AssociateLabelsFromFramework(ctx context.Context, v types.Map) ([]generated.AssociateLabel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v.IsNull() || v.IsUnknown() {
		return nil, diags
	}

	elems := make(map[string]types.String, len(v.Elements()))
	diags.Append(v.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil, diags
	}

	// Non-nil even when empty: the caller distinguishes "no labels configured"
	// (nil, skip the call) from "explicitly none" (empty, clear them).
	out := make([]generated.AssociateLabel, 0, len(elems))
	for k, val := range elems {
		out = append(out, generated.AssociateLabel{
			Key:   generated.NewOptString(k),
			Value: generated.NewOptString(val.ValueString()),
		})
	}
	return out, diags
}

// LabelsToFramework converts per-resource label records to a labels map, given
// an accessor for the key and value of each. Kion gives every resource its own
// record type (GetOULabel, GetAccountLabel, …) differing only in which parent
// id it carries, so this is generic over the two fields that matter rather than
// repeated once per resource.
//
// Kion returns an unset label list as an empty array, so an absent value maps
// to an empty map rather than null: null would differ from the empty map a
// configuration can legitimately hold, and diff for ever.
func LabelsToFramework[T any](ctx context.Context, in []T, kv func(T) (string, string)) (types.Map, diag.Diagnostics) {
	elems := make(map[string]types.String, len(in))
	for _, l := range in {
		k, v := kv(l)
		if k == "" {
			continue
		}
		elems[k] = types.StringValue(v)
	}
	return types.MapValueFrom(ctx, types.StringType, elems)
}
