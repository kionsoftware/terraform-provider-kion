package kalign

import "testing"

func TestTFFamily(t *testing.T) {
	cases := map[string]string{
		"types.String":  "string",
		"types.Bool":    "bool",
		"types.Int64":   "int",
		"types.Number":  "int",
		"types.Float64": "float",
		"types.List":    "",
		"CriteriaValue": "",
	}
	for in, want := range cases {
		if got := tfFamily(in); got != want {
			t.Errorf("tfFamily(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSDKFamily(t *testing.T) {
	cases := map[string]string{
		"string":       "string",
		"OptString":    "string",
		"OptNilString": "string",
		"NilString":    "string",
		"bool":         "bool",
		"OptNilBool":   "bool",
		"uint64":       "int",
		"OptUint64":    "int",
		"OptNilUint64": "int",
		"NilUint64":    "int",
		"OptNullUint":  "int",
		"*Nested":      "nested",
		"OptFloat64":   "float",
		// Terraform carries timestamps as RFC3339 strings.
		"OptNullTime":    "string",
		"OptDateTime":    "string",
		"OptNilDateTime": "string",
	}
	for in, want := range cases {
		if got := sdkFamily(in); got != want {
			t.Errorf("sdkFamily(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTypesCompatible(t *testing.T) {
	compatible := [][2]string{
		{"types.String", "OptString"},
		{"types.Int64", "OptNilUint64"},
		{"types.Bool", "OptNilBool"},
		{"CriteriaValue", "*Nested"}, // nested tf type => always compatible here
	}
	for _, c := range compatible {
		if !typesCompatible(c[0], c[1]) {
			t.Errorf("typesCompatible(%q,%q) = false, want true", c[0], c[1])
		}
	}
	incompatible := [][2]string{
		{"types.String", "OptUint64"},
		{"types.Bool", "OptString"},
		{"types.Int64", "OptString"},
	}
	for _, c := range incompatible {
		if typesCompatible(c[0], c[1]) {
			t.Errorf("typesCompatible(%q,%q) = true, want false", c[0], c[1])
		}
	}
}

func TestBestOverlapType(t *testing.T) {
	sdk := map[string][]SDKField{
		"Wrong":   {{JSON: "x"}, {JSON: "y"}},
		"Right":   {{JSON: "a"}, {JSON: "b"}, {JSON: "c"}},
		"Tie":     {{JSON: "a"}, {JSON: "b"}},
		"AlsoTie": {{JSON: "a"}, {JSON: "b"}},
	}
	set := map[string]bool{"a": true, "b": true, "c": true}
	got, ov := bestOverlapType("x", set, sdk)
	if got != "Right" || ov != 3 {
		t.Fatalf("bestOverlapType = (%q,%d), want (Right,3)", got, ov)
	}

	// no overlap
	got, ov = bestOverlapType("x", map[string]bool{"zz": true}, sdk)
	if got != "" || ov != 0 {
		t.Errorf("no-overlap = (%q,%d), want (\"\",0)", got, ov)
	}
}

// kion_ou tied CustomVariable and OUCreate at 6 fields and took CustomVariable
// on lexical order; a tie goes to the type named after the service.
func TestBestOverlapType_tieBreaksOnServiceName(t *testing.T) {
	sdk := map[string][]SDKField{
		"CustomVariable": {{JSON: "a"}, {JSON: "b"}},
		"OUCreate":       {{JSON: "a"}, {JSON: "b"}},
	}
	set := map[string]bool{"a": true, "b": true}
	if got, _ := bestOverlapType("ou", set, sdk); got != "OUCreate" {
		t.Errorf("tie = %q, want OUCreate", got)
	}
	if got, _ := bestOverlapType("unrelated", set, sdk); got != "CustomVariable" {
		t.Errorf("unnamed tie = %q, want lexical CustomVariable", got)
	}
}

// An ogen enum is a named integer; its Opt wrapper must not read as a mismatch
// against an Int64 attribute.
func TestResolve_namedIntegerIsIntFamily(t *testing.T) {
	model := ServiceModel{
		Service: "gcp", Name: "GcpModel",
		Fields: []ModelField{{GoName: "Stage", TFSDK: "stage", TFType: "types.Int64"}},
	}
	sdk := map[string][]SDKField{
		"Gcp": {{GoName: "Stage", JSON: "stage", GoType: "OptStage", Underlying: "uint64"}},
	}
	r := Resolve(model, sdk, map[string]bool{"OptStageToFramework": true})
	if len(r.TypeMismatch) != 0 {
		t.Errorf("TypeMismatch = %v, want none", r.TypeMismatch)
	}
}

func TestResolve_CleanMatch(t *testing.T) {
	model := ServiceModel{
		Service: "ou_note", Name: "OuNoteModel",
		Fields: []ModelField{
			{GoName: "Id", TFSDK: "id", TFType: "types.Int64"},
			{GoName: "Name", TFSDK: "name", TFType: "types.String"},
		},
	}
	sdk := map[string][]SDKField{
		"OUNote": {
			{GoName: "ID", JSON: "id", GoType: "OptUint64"},
			{GoName: "Name", JSON: "name", GoType: "OptString"},
		},
	}
	flex := map[string]bool{"OptUint64ToFramework": true, "OptStringToFramework": true}

	r := Resolve(model, sdk, flex)
	if r.SDKType != "OUNote" || r.Overlap != 2 || r.LowConfidence {
		t.Fatalf("resolution = %+v", r)
	}
	if r.Findings() != 0 {
		t.Errorf("expected 0 findings, got %d: %+v", r.Findings(), r)
	}
	if len(r.Pairs) != 2 || r.Pairs[0].FlexFn != "OptUint64ToFramework" || !r.Pairs[0].HaveFlex {
		t.Errorf("pairs = %+v", r.Pairs)
	}
}

func TestResolve_DriftAndMissingFlex(t *testing.T) {
	model := ServiceModel{
		Service: "acct", Name: "AcctModel",
		Fields: []ModelField{
			{GoName: "Id", TFSDK: "id", TFType: "types.Int64"},
			{GoName: "Ghost", TFSDK: "ghost", TFType: "types.String"}, // no SDK field
			{GoName: "Weird", TFSDK: "weird", TFType: "types.String"}, // SDK field but no flex converter
		},
	}
	sdk := map[string][]SDKField{
		"Acct": {
			{GoName: "ID", JSON: "id", GoType: "OptUint64"},
			// same primitive family as types.String (no type mismatch), but its
			// converter is absent from flex below -> isolates a missing-flex finding.
			{GoName: "Weird", JSON: "weird", GoType: "OptString"},
		},
	}
	flex := map[string]bool{"OptUint64ToFramework": true} // no OptStringToFramework

	r := Resolve(model, sdk, flex)
	if len(r.MissingInSDK) != 1 || r.MissingInSDK[0] != "ghost" {
		t.Errorf("MissingInSDK = %v", r.MissingInSDK)
	}
	if len(r.MissingFlex) != 1 {
		t.Errorf("MissingFlex = %v", r.MissingFlex)
	}
	if r.Findings() != 2 { // ghost + weird's missing flex
		t.Errorf("Findings = %d, want 2", r.Findings())
	}
}

func TestResolve_TypeMismatchAndNested(t *testing.T) {
	model := ServiceModel{
		Service: "sc", Name: "ScModel",
		Fields: []ModelField{
			{GoName: "Id", TFSDK: "id", TFType: "types.Int64"},
			{GoName: "Bad", TFSDK: "bad", TFType: "types.String"},            // SDK is int => mismatch
			{GoName: "Criteria", TFSDK: "criteria", TFType: "CriteriaValue"}, // nested
		},
	}
	sdk := map[string][]SDKField{
		"Sc": {
			{GoName: "ID", JSON: "id", GoType: "OptUint64"},
			{GoName: "Bad", JSON: "bad", GoType: "OptUint64"},
			{GoName: "Criteria", JSON: "criteria", GoType: "*ScCriteria"},
		},
	}
	flex := map[string]bool{"OptUint64ToFramework": true}

	r := Resolve(model, sdk, flex)
	if len(r.TypeMismatch) != 1 {
		t.Errorf("TypeMismatch = %v", r.TypeMismatch)
	}
	if len(r.NestedAttrs) != 1 || r.NestedAttrs[0] != "criteria" {
		t.Errorf("NestedAttrs = %v", r.NestedAttrs)
	}
	// nested pair carries no flex and is not counted as missing-flex
	var nested Pair
	for _, p := range r.Pairs {
		if p.Model.TFSDK == "criteria" {
			nested = p
		}
	}
	if !nested.Nested || nested.FlexFn != "" {
		t.Errorf("nested pair = %+v", nested)
	}
}

func TestResolve_LowConfidence(t *testing.T) {
	model := ServiceModel{
		Service: "x", Name: "XModel",
		Fields: []ModelField{
			{TFSDK: "a", TFType: "types.String"},
			{TFSDK: "b", TFType: "types.String"},
			{TFSDK: "c", TFType: "types.String"},
			{TFSDK: "d", TFType: "types.String"},
		},
	}
	sdk := map[string][]SDKField{"Y": {{JSON: "a", GoType: "OptString"}}} // only 1/4 overlap
	r := Resolve(model, sdk, map[string]bool{})
	if !r.LowConfidence {
		t.Errorf("expected LowConfidence for 1/4 overlap")
	}
}
