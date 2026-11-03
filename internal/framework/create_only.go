package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// CreateOnlyString marks an attribute Kion uses only when creating the record,
// with no way to change it after (an account's payer, its AWS OU placement). A
// change is warned about at plan rather than dropped in silence. Replacement is
// not an option there: it would create a different AWS account.
func CreateOnlyString() planmodifier.String { return createOnly{} }

// CreateOnlyInt64 is CreateOnlyString for an int64 attribute.
func CreateOnlyInt64() planmodifier.Int64 { return createOnly{} }

// CreateOnlyBool is CreateOnlyString for a bool attribute.
func CreateOnlyBool() planmodifier.Bool { return createOnly{} }

// CreateOnlyObject is CreateOnlyString for a nested object attribute.
func CreateOnlyObject() planmodifier.Object { return createOnly{} }

type createOnly struct{}

func (createOnly) Description(context.Context) string {
	return "only used when the resource is created; later changes are not sent"
}

func (m createOnly) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }

func (createOnly) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	resp.Diagnostics.Append(createOnlyWarn(req.Path, req.StateValue, req.PlanValue)...)
}

func (createOnly) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	resp.Diagnostics.Append(createOnlyWarn(req.Path, req.StateValue, req.PlanValue)...)
}

func (createOnly) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	resp.Diagnostics.Append(createOnlyWarn(req.Path, req.StateValue, req.PlanValue)...)
}

func (createOnly) PlanModifyObject(_ context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	resp.Diagnostics.Append(createOnlyWarn(req.Path, req.StateValue, req.PlanValue)...)
}

// createOnlyWarn warns when an existing resource's value would change. A null
// state is a create (or an attribute that was never set); an unknown plan value
// is decided at apply and cannot be compared.
func createOnlyWarn(at path.Path, state, plan attr.Value) diag.Diagnostics {
	if state == nil || plan == nil || state.IsNull() || plan.IsUnknown() || state.Equal(plan) {
		return nil
	}
	var d diag.Diagnostics
	d.AddAttributeWarning(at, "Change will not be applied",
		fmt.Sprintf("%s is only used when the resource is created, and Kion has no way to change it afterwards, "+
			"so this change will not reach Kion and the plan will keep showing it. To apply it, replace the "+
			"resource; for an account that means creating a new one.", at))
	return d
}
