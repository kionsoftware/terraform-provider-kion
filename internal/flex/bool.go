package flex

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
)

// --- Expand: TF → SDK ---

// BoolValueFromFramework returns the Go bool value from a types.Bool.
func BoolValueFromFramework(v types.Bool) bool {
	return v.ValueBool()
}

// OptBoolFromFramework converts a types.Bool to an OptBool.
// Null or unknown values produce an unset OptBool.
func OptBoolFromFramework(v types.Bool) generated.OptBool {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptBool{}
	}
	return generated.OptBool{Value: v.ValueBool(), Set: true}
}

// --- Flatten: SDK → TF ---

// BoolToFramework converts a Go bool to a types.Bool.
func BoolToFramework(v bool) types.Bool {
	return types.BoolValue(v)
}

// OptBoolToFramework converts an OptBool to a types.Bool.
// Unset values produce a null types.Bool.
func OptBoolToFramework(v generated.OptBool) types.Bool {
	if !v.IsSet() {
		return types.BoolNull()
	}
	return types.BoolValue(v.Value)
}

// --- OptNilBool converters ---

// OptNilBoolFromFramework converts a types.Bool to an OptNilBool.
// Null or unknown values produce an unset OptNilBool.
func OptNilBoolFromFramework(v types.Bool) generated.OptNilBool {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptNilBool{}
	}
	return generated.NewOptNilBool(v.ValueBool())
}

// OptNilBoolToFramework converts an OptNilBool to a types.Bool.
// Unset or null values produce a null types.Bool.
func OptNilBoolToFramework(v generated.OptNilBool) types.Bool {
	val, ok := v.Get()
	if !ok {
		return types.BoolNull()
	}
	return types.BoolValue(val)
}

// --- NilBool converters (always-present, nullable: {Value, Null}) ---

// NilBoolFromFramework converts a types.Bool to a NilBool.
// Null or unknown values produce a null NilBool.
func NilBoolFromFramework(v types.Bool) generated.NilBool {
	if v.IsNull() || v.IsUnknown() {
		return generated.NilBool{Null: true}
	}
	return generated.NilBool{Value: v.ValueBool()}
}

// NilBoolToFramework converts a NilBool to a types.Bool.
// Null values produce a null types.Bool.
func NilBoolToFramework(v generated.NilBool) types.Bool {
	if v.Null {
		return types.BoolNull()
	}
	return types.BoolValue(v.Value)
}
