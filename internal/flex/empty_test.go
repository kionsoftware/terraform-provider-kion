package flex_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-kion/internal/flex"
)

func TestSetEmptyLikePrior(t *testing.T) {
	ctx := context.Background()
	empty := types.SetValueMust(types.Int64Type, nil)
	one := types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(1)})
	null := types.SetNull(types.Int64Type)

	// A configured [] read back as null keeps the empty set.
	assert.True(t, flex.SetEmptyLikePrior(ctx, empty, null).Equal(empty))
	// An unset attribute (or an import, with no prior) stays null.
	assert.True(t, flex.SetEmptyLikePrior(ctx, null, null).IsNull())
	assert.True(t, flex.SetEmptyLikePrior(ctx, types.SetUnknown(types.Int64Type), null).IsNull())
	// A non-empty prior read back as null is drift, and reported as such.
	assert.True(t, flex.SetEmptyLikePrior(ctx, one, null).IsNull())
	// A non-null read is always taken as is.
	assert.True(t, flex.SetEmptyLikePrior(ctx, empty, one).Equal(one))
}

func TestListEmptyLikePrior(t *testing.T) {
	ctx := context.Background()
	empty := types.ListValueMust(types.Int64Type, nil)
	null := types.ListNull(types.Int64Type)

	assert.True(t, flex.ListEmptyLikePrior(ctx, empty, null).Equal(empty))
	assert.True(t, flex.ListEmptyLikePrior(ctx, null, null).IsNull())
}
