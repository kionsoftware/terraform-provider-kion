package framework_test

import (
	"context"
	"testing"

	"terraform-provider-kion/internal/framework"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func createOnlyString(t *testing.T, state, plan types.String) planmodifier.StringResponse {
	t.Helper()
	req := planmodifier.StringRequest{Path: path.Root("account_number"), StateValue: state, PlanValue: plan, ConfigValue: plan}
	resp := planmodifier.StringResponse{PlanValue: plan}
	framework.CreateOnlyString().PlanModifyString(context.Background(), req, &resp)
	return resp
}

// A change after create cannot reach Kion and replacing would create a new AWS
// account, so the change is reported rather than dropped in silence.
func TestCreateOnly_changeOnExistingWarns(t *testing.T) {
	t.Parallel()
	resp := createOnlyString(t, types.StringValue("111111111111"), types.StringValue("222222222222"))
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, 1, resp.Diagnostics.WarningsCount())
	assert.Contains(t, resp.Diagnostics.Warnings()[0].Detail(), "only used when the resource is created")
}

func TestCreateOnly_silentOnCreateAndNoChange(t *testing.T) {
	t.Parallel()
	assert.Empty(t, createOnlyString(t, types.StringNull(), types.StringValue("1")).Diagnostics, "create")
	assert.Empty(t, createOnlyString(t, types.StringValue("1"), types.StringValue("1")).Diagnostics, "unchanged")
	assert.Empty(t, createOnlyString(t, types.StringValue("1"), types.StringUnknown()).Diagnostics, "unknown until apply")
}

func TestCreateOnly_otherKinds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ir := planmodifier.Int64Response{}
	framework.CreateOnlyInt64().PlanModifyInt64(ctx, planmodifier.Int64Request{StateValue: types.Int64Value(1), PlanValue: types.Int64Value(2)}, &ir)
	assert.Equal(t, 1, ir.Diagnostics.WarningsCount())
	br := planmodifier.BoolResponse{}
	framework.CreateOnlyBool().PlanModifyBool(ctx, planmodifier.BoolRequest{StateValue: types.BoolValue(false), PlanValue: types.BoolValue(true)}, &br)
	assert.Equal(t, 1, br.Diagnostics.WarningsCount())
}
