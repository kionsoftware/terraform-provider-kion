package crud

import "testing"

func TestExpandConverter(t *testing.T) {
	cases := []struct {
		in   Field
		want string
	}{
		{Field{Type: "string"}, "flex.StringValueFromFramework"},
		{Field{Type: "OptString", Optional: true}, "flex.OptStringFromFramework"},
		{Field{Type: "OptUint64", Optional: true}, "flex.OptUint64FromFramework"},
		{Field{Type: "OptInt64", Optional: true}, "flex.OptInt64FromFramework"},
		{Field{Type: "int64"}, "flex.Int64ValueFromFramework"},
		{Field{Type: "uint64"}, "flex.Uint64FromFramework"},
		{Field{Type: "bool"}, "flex.BoolValueFromFramework"},
		{Field{Type: "OptBool", Optional: true}, "flex.OptBoolFromFramework"},
	}
	for _, c := range cases {
		got, ok := expandConverter(c.in)
		if !ok || got != c.want {
			t.Errorf("expandConverter(%s) = %q,%v want %q", c.in.Type, got, ok, c.want)
		}
	}
	if _, ok := expandConverter(Field{Type: "SomeNestedObject"}); ok {
		t.Error("unknown type must return ok=false (needs crud_override)")
	}
}

func TestFlattenConverter(t *testing.T) {
	cases := []struct {
		in   Field
		want string
	}{
		{Field{Type: "OptString", Optional: true}, "flex.OptStringToFramework"},
		{Field{Type: "string"}, "flex.StringToFramework"},
		{Field{Type: "OptInt64", Optional: true}, "flex.OptInt64ToFramework"},
		{Field{Type: "OptUint64", Optional: true}, "flex.OptUint64ToFramework"},
		{Field{Type: "int64"}, "flex.Int64ToFramework"},
		{Field{Type: "uint64"}, "flex.Uint64ToFramework"},
		{Field{Type: "OptBool", Optional: true}, "flex.OptBoolToFramework"},
		{Field{Type: "bool"}, "flex.BoolToFramework"},
	}
	for _, c := range cases {
		got, ok := flattenConverter(c.in)
		if !ok || got != c.want {
			t.Errorf("flattenConverter(%s) = %q,%v want %q", c.in.Type, got, ok, c.want)
		}
	}
	if _, ok := flattenConverter(Field{Type: "SomeNestedObject"}); ok {
		t.Error("unknown type must return ok=false")
	}
}

func TestFloatConverters(t *testing.T) {
	if got, ok := expandConverter(Field{Type: "float64"}); !ok || got != "flex.Float64ValueFromFramework" {
		t.Errorf("expand float64 = %q,%v", got, ok)
	}
	if got, ok := flattenConverter(Field{Type: "float64"}); !ok || got != "flex.Float64ToFramework" {
		t.Errorf("flatten float64 = %q,%v", got, ok)
	}
}

func TestSliceConverter(t *testing.T) {
	e, f, w, ok := sliceConverter("[]uint64", "types.List")
	if !ok || e != "flex.Uint64SliceFromFramework" || f != "flex.Uint64SliceToFramework" || w != "" {
		t.Errorf("[]uint64 = %q,%q,%q,%v", e, f, w, ok)
	}
	e, f, w, ok = sliceConverter("OptNilUint64Array", "types.List")
	if !ok || e != "flex.Uint64SliceFromFramework" || w != "OptNilUint64Array" {
		t.Errorf("OptNilUint64Array = %q,%q,%q,%v", e, f, w, ok)
	}
	if _, _, _, ok := sliceConverter("[]Foo", "types.List"); ok {
		t.Error("unknown slice elem must return ok=false")
	}
}
