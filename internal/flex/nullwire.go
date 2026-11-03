package flex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Kion's private endpoints render some columns through Go's sql.NullInt64 /
// sql.NullString, which marshal as an object rather than the bare scalar the
// public spec advertises: last_update_user_id comes back as {"Int":1,"Valid":true}
// and dashboard's description as the string wrapper. A wire struct that declares
// them int64/string fails to decode the whole response, which is why five
// project_note and four dashboard records could not be imported at all.
//
// NullInt and NullString accept either rendering on the way in and emit the bare
// scalar on the way out, so a single wire struct still serves both the read and
// the raw write. Used as a pointer field with `omitempty`, an unset value is
// omitted from a write body exactly as the plain scalar was.

// NullInt decodes either a bare JSON number or Kion's sql.NullInt64 rendering,
// {"Int":1,"Valid":true}. It encodes as the bare number, or null when not valid.
type NullInt struct {
	Int   int64
	Valid bool
}

// nullIntWire is the object rendering. Int64 is accepted alongside Int because
// database/sql's own field name is Int64; Kion's wrapper uses Int.
type nullIntWire struct {
	Int   *int64 `json:"Int"`
	Int64 *int64 `json:"Int64"`
	Valid *bool  `json:"Valid"`
}

// UnmarshalJSON accepts either rendering: a bare JSON number, or the
// sql.NullInt64 object. An absent, null or not-valid value leaves NullInt zero.
func (n *NullInt) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*n = NullInt{}
		return nil
	}
	if b[0] != '{' {
		var v int64
		if err := json.Unmarshal(b, &v); err != nil {
			return fmt.Errorf("decoding int: %w", err)
		}
		*n = NullInt{Int: v, Valid: true}
		return nil
	}
	var w nullIntWire
	if err := json.Unmarshal(b, &w); err != nil {
		return fmt.Errorf("decoding sql.NullInt64 wrapper: %w", err)
	}
	v := w.Int
	if v == nil {
		v = w.Int64
	}
	// A wrapper with no Valid key is trusted when it carries a value: the
	// alternative is discarding a value the API did send.
	valid := v != nil
	if w.Valid != nil {
		valid = *w.Valid && v != nil
	}
	if !valid {
		*n = NullInt{}
		return nil
	}
	*n = NullInt{Int: *v, Valid: true}
	return nil
}

// MarshalJSON writes the bare number, so a write body looks exactly as it did
// when the field was a plain int64. A not-valid value writes null.
func (n NullInt) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Int)
}

// NullIntFromFramework builds the wire value for a write body. A null or unknown
// model attribute yields nil, which `omitempty` drops from the body.
func NullIntFromFramework(v types.Int64) *NullInt {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return &NullInt{Int: v.ValueInt64(), Valid: true}
}

// NullIntPtrToFramework is its inverse: an absent or not-valid value is null.
func NullIntPtrToFramework(v *NullInt) types.Int64 {
	if v == nil || !v.Valid {
		return types.Int64Null()
	}
	return types.Int64Value(v.Int)
}

// NullString decodes either a bare JSON string or Kion's sql.NullString
// rendering, {"String":"x","Valid":true}. It encodes as the bare string, or null
// when not valid.
type NullString struct {
	String string
	Valid  bool
}

type nullStringWire struct {
	String *string `json:"String"`
	Valid  *bool   `json:"Valid"`
}

// UnmarshalJSON accepts either rendering: a bare JSON string, or the
// sql.NullString object. An absent, null or not-valid value leaves NullString zero.
func (n *NullString) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*n = NullString{}
		return nil
	}
	if b[0] != '{' {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return fmt.Errorf("decoding string: %w", err)
		}
		*n = NullString{String: v, Valid: true}
		return nil
	}
	var w nullStringWire
	if err := json.Unmarshal(b, &w); err != nil {
		return fmt.Errorf("decoding sql.NullString wrapper: %w", err)
	}
	valid := w.String != nil
	if w.Valid != nil {
		valid = *w.Valid && w.String != nil
	}
	if !valid {
		*n = NullString{}
		return nil
	}
	*n = NullString{String: *w.String, Valid: true}
	return nil
}

// MarshalJSON writes the bare string, so a write body looks exactly as it did
// when the field was a plain string. A not-valid value writes null.
func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.String)
}

// NullStringFromFramework builds the wire value for a write body. A null or
// unknown model attribute yields nil, which `omitempty` drops from the body.
func NullStringFromFramework(v types.String) *NullString {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return &NullString{String: v.ValueString(), Valid: true}
}

// NullStringPtrToFramework is its inverse. An absent, not-valid or empty value
// is null: `terraform plan -generate-config-out` writes an unset optional
// attribute as null, so storing "" would plan an update on `"" -> null` alone
// (the same churn OptStringToFrameworkNullIfEmpty exists to prevent).
func NullStringPtrToFramework(v *NullString) types.String {
	if v == nil || !v.Valid || v.String == "" {
		return types.StringNull()
	}
	return types.StringValue(v.String)
}

// NullTime decodes either a bare RFC 3339 timestamp string or Kion's sql.NullTime
// rendering, {"Time":"2026-07-09T14:48:06Z","Valid":true}. A record that has
// never been updated carries {"Time":"0001-01-01T00:00:00Z","Valid":false}, which
// reads as null rather than as the zero instant.
type NullTime struct {
	Time  string
	Valid bool
}

type nullTimeWire struct {
	Time  *string `json:"Time"`
	Valid *bool   `json:"Valid"`
}

// UnmarshalJSON accepts either rendering: a bare timestamp string, or the
// sql.NullTime object. An absent, null or not-valid value leaves NullTime zero.
func (n *NullTime) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*n = NullTime{}
		return nil
	}
	if b[0] != '{' {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return fmt.Errorf("decoding timestamp: %w", err)
		}
		*n = NullTime{Time: v, Valid: true}
		return nil
	}
	var w nullTimeWire
	if err := json.Unmarshal(b, &w); err != nil {
		return fmt.Errorf("decoding sql.NullTime wrapper: %w", err)
	}
	valid := w.Time != nil
	if w.Valid != nil {
		valid = *w.Valid && w.Time != nil
	}
	if !valid {
		*n = NullTime{}
		return nil
	}
	*n = NullTime{Time: *w.Time, Valid: true}
	return nil
}

// MarshalJSON writes the bare timestamp string, so a write body looks exactly as
// it did when the field was a plain string. A not-valid value writes null.
func (n NullTime) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Time)
}

// NullTimeFromFramework builds the wire value for a write body. A null or
// unknown model attribute yields nil, which `omitempty` drops from the body.
func NullTimeFromFramework(v types.String) *NullTime {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return &NullTime{Time: v.ValueString(), Valid: true}
}

// NullTimePtrToFramework is its inverse. An absent, not-valid or empty value is
// null, as is the zero instant Go marshals for an unset sql.NullTime: storing
// "0001-01-01T00:00:00Z" would show a timestamp for a record never updated.
func NullTimePtrToFramework(v *NullTime) types.String {
	if v == nil || !v.Valid || v.Time == "" || strings.HasPrefix(v.Time, "0001-01-01T00:00:00") {
		return types.StringNull()
	}
	return types.StringValue(v.Time)
}

// JSONString is a wire field the schema models as a string of JSON but the
// private read returns as real JSON: /v1/dashboard returns config as the object
// itself, so a wire struct declaring string failed to decode the response and no
// dashboard could be imported.
//
// It accepts either, and always encodes back as a JSON string, so a write body
// is unchanged. The decoded object is re-encoded canonically (compact, keys
// sorted) to match what Terraform's jsonencode produces, the byte equality a
// generated configuration needs to plan clean.
type JSONString struct {
	Value string
	Set   bool
}

// UnmarshalJSON accepts a JSON string containing JSON, or a JSON object/array,
// which is canonicalised into the string the schema models.
func (j *JSONString) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*j = JSONString{}
		return nil
	}
	if b[0] == '"' {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return fmt.Errorf("decoding JSON string: %w", err)
		}
		*j = JSONString{Value: CanonicalJSONString(v), Set: true}
		return nil
	}
	// Already JSON. Round-trip through a decode so the key order and spacing are
	// canonical rather than however the server happened to emit them.
	*j = JSONString{Value: CanonicalJSONString(string(b)), Set: true}
	return nil
}

// MarshalJSON writes the value as a JSON string, which is what the public create
// and update bodies expect; making the read tolerant does not change writes.
func (j JSONString) MarshalJSON() ([]byte, error) {
	if !j.Set {
		return []byte("null"), nil
	}
	return json.Marshal(j.Value)
}

// JSONStringFromFramework builds the wire value for a write body. A null or
// unknown model attribute yields nil, which `omitempty` drops from the body.
func JSONStringFromFramework(v types.String) *JSONString {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return &JSONString{Value: v.ValueString(), Set: true}
}

// JSONStringPtrToFramework is its inverse; an absent or empty value is null.
func JSONStringPtrToFramework(v *JSONString) types.String {
	if v == nil || !v.Set || v.Value == "" {
		return types.StringNull()
	}
	return types.StringValue(v.Value)
}
