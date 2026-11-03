package flex_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/flex"
)

var tagType = types.ObjectType{AttrTypes: map[string]attr.Type{"tag_key": types.StringType, "id": types.Int64Type}}

func tags(t *testing.T, keys []string, ids []attr.Value) types.List {
	t.Helper()
	elems := make([]attr.Value, len(keys))
	for i, k := range keys {
		elems[i] = types.ObjectValueMust(tagType.AttrTypes, map[string]attr.Value{"tag_key": types.StringValue(k), "id": ids[i]})
	}
	return types.ListValueMust(tagType, elems)
}

func keys(t *testing.T, l types.List) []string {
	t.Helper()
	var out []string
	for _, e := range l.Elements() {
		obj, ok := e.(types.Object)
		require.True(t, ok)
		k, ok := obj.Attributes()["tag_key"].(types.String)
		require.True(t, ok)
		out = append(out, k.ValueString())
	}
	return out
}

// The plan's computed ids are unknown; the read's are set and in Kion's order.
func TestOrderLikePrior(t *testing.T) {
	ctx := context.Background()
	unknown := types.Int64Unknown()
	prior := tags(t, []string{"team", "env"}, []attr.Value{unknown, unknown})
	read := tags(t, []string{"added", "env", "team"}, []attr.Value{types.Int64Value(3), types.Int64Value(1), types.Int64Value(2)})

	got, diags := flex.OrderLikePrior(ctx, prior, read, "tag_key")
	require.False(t, diags.HasError(), diags)
	assert.Equal(t, []string{"team", "env", "added"}, keys(t, got))

	got, diags = flex.OrderLikePrior(ctx, types.ListNull(tagType), read, "tag_key")
	require.False(t, diags.HasError(), diags)
	assert.Equal(t, []string{"added", "env", "team"}, keys(t, got))
}
