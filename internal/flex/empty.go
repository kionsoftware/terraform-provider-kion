package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SetEmptyLikePrior returns an empty set when read is null and prior is a known
// empty set, and read otherwise. Kion returns an empty membership list as null,
// so a configured `[]` read back as null and failed the apply with "was
// cty.SetValEmpty, but now null". Unlike the empty_collections converters it
// leaves an unset attribute null, so it is safe on every optional set.
func SetEmptyLikePrior(ctx context.Context, prior, read types.Set) types.Set {
	if read.IsNull() && !prior.IsNull() && !prior.IsUnknown() && len(prior.Elements()) == 0 {
		return types.SetValueMust(read.ElementType(ctx), nil)
	}
	return read
}

// ListEmptyLikePrior is SetEmptyLikePrior for a list.
func ListEmptyLikePrior(ctx context.Context, prior, read types.List) types.List {
	if read.IsNull() && !prior.IsNull() && !prior.IsUnknown() && len(prior.Elements()) == 0 {
		return types.ListValueMust(read.ElementType(ctx), nil)
	}
	return read
}
