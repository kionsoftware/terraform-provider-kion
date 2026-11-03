package crud

import (
	"cmp"
	"slices"
	"strings"
)

// rawValueBind collapses one polymorphic jx.Raw request-body field onto the
// typed schema attributes that stand in for it.
//
// The spec types such a field as nothing at all, so tfplugingen drops it (see
// codegen/README.md, "An untyped spec property drops the whole type") and
// schema_overrides.yaml re-expresses it as <name>_string / <name>_list /
// <name>_map. That leaves one SDK field facing three model attributes, which
// matched none of them: the provider accepted `default_value_string = "x"` and
// sent a body with no default_value in it at all.
type rawValueBind struct {
	SDKField string // "DefaultValue"
	Var      string // "defaultValue"
	Func     string // "buildDefaultValue"
	Attr     string // "default_value"
	Attrs    string // "default_value_string, default_value_list, or default_value_map"
	StringGo string // model field for <attr>_string ("" when the schema has none)
	ListGo   string
	MapGo    string
	// Required mirrors the SDK field's optionality: a bare jx.Raw is a required
	// property, so a configuration that sets none of the variants is an error
	// rather than an omitted field. Kion's custom-variable create is one --
	// it answers a missing default_value with "unknown custom variable type
	// encountered during value creation: <nil>".
	Required bool
}

// rawValueSuffixes are the typed stand-ins a polymorphic field is split into.
// The order is the order they are reported in a diagnostic.
var rawValueSuffixes = []string{"_string", "_list", "_map"}

// resolveRawValues finds the polymorphic jx.Raw fields of a request body and
// binds each to the typed attributes named after it. A jx.Raw field the model
// DOES carry directly (a normalized JSON document, e.g. scope.criteria) is not
// one of these and is left to bodyBinds.
func resolveRawValues(body *Struct, byTF map[string]ModelField) []rawValueBind {
	if body == nil {
		return nil
	}
	var out []rawValueBind
	for _, f := range body.Fields {
		if f.Type != "Raw" {
			continue
		}
		if _, direct := byTF[f.JSONName]; direct {
			continue
		}
		b := rawValueBind{
			SDKField: f.GoName,
			Var:      lowerFirst(f.GoName),
			Func:     "build" + f.GoName,
			Attr:     f.JSONName,
			Required: !strings.HasPrefix(f.Type, "Opt"),
		}
		var named []string
		for _, suffix := range rawValueSuffixes {
			attr := f.JSONName + suffix
			mf, ok := byTF[attr]
			if !ok {
				continue
			}
			named = append(named, attr)
			switch suffix {
			case "_string":
				b.StringGo = mf.GoName
			case "_list":
				b.ListGo = mf.GoName
			case "_map":
				b.MapGo = mf.GoName
			}
		}
		if len(named) == 0 {
			continue // nothing stands in for it; bodyBinds will refuse it loudly
		}
		b.Attrs = joinWithOr(named)
		out = append(out, b)
	}
	slices.SortFunc(out, func(a, b rawValueBind) int { return cmp.Compare(a.SDKField, b.SDKField) })
	return out
}

// rawValueFlat is the read half of a rawValueBind: the response carries the
// same polymorphic jx.Raw, and it has to come back apart into the typed
// attributes or the round trip loses the value on import and refresh.
type rawValueFlat struct {
	Func    string // "flattenDefaultValue"
	SDKPath string // "v.Data.Value.DefaultValue"
	Attr    string // "default_value"

	StringGo string
	ListGo   string
	MapGo    string
}

// resolveRawValueFlats pairs each read-payload jx.Raw field with the binds
// already derived for the request body, so the two halves cannot disagree
// about which attributes stand in for it.
func resolveRawValueFlats(fields []Field, byTF map[string]ModelField, prefix string, binds []rawValueBind) []rawValueFlat {
	byAttr := map[string]rawValueBind{}
	for _, b := range binds {
		byAttr[b.Attr] = b
	}
	var out []rawValueFlat
	for _, f := range fields {
		if f.Type != "Raw" {
			continue
		}
		if _, direct := byTF[f.JSONName]; direct {
			continue
		}
		b, ok := byAttr[f.JSONName]
		if !ok {
			continue
		}
		out = append(out, rawValueFlat{
			Func:     "flatten" + f.GoName,
			SDKPath:  prefix + f.GoName,
			Attr:     f.JSONName,
			StringGo: b.StringGo, ListGo: b.ListGo, MapGo: b.MapGo,
		})
	}
	slices.SortFunc(out, func(a, b rawValueFlat) int { return cmp.Compare(a.Func, b.Func) })
	return out
}

// joinWithOr renders a list for a diagnostic: "a", "a or b", "a, b, or c".
func joinWithOr(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	}
	return strings.Join(items[:len(items)-1], ", ") + ", or " + items[len(items)-1]
}

// mergeRawValues unions the create and update binds, so the helper each one
// calls is emitted once per resource.
func mergeRawValues(sets ...[]rawValueBind) []rawValueBind {
	seen := map[string]bool{}
	var out []rawValueBind
	for _, set := range sets {
		for _, b := range set {
			if seen[b.Func] {
				continue
			}
			seen[b.Func] = true
			out = append(out, b)
		}
	}
	slices.SortFunc(out, func(a, b rawValueBind) int { return cmp.Compare(a.Func, b.Func) })
	return out
}
