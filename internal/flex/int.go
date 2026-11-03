// Package flex provides helper functions for converting between Terraform Framework types and SDK types.
package flex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
	"strconv"
)

// --- Expand: TF → SDK ---

// Int64ValueFromFramework returns the Go int64 value from a types.Int64.
func Int64ValueFromFramework(v types.Int64) int64 {
	return v.ValueInt64()
}

// Uint64FromFramework converts a types.Int64 to a uint64.
// Terraform has no unsigned integer type, so we use Int64 and cast.
func Uint64FromFramework(v types.Int64) uint64 {
	return uint64(v.ValueInt64())
}

// OptInt64FromFramework converts a types.Int64 to an OptInt64.
// Null or unknown values produce an unset OptInt64.
func OptInt64FromFramework(v types.Int64) generated.OptInt64 {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptInt64{}
	}
	return generated.OptInt64{Value: v.ValueInt64(), Set: true}
}

// OptUint64FromFramework converts a types.Int64 to an OptUint64.
// Null or unknown values produce an unset OptUint64.
func OptUint64FromFramework(v types.Int64) generated.OptUint64 {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptUint64{}
	}
	return generated.OptUint64{Value: uint64(v.ValueInt64()), Set: true}
}

// OptUint64FromFrameworkFloat64 converts a types.Float64 to an OptUint64.
// Null or unknown values produce an unset OptUint64. It exists for an attribute
// the schema types as a number while the API takes a whole count -- a project
// budget total, in dollars -- so a fractional value truncates toward zero
// rather than being refused, and a negative one clamps to 0.
func OptUint64FromFrameworkFloat64(v types.Float64) generated.OptUint64 {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptUint64{}
	}
	f := v.ValueFloat64()
	if f < 0 {
		f = 0
	}
	return generated.OptUint64{Value: uint64(f), Set: true}
}

// --- Flatten: SDK → TF ---

// Int64ToFramework converts a Go int64 to a types.Int64.
func Int64ToFramework(v int64) types.Int64 {
	return types.Int64Value(v)
}

// Uint64ToFramework converts a Go uint64 to a types.Int64.
func Uint64ToFramework(v uint64) types.Int64 {
	return types.Int64Value(int64(v))
}

// OptInt64ToFramework converts an OptInt64 to a types.Int64.
// Unset values produce a null types.Int64.
func OptInt64ToFramework(v generated.OptInt64) types.Int64 {
	if !v.IsSet() {
		return types.Int64Null()
	}
	return types.Int64Value(v.Value)
}

// OptUint64ToFramework converts an OptUint64 to a types.Int64.
// Unset values produce a null types.Int64.
func OptUint64ToFramework(v generated.OptUint64) types.Int64 {
	if !v.IsSet() {
		return types.Int64Null()
	}
	return types.Int64Value(int64(v.Value))
}

// --- OptNilUint64 converters ---

// OptNilUint64FromFramework converts a types.Int64 to an OptNilUint64.
// Null or unknown values produce an unset OptNilUint64.
func OptNilUint64FromFramework(v types.Int64) generated.OptNilUint64 {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptNilUint64{}
	}
	return generated.NewOptNilUint64(uint64(v.ValueInt64()))
}

// OptNilUint64ToFramework converts an OptNilUint64 to a types.Int64.
// Unset or null values produce a null types.Int64.
func OptNilUint64ToFramework(v generated.OptNilUint64) types.Int64 {
	val, ok := v.Get()
	if !ok {
		return types.Int64Null()
	}
	return types.Int64Value(int64(val))
}

// NilUint64FromFramework converts a types.Int64 to a NilUint64.
// Null or unknown values produce a null NilUint64.
func NilUint64FromFramework(v types.Int64) generated.NilUint64 {
	if v.IsNull() || v.IsUnknown() {
		return generated.NilUint64{Null: true}
	}
	return generated.NilUint64{Value: uint64(v.ValueInt64())}
}

// NilUint64ToFramework converts a NilUint64 to a types.Int64.
// Null values produce a null types.Int64.
func NilUint64ToFramework(v generated.NilUint64) types.Int64 {
	if v.Null {
		return types.Int64Null()
	}
	return types.Int64Value(int64(v.Value))
}

// --- GCPRoleLaunchStage (a uint64-backed ogen enum) ---

// OptGCPRoleLaunchStageFromFramework converts a types.Int64 to an OptGCPRoleLaunchStage.
// Null or unknown values produce an unset OptGCPRoleLaunchStage.
func OptGCPRoleLaunchStageFromFramework(v types.Int64) generated.OptGCPRoleLaunchStage {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptGCPRoleLaunchStage{}
	}
	return generated.OptGCPRoleLaunchStage{Value: generated.GCPRoleLaunchStage(v.ValueInt64()), Set: true}
}

// OptGCPRoleLaunchStageToFramework converts an OptGCPRoleLaunchStage to a types.Int64.
// Unset values produce a null types.Int64.
func OptGCPRoleLaunchStageToFramework(v generated.OptGCPRoleLaunchStage) types.Int64 {
	if !v.IsSet() {
		return types.Int64Null()
	}
	return types.Int64Value(int64(v.Value))
}

// OptNilInt64FromFramework converts a Terraform Int64 to a nil-aware SDK
// OptNilInt64 (null/unknown -> unset).
func OptNilInt64FromFramework(v types.Int64) generated.OptNilInt64 {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptNilInt64{}
	}
	return generated.NewOptNilInt64(v.ValueInt64())
}

// OptNilInt64ToFramework converts a nil-aware SDK OptNilInt64 to a Terraform
// Int64 (unset -> null).
func OptNilInt64ToFramework(v generated.OptNilInt64) types.Int64 {
	val, ok := v.Get()
	if !ok {
		return types.Int64Null()
	}
	return types.Int64Value(val)
}

// DatecodeToFramework renders a Kion YYYYMM datecode as the YYYY-MM string the
// public schema validates.
//
// The private reads these resources use return billing_start_date as the number
// 202406, while the public /v4 endpoint the spec is generated from returns
// "2024-06". Formatting the number as a decimal string produced "202406", which
// the schema's own ^\d{4}-(?:0[1-9]|1[0-2])$ validator then rejected, so
// kion_billing_source could never be imported. 0 means unset and maps to null
// rather than "0".
func DatecodeToFramework(v int64) types.String {
	if v == 0 {
		return types.StringNull()
	}
	year, month := v/100, v%100
	if year < 1000 || month < 1 || month > 12 {
		// Not a datecode after all; hand it back unchanged rather than inventing
		// a shape, so the validator reports the real value.
		return types.StringValue(strconv.FormatInt(v, 10))
	}
	return types.StringValue(fmt.Sprintf("%04d-%02d", year, month))
}

// NullIntToFramework converts Kion's sql.NullInt64 wire shape,
// {"Int":1,"Valid":true}, to a types.Int64 that is null when not valid.
func NullIntToFramework(v int64, valid bool) types.Int64 {
	if !valid {
		return types.Int64Null()
	}
	return types.Int64Value(v)
}
