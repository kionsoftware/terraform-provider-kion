package framework

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// RequiredWhenInt64 returns a ConfigValidator that requires each attribute in
// require once trigger holds value. The framework's own validators cover
// presence relationships (AtLeastOneOf, RequiredTogether, Conflicting) but not
// ones conditional on a value.
//
// A null or unknown trigger does not fire: an unset discriminator takes the
// API's default, and an unknown one is not decidable at plan time.
func RequiredWhenInt64(trigger string, value int64, require ...string) resource.ConfigValidator {
	return requiredWhenInt64{trigger: trigger, value: value, require: require}
}

type requiredWhenInt64 struct {
	trigger string
	value   int64
	require []string
}

func (v requiredWhenInt64) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v requiredWhenInt64) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("%s must be set when %s is %d",
		strings.Join(v.require, " and "), v.trigger, v.value)
}

// ValidateResource reads the raw config rather than typed attributes, so one
// validator covers every attribute type: the required fields here are a string,
// a list of strings and a bool, and a typed read would need one implementation
// per type.
func (v requiredWhenInt64) ValidateResource(_ context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if req.Config.Raw.IsNull() || !req.Config.Raw.IsKnown() {
		return
	}
	var obj map[string]tftypes.Value
	if err := req.Config.Raw.As(&obj); err != nil {
		return // not an object; nothing to inspect
	}

	trigger, ok := obj[v.trigger]
	if !ok || trigger.IsNull() || !trigger.IsKnown() {
		return
	}
	var num big.Float
	if err := trigger.As(&num); err != nil {
		return
	}
	if got, _ := num.Int64(); got != v.value {
		return
	}

	for _, name := range v.require {
		attr, ok := obj[name]
		if ok && !attr.IsNull() {
			continue
		}
		resp.Diagnostics.AddAttributeError(
			path.Root(name),
			"Missing required attribute",
			fmt.Sprintf("%s must be set when %s is %d.", name, v.trigger, v.value),
		)
	}
}
