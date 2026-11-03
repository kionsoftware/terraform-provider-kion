package crud

import (
	"fmt"
	"sort"
	"strings"
)

// writeShapeField maps one model value into a read-modify-write update body.
//
// Some private updates overwrite every column of the record from the body
// (PUT /v1/payer/{id} is the case), so a body built from the model alone blanks
// whatever the model does not carry. Declaring `write_shape` turns the raw
// update into: read the record, overlay these fields, send it back. Unmodelled
// columns survive because they are echoed from the read.
type writeShapeField struct {
	TF     string         `yaml:"tf"`     // model attr, or attr.sub[.sub] into an object Value
	To     string         `yaml:"to"`     // dotted json path under the record
	Kind   string         `yaml:"kind"`   // string | int | bool | datecode | enum
	Values map[string]any `yaml:"values"` // enum only: lower-case value -> wire value
}

var mergeKinds = map[string]string{
	"string":   "flex.MergeString",
	"int":      "flex.MergeInt",
	"bool":     "flex.MergeBool",
	"datecode": "flex.MergeDatecode",
	"enum":     "flex.MergeEnum",
}

// buildMergeFields renders the []flex.MergeField elements for a write shape.
func buildMergeFields(fields []writeShapeField, byTF map[string]ModelField) (string, error) {
	var b strings.Builder
	seen := map[string]bool{}
	for _, f := range fields {
		kind, ok := mergeKinds[f.Kind]
		if !ok {
			return "", fmt.Errorf("write_shape %q: unknown kind %q", f.TF, f.Kind)
		}
		if f.To == "" {
			return "", fmt.Errorf("write_shape %q: no `to` path", f.TF)
		}
		if seen[f.To] {
			return "", fmt.Errorf("write_shape: %q is written twice", f.To)
		}
		seen[f.To] = true
		segs := strings.Split(f.TF, ".")
		mf, ok := byTF[segs[0]]
		if !ok {
			return "", fmt.Errorf("write_shape %q: not in model", f.TF)
		}
		if mf.TFSDK == "id" {
			return "", fmt.Errorf("write_shape %q: the id is echoed from the read, not written", f.TF)
		}
		if len(segs) > 1 && !strings.HasSuffix(mf.Type, "Value") {
			return "", fmt.Errorf("write_shape %q: %s is %s, not an object", f.TF, segs[0], mf.Type)
		}
		fmt.Fprintf(&b, "{Value: plan.%s, ", mf.GoName)
		if len(segs) > 1 {
			quoted := make([]string, len(segs)-1)
			for i, s := range segs[1:] {
				quoted[i] = fmt.Sprintf("%q", s)
			}
			fmt.Fprintf(&b, "Sub: []string{%s}, ", strings.Join(quoted, ", "))
		}
		fmt.Fprintf(&b, "To: %q, Kind: %s", f.To, kind)
		if f.Kind == "enum" {
			if len(f.Values) == 0 {
				return "", fmt.Errorf("write_shape %q: enum needs values", f.TF)
			}
			lit, err := enumLiteral(f.Values)
			if err != nil {
				return "", fmt.Errorf("write_shape %q: %w", f.TF, err)
			}
			fmt.Fprintf(&b, ", Enum: %s", lit)
		} else if len(f.Values) > 0 {
			return "", fmt.Errorf("write_shape %q: values only apply to kind enum", f.TF)
		}
		b.WriteString("},\n")
	}
	return b.String(), nil
}

func enumLiteral(values map[string]any) (string, error) {
	keys := make([]string, 0, len(values))
	for k := range values {
		if k != strings.ToLower(k) {
			return "", fmt.Errorf("enum key %q must be lower case (lookup is case-insensitive)", k)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		var v string
		switch x := values[k].(type) {
		case int, int64, bool:
			v = fmt.Sprint(x)
		case string:
			v = fmt.Sprintf("%q", x)
		default:
			return "", fmt.Errorf("enum value for %q has unsupported type %T", k, x)
		}
		parts[i] = fmt.Sprintf("%q: %s", k, v)
	}
	return "map[string]any{" + strings.Join(parts, ", ") + "}", nil
}

// objectSubExpr expands a plain SDK struct field whose model sub-attribute is
// an untyped object (a struct nested inside an already-nested object).
func objectSubExpr(sf Field, vf ModelField, prefix string) (string, bool) {
	if vf.Type != "basetypes.ObjectValue" || sf.Type == "" {
		return "", false
	}
	for _, p := range []string{"Opt", "Nil", "[]", "*"} {
		if strings.HasPrefix(sf.Type, p) {
			return "", false
		}
	}
	return "flex.ObjectAsSDK[generated." + sf.Type + "](" + prefix + "." + vf.GoName + ")", true
}

// kindModelType is the framework type a read_shape kind flattens to.
func kindModelType(kind string) string {
	switch kind {
	case "id", "string", "int_string", "datecode", "null_string", "opt_string":
		return "types.String"
	case "int", "null_int", "opt_int":
		return "types.Int64"
	case "bool":
		return "types.Bool"
	case "float":
		return "types.Float64"
	}
	return ""
}

// checkReadShapeKinds fails when a read_shape leaf would flatten into a model
// attribute of a different type. That compiles, and then every read fails in
// the Value constructor. subTypes maps a Value type to its sub-field types.
func checkReadShapeKinds(s readShape, byTF map[string]ModelField, subTypes map[string]map[string]string) error {
	check := func(where, kind, modelType string) error {
		want := kindModelType(kind)
		if want == "" || modelType == "" {
			return nil // unknown kinds are refused elsewhere; untyped models are not checked
		}
		if got := frameworkModelType(modelType); got != want {
			return fmt.Errorf("read_shape %s: kind %q flattens to %s but the model attribute is %s", where, kind, want, got)
		}
		return nil
	}
	for _, sc := range s.Scalars {
		if sc.WriteOnly {
			continue
		}
		if err := check(sc.TF, sc.Kind, byTF[sc.TF].Type); err != nil {
			return err
		}
	}
	for _, o := range s.Objects {
		for _, sub := range o.Subs {
			if sub.WriteOnly {
				continue
			}
			if err := check(o.TF+"."+sub.TF, sub.Kind, subTypes[o.ValueType][sub.TF]); err != nil {
				return err
			}
		}
	}
	return nil
}
