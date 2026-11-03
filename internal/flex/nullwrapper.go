package flex

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
)

// Converters for the SDK's SQL-style nullable wrappers (OptNullString /
// OptNullTime), whose inner {String|Time, Valid} shape ogen exposes but which
// are semantically a nullable scalar. The matching schema override retypes the
// attribute to a plain string so the model side stays a types.String.

// OptNullStringFromFramework converts a types.String to an OptNullString.
// Null or unknown values produce an unset wrapper; a value produces a set,
// valid NullString.
func OptNullStringFromFramework(v types.String) generated.OptNullString {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptNullString{}
	}
	return generated.OptNullString{
		Value: generated.NullString{
			String: generated.OptString{Value: v.ValueString(), Set: true},
			Valid:  generated.NewOptNilBool(true),
		},
		Set: true,
	}
}

// OptNullStringToFramework converts an OptNullString to a types.String.
// Unset, invalid (Valid=false), or unset-inner-String values produce a null.
func OptNullStringToFramework(v generated.OptNullString) types.String {
	nv, ok := v.Get()
	if !ok {
		return types.StringNull()
	}
	if valid, ok := nv.Valid.Get(); ok && !valid {
		return types.StringNull()
	}
	s, ok := nv.String.Get()
	if !ok {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// OptNullTimeFromFramework converts an RFC3339 types.String to an OptNullTime.
// Null, unknown, or unparsable values produce an unset wrapper.
func OptNullTimeFromFramework(v types.String) generated.OptNullTime {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptNullTime{}
	}
	t, err := time.Parse(time.RFC3339, v.ValueString())
	if err != nil {
		return generated.OptNullTime{}
	}
	return generated.OptNullTime{
		Value: generated.NullTime{
			Time:  generated.OptDateTime{Value: t, Set: true},
			Valid: generated.NewOptNilBool(true),
		},
		Set: true,
	}
}

// OptNullTimeToFramework converts an OptNullTime to an RFC3339 types.String.
// Unset, invalid, or unset-inner-Time values produce a null.
func OptNullTimeToFramework(v generated.OptNullTime) types.String {
	nv, ok := v.Get()
	if !ok {
		return types.StringNull()
	}
	if valid, ok := nv.Valid.Get(); ok && !valid {
		return types.StringNull()
	}
	t, ok := nv.Time.Get()
	if !ok {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}
