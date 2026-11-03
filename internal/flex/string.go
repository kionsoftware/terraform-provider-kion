package flex

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
)

// --- Expand: TF → SDK ---

// StringValueFromFramework returns the Go string value from a types.String.
func StringValueFromFramework(v types.String) string {
	return v.ValueString()
}

// OptStringFromFramework converts a types.String to an OptString.
// Null or unknown values produce an unset OptString.
func OptStringFromFramework(v types.String) generated.OptString {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptString{}
	}
	return generated.OptString{Value: v.ValueString(), Set: true}
}

// --- Flatten: SDK → TF ---

// StringToFramework converts a Go string to a types.String.
func StringToFramework(v string) types.String {
	v = CanonicalJSONString(v)
	return types.StringValue(v)
}

// OptStringToFramework converts an OptString to a types.String.
// Unset values produce a null types.String.
func OptStringToFramework(v generated.OptString) types.String {
	if v.Set {
		v.Value = CanonicalJSONString(v.Value)
	}
	if !v.IsSet() {
		return types.StringNull()
	}
	return types.StringValue(v.Value)
}

// OptStringToFrameworkLegacy converts an OptString to a types.String,
// treating unset values as empty string "" rather than null. This matches
// the behavior of terraform-plugin-sdk/v2 (the legacy SDK), which
// represented "no value" as an empty string rather than as a distinct
// null. Use this for resources migrated from the legacy SDK so existing
// configs don't see diffs when the API omits an optional attribute.
func OptStringToFrameworkLegacy(v generated.OptString) types.String {
	if !v.IsSet() {
		return types.StringValue("")
	}
	return types.StringValue(v.Value)
}

// OptStringToFrameworkNullIfEmpty converts an OptString to a types.String,
// treating a set-but-empty value as null.
//
// The account endpoints return account_alias as "" when an account has none.
// Storing that as "" made every imported account plan an update, because
// `terraform plan -generate-config-out` writes an unset optional attribute as
// null: 176 accounts changed on `"" -> null` alone, with nothing else to show
// for it. For these fields absent and empty mean the same thing.
func OptStringToFrameworkNullIfEmpty(v generated.OptString) types.String {
	if !v.IsSet() || v.Value == "" {
		return types.StringNull()
	}
	return types.StringValue(v.Value)
}

// NullStringToFramework converts Kion's sql.NullString wire shape,
// {"String":"x","Valid":true}, to a types.String that is null when not valid.
func NullStringToFramework(v string, valid bool) types.String {
	if !valid {
		return types.StringNull()
	}
	return types.StringValue(CanonicalJSONString(v))
}
