package kalign

import (
	"fmt"
	"sort"
	"strings"
)

// Resolve aligns one ServiceModel to the SDK type whose json tags best cover the
// model's tfsdk tags, then records drift: attributes with no SDK field, primitive
// type mismatches, and missing flex converters. flexFuncs is the set of existing
// flex function names (used to flag missing converters).
func Resolve(m ServiceModel, sdkTypes map[string][]SDKField, flexFuncs map[string]bool) Resolved {
	r := Resolved{Model: m}
	tfsdkSet := make(map[string]bool, len(m.Fields))
	for _, f := range m.Fields {
		tfsdkSet[f.TFSDK] = true
	}
	r.SDKType, r.Overlap = bestOverlapType(m.Service, tfsdkSet, sdkTypes)
	if r.SDKType == "" || r.Overlap*2 < len(m.Fields) {
		r.LowConfidence = true
	}

	sdkByJSON := make(map[string]SDKField, len(sdkTypes[r.SDKType]))
	for _, f := range sdkTypes[r.SDKType] {
		sdkByJSON[f.JSON] = f
	}
	for _, mf := range m.Fields {
		sf, ok := sdkByJSON[mf.TFSDK]
		if !ok {
			r.MissingInSDK = append(r.MissingInSDK, mf.TFSDK)
			continue
		}
		p := Pair{Model: mf, SDK: sf}
		if tfFamily(mf.TFType) == "" { // nested/object/list
			p.Nested = true
			r.NestedAttrs = append(r.NestedAttrs, mf.TFSDK)
		} else {
			sdkType := sf.GoType
			if sf.Underlying != "" {
				sdkType = sf.Underlying
			}
			if !typesCompatible(mf.TFType, sdkType) && !deliberateStringID(mf.TFSDK, mf.TFType, sf.GoType) {
				r.TypeMismatch = append(r.TypeMismatch,
					fmt.Sprintf("%s: schema %s vs SDK %s", mf.TFSDK, mf.TFType, sf.GoType))
				r.findings = append(r.findings, Finding{m.Service, mf.TFSDK, KindType})
			}
			// Exported: flex converters are StringToFramework, not
			// stringToFramework. Concatenating the Go type name verbatim made
			// every PRIMITIVE field a guaranteed false positive -- string, int,
			// uint and float could never match, which was 48 of the 110 "missing
			// converter" findings. SDK wrapper types (OptString, ...) already
			// start with a capital, which is why only primitives showed.
			p.FlexFn = exportedConverter(strings.TrimPrefix(sf.GoType, "*"))
			p.HaveFlex = flexFuncs[p.FlexFn]
			if !p.HaveFlex {
				r.MissingFlex = append(r.MissingFlex,
					fmt.Sprintf("%s (for field %q of type %s)", p.FlexFn, mf.TFSDK, sf.GoType))
				r.findings = append(r.findings, Finding{m.Service, mf.TFSDK, KindFlex})
			}
		}
		r.Pairs = append(r.Pairs, p)
	}
	return r
}

// bestOverlapType picks the SDK struct whose json tags cover the most of the
// model's tfsdk tags. A tie goes to a type named after the service, then breaks
// lexically for determinism. Returns ("", 0) when nothing overlaps.
func bestOverlapType(service string, tfsdkSet map[string]bool, sdkTypes map[string][]SDKField) (string, int) {
	names := make([]string, 0, len(sdkTypes))
	for n := range sdkTypes {
		names = append(names, n)
	}
	sort.Strings(names)
	best, bestOverlap := "", 0
	for _, name := range names {
		overlap := 0
		for _, f := range sdkTypes[name] {
			if tfsdkSet[f.JSON] {
				overlap++
			}
		}
		switch {
		case overlap > bestOverlap:
			best, bestOverlap = name, overlap
		case overlap == bestOverlap && overlap > 0 && namedFor(name, service) && !namedFor(best, service):
			best = name
		}
	}
	return best, bestOverlap
}

// namedFor reports whether an SDK type name starts with the service name,
// ignoring case and underscores ("OUCreate" is named for "ou").
func namedFor(typeName, service string) bool {
	s := strings.ReplaceAll(strings.ToLower(service), "_", "")
	return s != "" && strings.HasPrefix(strings.ToLower(typeName), s)
}

// typesCompatible reports whether a Framework type and an SDK Go type describe
// the same primitive family. Nested/list/object Framework types return true
// (they are handled as nested, not primitive drift).
func typesCompatible(tf, sdk string) bool {
	fam := tfFamily(tf)
	if fam == "" {
		return true
	}
	return fam == sdkFamily(sdk)
}

// tfFamily maps a terraform-plugin-framework type to a primitive family, or ""
// for nested/collection types.
func tfFamily(tf string) string {
	switch tf {
	case "types.String":
		return "string"
	case "types.Bool":
		return "bool"
	case "types.Int64", "types.Int32", "types.Number":
		return "int"
	case "types.Float64", "types.Float32":
		return "float"
	default:
		return ""
	}
}

// sdkFamily strips ogen optionality prefixes (Opt/Nil/Null) and a leading
// pointer, then maps the base type to a primitive family.
func sdkFamily(sdk string) string {
	base := strings.ToLower(stripWrappers(sdk))
	switch {
	case base == "string":
		return "string"
	case base == "bool":
		return "bool"
	case strings.HasPrefix(base, "int") || strings.HasPrefix(base, "uint"):
		return "int"
	case strings.HasPrefix(base, "float"):
		return "float"
	case base == "time" || base == "datetime":
		return "string" // Terraform carries timestamps as RFC3339 strings
	default:
		return base
	}
}

// exportedConverter builds the flex converter name for an SDK field type.
// flex's converters are exported, so the type's first letter is capitalized:
// "string" -> "StringToFramework", "OptString" -> "OptStringToFramework".
func exportedConverter(goType string) string {
	if goType == "" {
		return "ToFramework"
	}
	return strings.ToUpper(goType[:1]) + goType[1:] + "ToFramework"
}

// deliberateStringID reports the one type mismatch this provider chooses on
// purpose: every resource id is a Terraform string (the framework's
// IDAttribute convention, and what the shipped SDKv2 provider used) while Kion
// models ids as integers.
//
// It is 43 of the 51 TYPE findings. Reporting it drowns the other 8, each of
// which is a real question about how an attribute is modeled.
func deliberateStringID(tfsdk, tfType, sdkType string) bool {
	if tfsdk != "id" || tfType != "types.String" {
		return false
	}
	switch strings.TrimPrefix(strings.TrimPrefix(sdkType, "*"), "Opt") {
	case "Uint64", "Int64", "Uint", "Int", "NilUint64", "NilInt64":
		return true
	}
	return false
}
