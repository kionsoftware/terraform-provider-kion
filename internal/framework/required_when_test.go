package framework_test

import (
	"context"
	"testing"

	"terraform-provider-kion/internal/framework"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// carSchema mirrors the shape the rule guards: a role-type discriminator and
// the field each non-User type needs.
func carSchema() rsschema.Schema {
	return rsschema.Schema{
		Attributes: map[string]rsschema.Attribute{
			"cloud_access_role_type_id": rsschema.Int64Attribute{Optional: true},
			"aws_iam_role_trust_policy": rsschema.StringAttribute{Optional: true},
		},
	}
}

func carConfig(t *testing.T, roleType *int64, trustPolicy *string) tfsdk.Config {
	t.Helper()
	obj := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"cloud_access_role_type_id": tftypes.Number,
		"aws_iam_role_trust_policy": tftypes.String,
	}}
	vals := map[string]tftypes.Value{
		"cloud_access_role_type_id": tftypes.NewValue(tftypes.Number, nil),
		"aws_iam_role_trust_policy": tftypes.NewValue(tftypes.String, nil),
	}
	if roleType != nil {
		vals["cloud_access_role_type_id"] = tftypes.NewValue(tftypes.Number, *roleType)
	}
	if trustPolicy != nil {
		vals["aws_iam_role_trust_policy"] = tftypes.NewValue(tftypes.String, *trustPolicy)
	}
	return tfsdk.Config{Raw: tftypes.NewValue(obj, vals), Schema: carSchema()}
}

func validate(t *testing.T, roleType *int64, trustPolicy *string) *resource.ValidateConfigResponse {
	t.Helper()
	v := framework.RequiredWhenInt64("cloud_access_role_type_id", 2, "aws_iam_role_trust_policy")
	resp := &resource.ValidateConfigResponse{}
	v.ValidateResource(context.Background(),
		resource.ValidateConfigRequest{Config: carConfig(t, roleType, trustPolicy)}, resp)
	return resp
}

func i64(v int64) *int64   { return &v }
func str(v string) *string { return &v }

// Custom Trust without a trust policy is the case the API rejects at apply.
func TestRequiredWhenInt64_firesOnMatchingValue(t *testing.T) {
	resp := validate(t, i64(2), nil)
	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics[0].Detail(), "aws_iam_role_trust_policy")
	assert.Contains(t, resp.Diagnostics[0].Detail(), "cloud_access_role_type_id is 2")
}

func TestRequiredWhenInt64_quietWhenSatisfied(t *testing.T) {
	resp := validate(t, i64(2), str(`{"Version":"2012-10-17"}`))
	assert.False(t, resp.Diagnostics.HasError())
}

// Another type has its own required field; this rule must not fire for it.
func TestRequiredWhenInt64_quietOnOtherValue(t *testing.T) {
	resp := validate(t, i64(3), nil)
	assert.False(t, resp.Diagnostics.HasError())
}

// An unset discriminator takes the API default, so nothing is required yet.
func TestRequiredWhenInt64_quietWhenTriggerUnset(t *testing.T) {
	resp := validate(t, nil, nil)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestRequiredWhenInt64_descriptionNamesBothSides(t *testing.T) {
	v := framework.RequiredWhenInt64("cloud_access_role_type_id", 4, "aws_trusted_services")
	d := v.Description(context.Background())
	assert.Contains(t, d, "aws_trusted_services")
	assert.Contains(t, d, "cloud_access_role_type_id")
	assert.Contains(t, d, "4")
}
