package flex_test

import (
	"context"
	"testing"

	"terraform-provider-kion/internal/flex"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

var exportType = types.ObjectType{AttrTypes: map[string]attr.Type{"dataset_name": types.StringType}}

// On import there is no prior state, so the carried value is a zero
// types.Object with no attribute types, which the object constructor rejects.
func TestCarryOrNull_zeroObjectBecomesTypedNull(t *testing.T) {
	ctx := context.Background()
	got := flex.CarryOrNull(ctx, types.Object{}, exportType)
	assert.True(t, got.IsNull())
	assert.True(t, got.Type(ctx).Equal(exportType))
}

func TestCarryOrNull_keepsPriorValue(t *testing.T) {
	ctx := context.Background()
	prior := types.ObjectValueMust(exportType.AttrTypes, map[string]attr.Value{"dataset_name": types.StringValue("finances")})
	assert.Equal(t, prior, flex.CarryOrNull(ctx, prior, exportType))
	assert.Equal(t, types.StringValue("x"), flex.CarryOrNull(ctx, types.StringValue("x"), types.StringType))
}

func TestObjectFromAttrs(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics
	got := flex.ObjectFromAttrs(ctx, &diags, exportType, map[string]attr.Value{"dataset_name": types.StringValue("finances")})
	assert.False(t, diags.HasError())
	assert.True(t, got.Type(ctx).Equal(exportType))

	flex.ObjectFromAttrs(ctx, &diags, types.StringType, nil)
	assert.True(t, diags.HasError(), "a non-object type is reported, not panicked on")
}
