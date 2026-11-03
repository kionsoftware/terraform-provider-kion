package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"terraform-provider-kion/internal/service/billing_source_aws"
)

// The API uses an AWS access key pair only when both halves are set, and an
// update that sends one keeps the stored pair; either alone must fail at plan.
func TestBillingSourceAwsKeysRequiredTogether(t *testing.T) {
	ctx := context.Background()
	r := billing_source_aws.NewBillingSourceAwsResource()
	var sresp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sresp)
	cv, ok := r.(resource.ResourceWithConfigValidators)
	if !ok {
		t.Fatal("kion_billing_source_aws declares no config validators")
	}
	objType, ok := sresp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("schema type is not an object")
	}

	config := func(set map[string]string) tfsdk.Config {
		vals := map[string]tftypes.Value{}
		for name, typ := range objType.AttributeTypes {
			if v, ok := set[name]; ok {
				vals[name] = tftypes.NewValue(tftypes.String, v)
			} else {
				vals[name] = tftypes.NewValue(typ, nil)
			}
		}
		return tfsdk.Config{Schema: sresp.Schema, Raw: tftypes.NewValue(objType, vals)}
	}

	cases := map[string]struct {
		set     map[string]string
		wantErr bool
	}{
		"neither":         {map[string]string{}, false},
		"both":            {map[string]string{"key_id": "AKIA", "key_secret": "s"}, false},
		"key_id alone":    {map[string]string{"key_id": "AKIA"}, true},
		"key_secret only": {map[string]string{"key_secret": "s"}, true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := resource.ValidateConfigRequest{Config: config(tc.set)}
			var resp resource.ValidateConfigResponse
			for _, v := range cv.ConfigValidators(ctx) {
				v.ValidateResource(ctx, req, &resp)
			}
			if got := resp.Diagnostics.HasError(); got != tc.wantErr {
				t.Fatalf("error = %v, want %v: %v", got, tc.wantErr, resp.Diagnostics)
			}
		})
	}
}
