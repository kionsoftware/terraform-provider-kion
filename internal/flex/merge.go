package flex

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// MergeKind says how a MergeField's value is written into the JSON body.
type MergeKind string

// The MergeKind values.
const (
	MergeString   MergeKind = "string"
	MergeInt      MergeKind = "int"
	MergeBool     MergeKind = "bool"
	MergeDatecode MergeKind = "datecode" // "YYYY-MM" in the schema, YYYYMM on the wire
	MergeEnum     MergeKind = "enum"     // a string looked up (case-insensitively) in Enum
)

// MergeField is one model value written into a read-modify-write body.
type MergeField struct {
	Value attr.Value     // the model attribute
	Sub   []string       // sub-attribute path within an object Value, if any
	To    string         // dotted JSON path under the record
	Kind  MergeKind      //
	Enum  map[string]any // MergeEnum only: lower-case value -> wire value
}

var datecodeRE = regexp.MustCompile(`^(\d{4})-(\d{2})$`)

// MergeJSON overlays fields onto the record inside a {"data": …} read response
// and returns the record, for endpoints whose update overwrites every column.
// A null or unknown value leaves the read value in place.
func MergeJSON(ctx context.Context, current []byte, fields []MergeField) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	dec := json.NewDecoder(strings.NewReader(string(current)))
	dec.UseNumber()
	if err := dec.Decode(&envelope); err != nil {
		diags.AddError("Decoding current record", err.Error())
		return nil, diags
	}
	if envelope.Data == nil {
		diags.AddError("Decoding current record", "response carries no data object")
		return nil, diags
	}
	for _, f := range fields {
		v, err := mergeLeaf(ctx, f.Value, f.Sub)
		if err != nil {
			diags.AddError(fmt.Sprintf("Building update body (%s)", f.To), err.Error())
			continue
		}
		if v == nil || v.IsNull() || v.IsUnknown() {
			continue
		}
		wire, err := mergeWire(v, f)
		if err != nil {
			diags.AddError(fmt.Sprintf("Building update body (%s)", f.To), err.Error())
			continue
		}
		if err := setPath(envelope.Data, f.To, wire); err != nil {
			diags.AddError(fmt.Sprintf("Building update body (%s)", f.To), err.Error())
		}
	}
	if diags.HasError() {
		return nil, diags
	}
	body, err := json.Marshal(envelope.Data)
	if err != nil {
		diags.AddError("Encoding update body", err.Error())
	}
	return body, diags
}

// mergeLeaf walks sub through nested object values.
func mergeLeaf(ctx context.Context, v attr.Value, sub []string) (attr.Value, error) {
	for _, name := range sub {
		if v == nil || v.IsNull() || v.IsUnknown() {
			return nil, nil
		}
		ov, ok := v.(basetypes.ObjectValuable)
		if !ok {
			return nil, fmt.Errorf("%q: %T is not an object", name, v)
		}
		obj, d := ov.ToObjectValue(ctx)
		if d.HasError() {
			return nil, fmt.Errorf("%q: converting to object", name)
		}
		next, ok := obj.Attributes()[name]
		if !ok {
			return nil, fmt.Errorf("object has no attribute %q", name)
		}
		v = next
	}
	return v, nil
}

func mergeWire(v attr.Value, f MergeField) (any, error) {
	switch f.Kind {
	case MergeString, MergeDatecode, MergeEnum:
		sv, ok := v.(basetypes.StringValuable)
		if !ok {
			return nil, fmt.Errorf("%T is not a string", v)
		}
		s, d := sv.ToStringValue(context.Background())
		if d.HasError() {
			return nil, fmt.Errorf("converting to string")
		}
		str := s.ValueString()
		switch f.Kind {
		case MergeDatecode:
			m := datecodeRE.FindStringSubmatch(str)
			if m == nil {
				return nil, fmt.Errorf("%q is not a YYYY-MM date", str)
			}
			return strconv.ParseInt(m[1]+m[2], 10, 64)
		case MergeEnum:
			w, ok := f.Enum[strings.ToLower(str)]
			if !ok {
				return nil, fmt.Errorf("unsupported value %q", str)
			}
			return w, nil
		}
		return str, nil
	case MergeInt:
		iv, ok := v.(basetypes.Int64Valuable)
		if !ok {
			return nil, fmt.Errorf("%T is not an integer", v)
		}
		i, d := iv.ToInt64Value(context.Background())
		if d.HasError() {
			return nil, fmt.Errorf("converting to integer")
		}
		return i.ValueInt64(), nil
	case MergeBool:
		bv, ok := v.(basetypes.BoolValuable)
		if !ok {
			return nil, fmt.Errorf("%T is not a bool", v)
		}
		b, d := bv.ToBoolValue(context.Background())
		if d.HasError() {
			return nil, fmt.Errorf("converting to bool")
		}
		return b.ValueBool(), nil
	}
	return nil, fmt.Errorf("unknown merge kind %q", f.Kind)
}

// ObjectAsSDK converts an untyped object attribute into an SDK struct through
// its JSON form; null attributes are omitted. The SDK structs it serves carry
// only Opt fields of the object's scalar types, so decoding a value the schema
// accepts does not fail.
func ObjectAsSDK[T any, PT interface {
	*T
	json.Unmarshaler
}](v basetypes.ObjectValue) T {
	var out T
	if v.IsNull() || v.IsUnknown() {
		return out
	}
	m := map[string]any{}
	for k, a := range v.Attributes() {
		if a.IsNull() || a.IsUnknown() {
			continue
		}
		switch x := a.(type) {
		case basetypes.StringValue:
			m[k] = x.ValueString()
		case basetypes.Int64Value:
			m[k] = x.ValueInt64()
		case basetypes.BoolValue:
			m[k] = x.ValueBool()
		case basetypes.Float64Value:
			m[k] = x.ValueFloat64()
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return out
	}
	if err := PT(&out).UnmarshalJSON(b); err != nil {
		var zero T
		return zero
	}
	return out
}

// setPath assigns val at a dotted path, creating (or replacing a null with)
// intermediate objects.
func setPath(m map[string]any, dotted string, val any) error {
	segs := strings.Split(dotted, ".")
	cur := m
	for _, s := range segs[:len(segs)-1] {
		switch next := cur[s].(type) {
		case map[string]any:
			cur = next
		case nil:
			n := map[string]any{}
			cur[s] = n
			cur = n
		default:
			return fmt.Errorf("%q is not an object", s)
		}
	}
	cur[segs[len(segs)-1]] = val
	return nil
}
