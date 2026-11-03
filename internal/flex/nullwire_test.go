package flex_test

import (
	"encoding/json"
	"testing"

	"terraform-provider-kion/internal/flex"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNullIntDecodesBothRenderings covers the shape that made five project_note
// records unimportable: the private read returns last_update_user_id as
// {"Int":1,"Valid":true}, while the wire struct declared int64 and so failed to
// decode the whole response.
func TestNullIntDecodesBothRenderings(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want flex.NullInt
	}{
		{"bare number", `5`, flex.NullInt{Int: 5, Valid: true}},
		{"wrapper", `{"Int":1,"Valid":true}`, flex.NullInt{Int: 1, Valid: true}},
		{"wrapper not valid", `{"Int":0,"Valid":false}`, flex.NullInt{}},
		{"database/sql field name", `{"Int64":7,"Valid":true}`, flex.NullInt{Int: 7, Valid: true}},
		{"wrapper without Valid", `{"Int":3}`, flex.NullInt{Int: 3, Valid: true}},
		{"null", `null`, flex.NullInt{}},
		{"zero is still valid", `0`, flex.NullInt{Int: 0, Valid: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got flex.NullInt
			require.NoError(t, json.Unmarshal([]byte(tc.in), &got))
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNullIntDecodeRejectsGarbage(t *testing.T) {
	var got flex.NullInt
	assert.Error(t, json.Unmarshal([]byte(`"not a number"`), &got))
}

// TestNullIntEncodesBareScalar is the write-side guarantee: the raw PATCH body
// must look exactly as it did when the field was a plain int64, so making the
// read tolerant cannot change what is sent.
func TestNullIntEncodesBareScalar(t *testing.T) {
	body, err := json.Marshal(struct {
		X *flex.NullInt `json:"x,omitempty"`
	}{X: &flex.NullInt{Int: 5, Valid: true}})
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":5}`, string(body))
}

// TestNullIntOmittedWhenUnset covers the other half: a null model attribute must
// leave the key out of the body, which is what `omitempty` did for a plain int64.
func TestNullIntOmittedWhenUnset(t *testing.T) {
	body, err := json.Marshal(struct {
		X *flex.NullInt `json:"x,omitempty"`
	}{X: flex.NullIntFromFramework(types.Int64Null())})
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(body))
}

func TestNullIntFrameworkRoundTrip(t *testing.T) {
	assert.Equal(t, types.Int64Value(9), flex.NullIntPtrToFramework(flex.NullIntFromFramework(types.Int64Value(9))))
	assert.Equal(t, types.Int64Null(), flex.NullIntPtrToFramework(nil))
	assert.Equal(t, types.Int64Null(), flex.NullIntPtrToFramework(&flex.NullInt{Int: 4}), "not valid is null, not 4")
	assert.Nil(t, flex.NullIntFromFramework(types.Int64Unknown()), "unknown is omitted, not sent as 0")
}

// TestNullStringDecodesBothRenderings covers the dashboard failure: the private
// read returns description as an object where the wire struct declared string.
func TestNullStringDecodesBothRenderings(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want flex.NullString
	}{
		{"bare string", `"hi"`, flex.NullString{String: "hi", Valid: true}},
		{"wrapper", `{"String":"hi","Valid":true}`, flex.NullString{String: "hi", Valid: true}},
		{"wrapper not valid", `{"String":"","Valid":false}`, flex.NullString{}},
		{"wrapper without Valid", `{"String":"hi"}`, flex.NullString{String: "hi", Valid: true}},
		{"null", `null`, flex.NullString{}},
		{"empty string is still valid on the wire", `""`, flex.NullString{String: "", Valid: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got flex.NullString
			require.NoError(t, json.Unmarshal([]byte(tc.in), &got))
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNullStringDecodeRejectsGarbage(t *testing.T) {
	var got flex.NullString
	assert.Error(t, json.Unmarshal([]byte(`12`), &got))
}

func TestNullStringEncodesBareScalar(t *testing.T) {
	body, err := json.Marshal(struct {
		X *flex.NullString `json:"x,omitempty"`
	}{X: &flex.NullString{String: "hi", Valid: true}})
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":"hi"}`, string(body))
}

func TestNullStringOmittedWhenUnset(t *testing.T) {
	body, err := json.Marshal(struct {
		X *flex.NullString `json:"x,omitempty"`
	}{X: flex.NullStringFromFramework(types.StringNull())})
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(body))
}

// TestNullStringEmptyReadsAsNull mirrors OptStringToFrameworkNullIfEmpty: a
// generated configuration writes an unset optional as null, so a read that
// stored "" would plan an update whose entire content is `"" -> null`.
func TestNullStringEmptyReadsAsNull(t *testing.T) {
	assert.Equal(t, types.StringNull(), flex.NullStringPtrToFramework(&flex.NullString{String: "", Valid: true}))
	assert.Equal(t, types.StringNull(), flex.NullStringPtrToFramework(nil))
	assert.Equal(t, types.StringValue("hi"), flex.NullStringPtrToFramework(&flex.NullString{String: "hi", Valid: true}))
}

// TestNullTimeDecodesBothRenderings covers project_note.updated_at and
// dashboard.updated_at, which arrive as sql.NullTime rather than a string.
func TestNullTimeDecodesBothRenderings(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want flex.NullTime
	}{
		{"bare string", `"2026-07-09T14:48:06Z"`, flex.NullTime{Time: "2026-07-09T14:48:06Z", Valid: true}},
		{"wrapper", `{"Time":"2026-07-09T14:48:06Z","Valid":true}`, flex.NullTime{Time: "2026-07-09T14:48:06Z", Valid: true}},
		{"never updated", `{"Time":"0001-01-01T00:00:00Z","Valid":false}`, flex.NullTime{}},
		{"null", `null`, flex.NullTime{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got flex.NullTime
			require.NoError(t, json.Unmarshal([]byte(tc.in), &got))
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestNullTimeZeroInstantIsNull keeps the Go zero time out of state: a record
// that was never updated must read as null, not as year 1.
func TestNullTimeZeroInstantIsNull(t *testing.T) {
	assert.Equal(t, types.StringNull(), flex.NullTimePtrToFramework(&flex.NullTime{Time: "0001-01-01T00:00:00Z", Valid: true}))
	assert.Equal(t, types.StringNull(), flex.NullTimePtrToFramework(nil))
	assert.Equal(t, types.StringValue("2026-07-09T14:48:06Z"),
		flex.NullTimePtrToFramework(&flex.NullTime{Time: "2026-07-09T14:48:06Z", Valid: true}))
}

func TestNullTimeEncodesBareString(t *testing.T) {
	body, err := json.Marshal(struct {
		X *flex.NullTime `json:"x,omitempty"`
	}{X: &flex.NullTime{Time: "2026-07-09T14:48:06Z", Valid: true}})
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":"2026-07-09T14:48:06Z"}`, string(body))
}

// TestJSONStringDecodesObjectAndString covers dashboard.config, which the schema
// models as a string of JSON but /v1/dashboard returns as the object itself.
func TestJSONStringDecodesObjectAndString(t *testing.T) {
	var fromObject flex.JSONString
	require.NoError(t, json.Unmarshal([]byte(`{"b":2,"a":1}`), &fromObject))
	assert.Equal(t, `{"a":1,"b":2}`, fromObject.Value, "canonicalised: compact, keys sorted")
	assert.True(t, fromObject.Set)

	var fromString flex.JSONString
	require.NoError(t, json.Unmarshal([]byte(`"{\"b\": 2, \"a\": 1}"`), &fromString))
	assert.Equal(t, `{"a":1,"b":2}`, fromString.Value, "a string of JSON canonicalises the same way")

	// Both renderings land on the same bytes, which is what lets a generated
	// configuration plan clean against either.
	assert.Equal(t, fromObject.Value, fromString.Value)
}

// TestJSONStringEncodesAsString is the write-side guarantee: the public create
// and update bodies still receive config as a string, not as an object.
func TestJSONStringEncodesAsString(t *testing.T) {
	body, err := json.Marshal(struct {
		X *flex.JSONString `json:"x,omitempty"`
	}{X: flex.JSONStringFromFramework(types.StringValue(`{"a":1}`))})
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":"{\"a\":1}"}`, string(body))
}

func TestJSONStringOmittedWhenUnset(t *testing.T) {
	body, err := json.Marshal(struct {
		X *flex.JSONString `json:"x,omitempty"`
	}{X: flex.JSONStringFromFramework(types.StringNull())})
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(body))
	assert.Equal(t, types.StringNull(), flex.JSONStringPtrToFramework(nil))
}
