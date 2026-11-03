package flex

import "github.com/hashicorp/terraform-plugin-framework/types"

// Pointer expanders for raw request bodies: nil for null or unknown, so an
// omitempty field is left out, and a pointer for any known value, zero included.

// StringPointerFromFramework returns nil for a null or unknown string.
func StringPointerFromFramework(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// BoolPointerFromFramework returns nil for a null or unknown bool.
func BoolPointerFromFramework(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// Int64PointerFromFramework returns nil for a null or unknown int64.
func Int64PointerFromFramework(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}
