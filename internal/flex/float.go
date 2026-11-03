package flex

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
)

// --- Expand: TF → SDK ---

// Float64ValueFromFramework returns the Go float64 value from a types.Float64.
func Float64ValueFromFramework(v types.Float64) float64 {
	return v.ValueFloat64()
}

// OptFloat64FromFramework converts a types.Float64 to a generated.OptFloat64,
// leaving it unset when null or unknown.
func OptFloat64FromFramework(v types.Float64) generated.OptFloat64 {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptFloat64{}
	}
	return generated.NewOptFloat64(v.ValueFloat64())
}

// --- Flatten: SDK → TF ---

// Float64ToFramework converts a Go float64 to a types.Float64.
func Float64ToFramework(v float64) types.Float64 {
	return types.Float64Value(v)
}

// OptFloat64ToFramework converts a generated.OptFloat64 to a types.Float64,
// yielding null when unset.
func OptFloat64ToFramework(v generated.OptFloat64) types.Float64 {
	if !v.Set {
		return types.Float64Null()
	}
	return types.Float64Value(v.Value)
}
