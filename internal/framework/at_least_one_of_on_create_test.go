package framework_test

import (
	"context"
	"testing"

	"terraform-provider-kion/internal/framework"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ownerSchema() schema.Schema {
	return schema.Schema{Attributes: map[string]schema.Attribute{
		"name":                 schema.StringAttribute{Required: true},
		"owner_user_ids":       schema.SetAttribute{ElementType: types.Int64Type, Optional: true},
		"owner_user_group_ids": schema.SetAttribute{ElementType: types.Int64Type, Optional: true},
	}}
}

var ownerObj = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"name":                 tftypes.String,
	"owner_user_ids":       tftypes.Set{ElementType: tftypes.Number},
	"owner_user_group_ids": tftypes.Set{ElementType: tftypes.Number},
}}

func ownerRaw(users []int64) tftypes.Value {
	set := tftypes.NewValue(tftypes.Set{ElementType: tftypes.Number}, nil)
	if users != nil {
		elems := make([]tftypes.Value, len(users))
		for i, u := range users {
			elems[i] = tftypes.NewValue(tftypes.Number, u)
		}
		set = tftypes.NewValue(tftypes.Set{ElementType: tftypes.Number}, elems)
	}
	return tftypes.NewValue(ownerObj, map[string]tftypes.Value{
		"name":                 tftypes.NewValue(tftypes.String, "x"),
		"owner_user_ids":       set,
		"owner_user_group_ids": tftypes.NewValue(tftypes.Set{ElementType: tftypes.Number}, nil),
	})
}

// plan runs the modifier the helper attached to owner_user_group_ids, the first
// attribute named, with config users and a null or populated prior state.
func plan(t *testing.T, users []int64, creating bool) planmodifier.SetResponse {
	t.Helper()
	s := ownerSchema()
	framework.AtLeastOneOfOnCreate(&s, "owner_user_group_ids", "owner_user_ids")

	attr, ok := s.Attributes["owner_user_group_ids"].(schema.SetAttribute)
	require.True(t, ok)
	mods := attr.PlanModifiers
	require.Len(t, mods, 1)

	state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(ownerObj, nil)}
	if !creating {
		state.Raw = ownerRaw(nil)
	}
	req := planmodifier.SetRequest{
		Config: tfsdk.Config{Schema: s, Raw: ownerRaw(users)},
		Plan:   tfsdk.Plan{Schema: s, Raw: ownerRaw(users)},
		State:  state,
	}
	resp := planmodifier.SetResponse{}
	mods[0].PlanModifySet(context.Background(), req, &resp)
	return resp
}

func TestAtLeastOneOfOnCreate_createWithNoneFails(t *testing.T) {
	t.Parallel()
	resp := plan(t, nil, true)
	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(),
		"At least one of these attributes must be configured: [owner_user_group_ids,owner_user_ids]")
}

func TestAtLeastOneOfOnCreate_createWithOnePasses(t *testing.T) {
	t.Parallel()
	assert.False(t, plan(t, []int64{1}, true).Diagnostics.HasError())
}

// An imported or existing record Kion holds with no owners -- built-in
// compliance checks, Kion-managed policies -- must stay plannable.
func TestAtLeastOneOfOnCreate_existingRecordWithNonePasses(t *testing.T) {
	t.Parallel()
	assert.False(t, plan(t, nil, false).Diagnostics.HasError())
}

func TestAtLeastOneOfOnCreate_unsupportedTypePanics(t *testing.T) {
	t.Parallel()
	s := schema.Schema{Attributes: map[string]schema.Attribute{
		"flag":  schema.BoolAttribute{Optional: true},
		"other": schema.BoolAttribute{Optional: true},
	}}
	assert.Panics(t, func() { framework.AtLeastOneOfOnCreate(&s, "flag", "other") })
}
