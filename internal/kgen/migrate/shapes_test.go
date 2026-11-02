package migrate

import (
	"strings"
	"testing"
)

// The generated upgraders hand each old attribute's raw state JSON to
// tftypes.ValueFromJSON against the NEW schema type. So the only question that
// decides whether a customer's `terraform plan` survives the upgrade is: does
// the JSON *shape* the old type produced still decode as the new type? These two
// tests answer it for every attribute of every migrated resource, from the
// snapshots, so the class of bug that shipped cft.tags (a map(string) passed
// straight into a list of {tag_key, tag_value}) cannot come back anywhere.

// jsonShape is the shape an attribute's value takes in state JSON. Two
// attributes with different shapes cannot decode as one another, full stop.
type jsonShape string

const (
	shapeArray  jsonShape = "array"  // '[', list/set/tuple, and SDKv2 list/set blocks
	shapeObject jsonShape = "object" // '{', map/object, single-nested
	shapeString jsonShape = "string"
	shapeNumber jsonShape = "number"
	shapeBool   jsonShape = "bool"
	shapeOther  jsonShape = "other"
)

func shapeOf(a Attr) jsonShape {
	if a.Kind == "block" {
		if a.Nesting == "list" || a.Nesting == "set" {
			return shapeArray
		}
		return shapeObject
	}
	if a.NestedObj {
		if a.NestedMode == "list" || a.NestedMode == "set" {
			return shapeArray
		}
		return shapeObject
	}
	switch {
	case a.TypeJSON == `"string"`:
		return shapeString
	case a.TypeJSON == `"number"`:
		return shapeNumber
	case a.TypeJSON == `"bool"`:
		return shapeBool
	case strings.HasPrefix(a.TypeJSON, `["list"`),
		strings.HasPrefix(a.TypeJSON, `["set"`),
		strings.HasPrefix(a.TypeJSON, `["tuple"`):
		return shapeArray
	case strings.HasPrefix(a.TypeJSON, `["map"`),
		strings.HasPrefix(a.TypeJSON, `["object"`):
		return shapeObject
	}
	return shapeOther
}

// migratedPairs yields (oldType, newType) for every resource with a state
// upgrader, resolving the alias indirection (old_type).
func migratedPairs(t *testing.T) (map[string]string, map[string]Resource, map[string]Resource, map[string]Transform) {
	t.Helper()
	oldS, err := LoadSchema(oldSnap)
	if err != nil {
		t.Fatal(err)
	}
	newS, err := LoadSchema(newSnap)
	if err != nil {
		t.Fatal(err)
	}
	ups, err := LoadUpgrades("../../../codegen/state_upgrades.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pairs := map[string]string{} // oldType -> newType
	for newType, tr := range ups {
		oldType := newType
		if tr.OldType != "" {
			oldType = tr.OldType
		}
		if _, ok := oldS[oldType]; !ok {
			continue
		}
		if _, ok := newS[newType]; !ok {
			continue
		}
		pairs[oldType] = newType
	}
	return pairs, oldS, newS, ups
}

// handles reports whether the transform restructures this new attribute rather
// than passing the old value through. I.e. whether a rule exists that can fix a
// shape mismatch.
func handles(tr Transform, attr string) (string, bool) {
	if pr, ok := tr.Project[attr]; ok && pr.From != "" {
		return "project", true
	}
	if kv, ok := tr.KVList[attr]; ok && kv.KeyField != "" {
		return "kv_list", true
	}
	if contains(tr.Unwrap, attr) {
		return "unwrap", true
	}
	if attr == "id" && tr.IDInt {
		return "id_int", true
	}
	return "", false
}

// TestUpgradeShapes_allIncompatiblesHandled is the backstop for the whole
// migration surface: every attribute whose JSON shape changed between the old
// and new schema MUST be covered by a state_upgrades rule. An uncovered one is
// passed through verbatim by the generated upgrader and blows up in
// ValueFromJSON on a real customer's state, which is exactly what
// kion_aws_cloudformation_template.tags did.
func TestUpgradeShapes_allIncompatiblesHandled(t *testing.T) {
	pairs, oldS, newS, ups := migratedPairs(t)

	// The complete, deliberate inventory of TOP-LEVEL shape changes on the
	// migrated surface, the ones where the old JSON's very first byte no longer
	// matches what the new type wants, each with the rule that restructures it.
	// Anything not on this list is a new shape change that needs a rule (or a
	// considered addition here); anything on it that no longer mismatches is
	// stale. This list is short on purpose: it is the whole set of ways a
	// customer's upgrade can hard-fail, and it should stay reviewable.
	want := map[string]string{
		"kion_aws_account.aws_organizational_unit":        "unwrap",  // block set → single nested
		"kion_aws_account.move_project_settings":          "unwrap",  // block set → single nested
		"kion_aws_account.id":                             "id_int",  // string → number
		"kion_aws_cloudformation_template.tags":           "kv_list", // map(string) → list of {tag_key, tag_value}
		"kion_ou_cloud_access_role.aws_session_tags":      "kv_list", // map(string) → list of {tag_key, tag_value, computed ids}
		"kion_project_cloud_access_role.aws_session_tags": "kv_list",
	}
	got := map[string]string{}

	for oldType, newType := range pairs {
		tr := ups[newType]
		for name, oa := range oldS[oldType].Attrs {
			na, ok := newS[newType].Attrs[name]
			if !ok {
				// Gone from the new schema (renamed or dropped): the upgrader
				// never emits this key, so no shape question arises. Renames are
				// type-checked by TestUpgradeShapes_renamesKeepType.
				continue
			}
			so, sn := shapeOf(oa), shapeOf(na)
			key := oldType + "." + name
			rule, covered := handles(tr, name)

			if so != sn {
				if !covered {
					t.Errorf("%s: %s → %s but no state_upgrades rule restructures it; "+
						"the generated upgrader passes the old value through and ValueFromJSON will reject it",
						key, so, sn)
					continue
				}
				got[key] = rule
				if wantRule, listed := want[key]; !listed {
					t.Errorf("%s: %s → %s handled by %q but not in the reviewed inventory; "+
						"add it once you have confirmed the rule is right", key, so, sn, rule)
				} else if wantRule != rule {
					t.Errorf("%s: handled by %q, inventory says %q", key, rule, wantRule)
				}
				continue
			}

			// Same top-level shape, but a block's array holds objects while a
			// plain collection attribute holds scalars, an ELEMENT-level
			// mismatch that decodes no better. These are the ubiquitous
			// ownership projections (owner_users { id } → owner_user_ids = [id]),
			// so they need a rule but not an inventory entry.
			if so == shapeArray && oa.Kind == "block" && na.Kind == "attr" && !na.NestedObj && !covered {
				t.Errorf("%s: block of objects → %s of scalars but no state_upgrades rule projects it",
					key, na.TypeJSON)
			}
		}
	}
	for key := range want {
		if _, ok := got[key]; !ok {
			t.Errorf("%s: listed as a shape change but the snapshots no longer show one; stale entry", key)
		}
	}
}

// TestUpgradeShapes_passthroughObjectsMatchFieldForField covers the subtler half:
// an old block-set and a new list/set-nested attribute have the SAME json shape
// (both arrays of objects), so the upgrader passes them straight through, which
// is only correct while the object's fields are identical. tftypes rejects both a
// missing and an extra object key, so one renamed field inside
// kion_project.project_funding would break every project's upgrade silently
// until a customer hit it.
func TestUpgradeShapes_passthroughObjectsMatchFieldForField(t *testing.T) {
	pairs, oldS, newS, ups := migratedPairs(t)

	checked := 0
	for oldType, newType := range pairs {
		tr := ups[newType]
		for name, oa := range oldS[oldType].Attrs {
			na, ok := newS[newType].Attrs[name]
			if !ok {
				continue
			}
			// Only passthroughs matter: a rule-covered attribute is restructured.
			if _, covered := handles(tr, name); covered {
				continue
			}
			// Old block (array/object of objects) → new nested-object attribute.
			if oa.Kind != "block" || !na.NestedObj {
				continue
			}
			checked++
			// Fields only the new object has are padded with null by the
			// generated upgrader; an old field the new object lacks is not.
			for _, f := range oa.NestedAttrs {
				if !contains(na.NestedAttrs, f) {
					t.Errorf("%s.%s: passed through but old field %q is gone from the new object "+
						"(old %v, new %v); tftypes rejects the extra key",
						oldType, name, f, oa.NestedAttrs, na.NestedAttrs)
				}
			}
		}
	}
	// kion_project's budget, move_ou_settings and project_funding are the three.
	if checked != 3 {
		t.Errorf("checked %d block→nested-object passthroughs, expected 3 "+
			"(kion_project budget/move_ou_settings/project_funding): "+
			"update this count deliberately if the surface changed", checked)
	}
}

// TestUpgradeShapes_renamesKeepType checks the rename rules, which carry the old
// value over verbatim under a new key: the two types must be identical, or the
// renamed value cannot decode.
func TestUpgradeShapes_renamesKeepType(t *testing.T) {
	pairs, oldS, newS, ups := migratedPairs(t)
	for oldType, newType := range pairs {
		for oldA, newA := range ups[newType].Rename {
			oa, ok := oldS[oldType].Attrs[oldA]
			if !ok {
				t.Errorf("%s: rename source %q is not an old attribute", newType, oldA)
				continue
			}
			na, ok := newS[newType].Attrs[newA]
			if !ok {
				t.Errorf("%s: rename target %q is not a new attribute", newType, newA)
				continue
			}
			if shapeOf(oa) != shapeOf(na) || oa.TypeJSON != na.TypeJSON {
				t.Errorf("%s: rename %s → %s changes type (%s → %s); a rename passes the value through unchanged",
					newType, oldA, newA, oa.TypeJSON, na.TypeJSON)
			}
		}
	}
}
