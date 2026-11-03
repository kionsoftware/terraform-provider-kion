// Package migratehelper holds the runtime converters the generated state
// upgraders (<name>_upgrade_gen.go) call. The upgraders transform the old-schema
// (SDKv2) state as a raw JSON map into a new-schema JSON map, then decode it with
// tftypes.ValueFromJSON, so nested objects, sets, maps, and scalars all migrate
// uniformly. These helpers operate at the JSON level. It is hand-written
// infrastructure (like internal/flex), not a resource body.
package migratehelper

import (
	"bytes"
	"encoding/json"
	"sort"
)

// Null is the JSON null literal, for new attributes with no old source (the
// provider's next Read repopulates them).
var Null = json.RawMessage("null")

// OrNull returns raw, or JSON null when the old attribute was absent.
func OrNull(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return Null
	}
	return raw
}

// ProjectIDs maps an old set/list of single-field objects, [{<field>: n}, …],
// to a JSON array of those field values ([n, …]). This is the block-set → id-list
// transform (owner_users:[{id}] → owner_user_ids:[int]). Null/absent → null.
func ProjectIDs(raw json.RawMessage, field string) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return Null
	}
	var elems []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &elems); err != nil {
		return Null
	}
	ids := make([]json.RawMessage, 0, len(elems))
	for _, e := range elems {
		if v, ok := e[field]; ok {
			ids = append(ids, v)
		}
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return Null
	}
	return b
}

// Unwrap turns an old single-element block set/list, [{…}], into a single
// object {…}, for blocks that became a SingleNestedAttribute (e.g. aws_account's
// aws_organizational_unit). Empty/absent → null.
func Unwrap(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return Null
	}
	var elems []json.RawMessage
	if err := json.Unmarshal(raw, &elems); err != nil || len(elems) == 0 {
		return Null
	}
	return elems[0]
}

// MapToObjectList explodes an old map attribute, {"k": v, …}, into the list of
// two-field objects the new schema declares: [{keyField: "k", valField: v}, …].
// This is the map→object-list transform (kion_aws_cloudformation_template's
// `tags = {env = "prod"}` map(string) → cft's `tags` list of
// {tag_key, tag_value}); without it the old map's JSON reaches ValueFromJSON as
// an object where a list is expected and the whole upgrade fails.
//
// Entries are emitted in sorted key order. That is both deterministic (Go map
// iteration is not) and faithful: cty serializes a map's keys sorted, so sorted
// order is the order the old state itself recorded. The provider's next Read
// replaces it with the API's order regardless.
//
// An already-migrated array passes through untouched, so re-running the upgrader
// over upgraded state is a no-op. Null/absent → null.
func MapToObjectList(raw json.RawMessage, keyField, valField string) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return Null
	}
	// Already a list of objects (state upgraded once already): leave it alone.
	if trimmed := bytes.TrimLeft(raw, " \t\r\n"); len(trimmed) > 0 && trimmed[0] == '[' {
		return raw
	}
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return Null
	}
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	objs := make([]map[string]json.RawMessage, 0, len(keys))
	for _, k := range keys {
		key, err := json.Marshal(k)
		if err != nil {
			return Null
		}
		objs = append(objs, map[string]json.RawMessage{
			keyField: key,
			valField: entries[k],
		})
	}
	b, err := json.Marshal(objs)
	if err != nil {
		return Null
	}
	return b
}

// PadObjectKeys adds each key, as null, to every object of an old block array
// that lacks it, for a nested object that gained fields: tftypes rejects a
// missing key. Existing keys are kept. Null/absent → null.
func PadObjectKeys(raw json.RawMessage, keys ...string) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return Null
	}
	var elems []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &elems); err != nil {
		return raw
	}
	for _, e := range elems {
		for _, k := range keys {
			if _, ok := e[k]; !ok {
				e[k] = Null
			}
		}
	}
	b, err := json.Marshal(elems)
	if err != nil {
		return raw
	}
	return b
}

// StringToNum converts an old string id ("42") to a JSON number (42), used where
// the new schema declares id as a number (e.g. aws_account). Null/absent → null.
func StringToNum(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return Null
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return raw // already a number
	}
	if s == "" {
		return Null
	}
	return json.RawMessage(s)
}
