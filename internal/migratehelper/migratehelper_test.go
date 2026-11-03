package migratehelper

import (
	"encoding/json"
	"testing"
)

func eq(t *testing.T, got json.RawMessage, want string) {
	t.Helper()
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestOrNull(t *testing.T) {
	eq(t, OrNull(nil), "null")
	eq(t, OrNull(json.RawMessage("")), "null")
	eq(t, OrNull(json.RawMessage(`"x"`)), `"x"`)
	eq(t, OrNull(json.RawMessage(`42`)), `42`)
	eq(t, OrNull(json.RawMessage(`false`)), `false`)
	// A JSON null passes through as null (already the right shape).
	eq(t, OrNull(json.RawMessage(`null`)), `null`)
}

func TestProjectIDs(t *testing.T) {
	// The core block-set → id-list transform.
	eq(t, ProjectIDs(json.RawMessage(`[{"id":5},{"id":6}]`), "id"), `[5,6]`)
	// Single element.
	eq(t, ProjectIDs(json.RawMessage(`[{"id":9}]`), "id"), `[9]`)
	// Empty list projects to an empty list, NOT null, an explicitly-empty
	// membership set must survive as [] so it round-trips.
	eq(t, ProjectIDs(json.RawMessage(`[]`), "id"), `[]`)
	// Absent / null source → null (added attr the provider will repopulate).
	eq(t, ProjectIDs(nil, "id"), `null`)
	eq(t, ProjectIDs(json.RawMessage(`null`), "id"), `null`)
	// A different field name.
	eq(t, ProjectIDs(json.RawMessage(`[{"account_id":1},{"account_id":2}]`), "account_id"), `[1,2]`)
	// Objects missing the field are skipped, not errored.
	eq(t, ProjectIDs(json.RawMessage(`[{"id":1},{"other":2}]`), "id"), `[1]`)
	// Malformed JSON degrades to null rather than panicking.
	eq(t, ProjectIDs(json.RawMessage(`{not json`), "id"), `null`)
}

func TestProjectIDs_preservesOrder(t *testing.T) {
	got := ProjectIDs(json.RawMessage(`[{"id":3},{"id":1},{"id":2}]`), "id")
	eq(t, got, `[3,1,2]`)
}

func TestUnwrap(t *testing.T) {
	// [{…}] → {…} for block → single-nested-attribute.
	eq(t, Unwrap(json.RawMessage(`[{"a":1}]`)), `{"a":1}`)
	// Empty list → null.
	eq(t, Unwrap(json.RawMessage(`[]`)), `null`)
	// Absent / null → null.
	eq(t, Unwrap(nil), `null`)
	eq(t, Unwrap(json.RawMessage(`null`)), `null`)
	// More than one element: takes the first (old single-block sets hold ≤1).
	eq(t, Unwrap(json.RawMessage(`[{"a":1},{"a":2}]`)), `{"a":1}`)
	// Malformed → null.
	eq(t, Unwrap(json.RawMessage(`nope`)), `null`)
}

func TestStringToNum(t *testing.T) {
	// The aws_account id string→number case.
	eq(t, StringToNum(json.RawMessage(`"42"`)), `42`)
	// Already a number passes through.
	eq(t, StringToNum(json.RawMessage(`42`)), `42`)
	// Empty string → null (not 0. The value was effectively unset).
	eq(t, StringToNum(json.RawMessage(`""`)), `null`)
	// Absent / null → null.
	eq(t, StringToNum(nil), `null`)
	eq(t, StringToNum(json.RawMessage(`null`)), `null`)
	// Large numeric id keeps full precision (no float rounding).
	eq(t, StringToNum(json.RawMessage(`"900719925474099"`)), `900719925474099`)
}

// TestStringToNum_roundTripsThroughJSON guards against precision loss: the
// output must be valid JSON that unmarshals back to the same integer.
func TestStringToNum_roundTripsThroughJSON(t *testing.T) {
	out := StringToNum(json.RawMessage(`"12345678901234"`))
	var n json.Number
	if err := json.Unmarshal(out, &n); err != nil {
		t.Fatalf("output not valid JSON number: %v", err)
	}
	if n.String() != "12345678901234" {
		t.Errorf("round-trip = %s, want 12345678901234", n.String())
	}
}

func TestMapToObjectList(t *testing.T) {
	// The cft tags transform: map(string) → list of {tag_key, tag_value}.
	eq(t, MapToObjectList(json.RawMessage(`{"env":"prod"}`), "tag_key", "tag_value"),
		`[{"tag_key":"env","tag_value":"prod"}]`)
	// Multiple entries come out in sorted key order, deterministically.
	eq(t, MapToObjectList(json.RawMessage(`{"team":"platform","env":"prod"}`), "tag_key", "tag_value"),
		`[{"tag_key":"env","tag_value":"prod"},{"tag_key":"team","tag_value":"platform"}]`)
	// An empty map explodes to an empty list, NOT null, an explicitly-empty tag
	// set must survive as [] so it round-trips.
	eq(t, MapToObjectList(json.RawMessage(`{}`), "tag_key", "tag_value"), `[]`)
	// Absent / null → null (the provider's next Read repopulates).
	eq(t, MapToObjectList(nil, "tag_key", "tag_value"), `null`)
	eq(t, MapToObjectList(json.RawMessage(`null`), "tag_key", "tag_value"), `null`)
	// Non-string values keep their JSON type.
	eq(t, MapToObjectList(json.RawMessage(`{"n":42}`), "k", "v"), `[{"k":"n","v":42}]`)
	// Keys needing escaping are quoted correctly.
	eq(t, MapToObjectList(json.RawMessage(`{"a\"b":"x"}`), "k", "v"), `[{"k":"a\"b","v":"x"}]`)
	// Malformed JSON degrades to null rather than panicking.
	eq(t, MapToObjectList(json.RawMessage(`{not json`), "k", "v"), `null`)
}

// TestMapToObjectList_idempotent: an already-exploded list passes through
// untouched, so re-running the upgrader over upgraded state is a no-op rather
// than a double transform.
func TestMapToObjectList_idempotent(t *testing.T) {
	already := `[{"tag_key":"env","tag_value":"prod"}]`
	eq(t, MapToObjectList(json.RawMessage(already), "tag_key", "tag_value"), already)
	// Leading whitespace does not defeat the detection.
	eq(t, MapToObjectList(json.RawMessage("\n  "+already), "tag_key", "tag_value"), "\n  "+already)
}

// TestMapToObjectList_sortedOrderMatchesOldState: cty serializes a map's keys in
// sorted order, so sorted output is the order the old state itself recorded,
// the upgrade introduces no spurious reordering.
func TestMapToObjectList_sortedOrderMatchesOldState(t *testing.T) {
	// What Terraform would have written for map(string) tags.
	oldState, err := json.Marshal(map[string]string{"zeta": "3", "alpha": "1", "mid": "2"})
	if err != nil {
		t.Fatal(err)
	}
	got := MapToObjectList(oldState, "tag_key", "tag_value")
	var objs []map[string]string
	if err := json.Unmarshal(got, &objs); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	for i, w := range []string{"alpha", "mid", "zeta"} {
		if objs[i]["tag_key"] != w {
			t.Errorf("element %d tag_key = %q, want %q", i, objs[i]["tag_key"], w)
		}
	}
}

func TestPadObjectKeys(t *testing.T) {
	eq(t, PadObjectKeys(nil, "c"), "null")
	eq(t, PadObjectKeys(json.RawMessage(`null`), "c"), "null")
	eq(t, PadObjectKeys(json.RawMessage(`[{"a":"x","b":"y"}]`), "c"), `[{"a":"x","b":"y","c":null}]`)
	// An existing key is kept, so re-running over upgraded state is a no-op.
	eq(t, PadObjectKeys(json.RawMessage(`[{"a":"x","c":"z"}]`), "c"), `[{"a":"x","c":"z"}]`)
	eq(t, PadObjectKeys(json.RawMessage(`[]`), "c"), `[]`)
}
