package crud

// Converter selection is mechanical: an SDK field's Go type + optionality maps
// to exactly one internal/flex helper. expandConverter is TF->SDK (request
// bodies); flattenConverter is SDK->TF (response envelopes). Both return
// ok=false for a type the entity archetype does not handle (nested objects,
// slices, sets), the caller then refuses the resource for a crud_override.

// expandConverter returns the flex.*FromFramework helper for a request-body field.
// normalizedConverter handles a field the schema types as jsontypes.Normalized
// while the SDK carries it as a plain string. The SDK type alone cannot tell
// them apart, and comparing JSON byte-for-byte made every plan report a
// whitespace change.
func normalizedConverter(sdkType, tfType string, expand bool) (string, bool) {
	if tfType != "jsontypes.Normalized" {
		return "", false
	}
	switch sdkType {
	case "string":
		if expand {
			return "flex.NormalizedStringFromFramework", true
		}
		return "flex.NormalizedStringToFramework", true
	case "OptString":
		if expand {
			return "flex.OptNormalizedStringFromFramework", true
		}
		return "flex.OptNormalizedStringToFramework", true
	}
	return "", false
}

func expandConverter(f Field) (string, bool) {
	switch f.Type {
	case "string":
		return "flex.StringValueFromFramework", true
	case "OptString":
		return "flex.OptStringFromFramework", true
	case "int64":
		return "flex.Int64ValueFromFramework", true
	case "uint64":
		return "flex.Uint64FromFramework", true
	case "OptInt64":
		return "flex.OptInt64FromFramework", true
	case "OptUint64":
		return "flex.OptUint64FromFramework", true
	case "OptGCPRoleLaunchStage":
		return "flex.OptGCPRoleLaunchStageFromFramework", true
	case "OptNullString":
		return "flex.OptNullStringFromFramework", true
	case "OptNullTime":
		return "flex.OptNullTimeFromFramework", true
	case "float64":
		return "flex.Float64ValueFromFramework", true
	case "OptFloat64":
		return "flex.OptFloat64FromFramework", true
	case "bool":
		return "flex.BoolValueFromFramework", true
	case "Raw":
		return "flex.NormalizedFromFramework", true
	case "OptBool":
		return "flex.OptBoolFromFramework", true
	case "OptNilBool":
		return "flex.OptNilBoolFromFramework", true
	case "NilBool":
		return "flex.NilBoolFromFramework", true
	case "NilUint64":
		return "flex.NilUint64FromFramework", true
	case "OptNilUint64":
		return "flex.OptNilUint64FromFramework", true
	case "OptNilInt64":
		return "flex.OptNilInt64FromFramework", true
	}
	return "", false
}

// flattenConverter returns the flex.*ToFramework helper for a response field.
func flattenConverter(f Field) (string, bool) {
	switch f.Type {
	case "string":
		return "flex.StringToFramework", true
	case "OptString":
		return "flex.OptStringToFramework", true
	case "int64":
		return "flex.Int64ToFramework", true
	case "uint64":
		return "flex.Uint64ToFramework", true
	case "OptInt64":
		return "flex.OptInt64ToFramework", true
	case "OptUint64":
		return "flex.OptUint64ToFramework", true
	case "OptGCPRoleLaunchStage":
		return "flex.OptGCPRoleLaunchStageToFramework", true
	case "OptNullString":
		return "flex.OptNullStringToFramework", true
	case "OptNullTime":
		return "flex.OptNullTimeToFramework", true
	case "float64":
		return "flex.Float64ToFramework", true
	case "OptFloat64":
		return "flex.OptFloat64ToFramework", true
	case "bool":
		return "flex.BoolToFramework", true
	case "Raw":
		return "flex.NormalizedToFramework", true
	case "OptBool":
		return "flex.OptBoolToFramework", true
	case "OptNilBool":
		return "flex.OptNilBoolToFramework", true
	case "NilBool":
		return "flex.NilBoolToFramework", true
	case "NilUint64":
		return "flex.NilUint64ToFramework", true
	case "OptNilUint64":
		return "flex.OptNilUint64ToFramework", true
	case "OptNilInt64":
		return "flex.OptNilInt64ToFramework", true
	}
	return "", false
}

// sliceConverter returns the flex slice helpers (expand, flatten) for a slice
// body/response field. These take (ctx, …) and return diagnostics, so the
// templates emit them as a prelude rather than an inline expression. wrap names
// the ogen nil-aware wrapper (OptNil<T>Array) when the field is wrapped, the
// expand result is boxed as generated.<wrap>{Value: v, Set: true} and the
// flatten source reads its .Value; wrap is "" for a bare slice.
// tfType is the Terraform model field's Go type ("types.List" or "types.Set").
// Association id attributes are sets, see applyAssociationSetDefault in
// internal/kgen/schemas. And a set needs different flex helpers from a list, so
// the converter has to be chosen from the TERRAFORM type, not only the SDK one.
// Keying on the SDK type alone emitted list helpers for set-typed fields, which
// does not compile.
func sliceConverter(sdkType, tfType string) (expand, flatten, wrap string, ok bool) {
	set := tfType == "types.Set"
	pick := func(base, wrap string) (string, string, string, bool) {
		if set {
			return "flex." + base + "SliceFromFrameworkSet", "flex." + base + "SliceToFrameworkSet", wrap, true
		}
		return "flex." + base + "SliceFromFramework", "flex." + base + "SliceToFramework", wrap, true
	}
	switch sdkType {
	case "[]uint64":
		return pick("Uint64", "")
	case "[]string":
		return pick("String", "")
	case "OptNilUint64Array":
		return pick("Uint64", "OptNilUint64Array")
	case "OptNilStringArray":
		return pick("String", "OptNilStringArray")
	}
	return "", "", "", false
}
