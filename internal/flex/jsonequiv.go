package flex

import (
	"encoding/json"
	"reflect"
)

// JSONEquivalent reports whether two JSON documents differ only by keys the
// other side omits and whose value is empty.
//
// It exists for attributes a Kion endpoint stores in a canonical form of its
// own. kion_azure_role is the case: role_permissions is parsed into Azure's
// Permissions struct and re-marshaled, so a configured
//
//	{"actions":["…"],"notActions":[]}
//
// is read back as
//
//	{"actions":["…"],"dataActions":[],"notActions":[],"notDataActions":[]}
//
// Those describe the same role. Compared as strings they are a diff that can
// never converge: every plan proposes rewriting the value to what the
// practitioner wrote, and every apply is answered with the expanded form again.
//
// The rule is deliberately narrow. Only an ADDED key holding an empty value is
// ignored -- an empty array, empty object, or null. A key present on both sides
// with different contents is a difference, a key holding a non-empty value the
// other lacks is a difference, and anything that is not valid JSON is compared
// verbatim. So a genuine change is still a change: this cannot mask drift, only
// the API filling in its own defaults.
func JSONEquivalent(a, b string) bool {
	if a == b {
		return true
	}
	var av, bv any
	if json.Unmarshal([]byte(a), &av) != nil || json.Unmarshal([]byte(b), &bv) != nil {
		// Not both JSON: the only honest comparison left is the one above.
		return false
	}
	return equivalent(av, bv)
}

func equivalent(a, b any) bool {
	am, aok := a.(map[string]any)
	bm, bok := b.(map[string]any)
	if aok && bok {
		for k, av := range am {
			bv, present := bm[k]
			if !present {
				if isEmptyJSON(av) {
					continue
				}
				return false
			}
			if !equivalent(av, bv) {
				return false
			}
		}
		for k, bv := range bm {
			if _, present := am[k]; !present && !isEmptyJSON(bv) {
				return false
			}
		}
		return true
	}

	as, aok := a.([]any)
	bs, bok := b.([]any)
	if aok && bok {
		// Order is meaningful in a JSON array and in every Kion payload that
		// reaches here, so this does not sort.
		if len(as) != len(bs) {
			return false
		}
		for i := range as {
			if !equivalent(as[i], bs[i]) {
				return false
			}
		}
		return true
	}

	return reflect.DeepEqual(a, b)
}

// isEmptyJSON reports whether a decoded value carries no information, and so
// may be absent on the other side without that being a difference.
func isEmptyJSON(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case []any:
		return len(t) == 0
	case map[string]any:
		return len(t) == 0
	default:
		return false
	}
}
