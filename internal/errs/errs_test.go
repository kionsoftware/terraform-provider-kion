package errs_test

import (
	"testing"

	"terraform-provider-kion/internal/errs"

	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatedID_Success(t *testing.T) {
	t.Parallel()

	res := &generated.CreatedResponse{RecordID: generated.NewOptUint64(42)}
	id, diags := errs.CreatedID(res)

	require.False(t, diags.HasError())
	assert.Equal(t, int64(42), id)
}

func TestCreatedID_MissingRecordID(t *testing.T) {
	t.Parallel()

	res := &generated.CreatedResponse{} // RecordID unset
	_, diags := errs.CreatedID(res)

	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Record ID")
}

// An explicit record_id of 0 passes IsSet but is never a real Kion id; taking
// it would leave a resource in state that can never be refreshed or deleted.
func TestCreatedID_ZeroRecordID(t *testing.T) {
	t.Parallel()

	res := &generated.CreatedResponse{RecordID: generated.NewOptUint64(0)}
	_, diags := errs.CreatedID(res)

	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Record ID")
}

func TestRawCreatedID(t *testing.T) {
	t.Parallel()

	id, diags := errs.RawCreatedID(42)
	require.False(t, diags.HasError())
	assert.Equal(t, int64(42), id)
}

// The funding_source_note leak (#71): the create response carries no record_id
// at all, so the hand-rolled envelope decodes 0 with no error.
func TestRawCreatedID_Zero(t *testing.T) {
	t.Parallel()

	id, diags := errs.RawCreatedID(0)
	require.True(t, diags.HasError())
	assert.Zero(t, id)
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Record ID")
	assert.Contains(t, diags.Errors()[0].Detail(), "Terraform cannot manage it")
}

func TestRawCreatedID_Negative(t *testing.T) {
	t.Parallel()

	_, diags := errs.RawCreatedID(-1)
	require.True(t, diags.HasError())
}

func TestRawCreatedStringID(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, in string
		wantErr  bool
	}{
		{name: "empty", in: "", wantErr: true},
		{name: "ok", in: "abc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			id, diags := errs.RawCreatedStringID(tc.in)
			require.Equal(t, tc.wantErr, diags.HasError())
			if tc.wantErr {
				assert.Empty(t, id)
				assert.Contains(t, diags.Errors()[0].Detail(), "did not include an id")
				return
			}
			assert.Equal(t, tc.in, id)
		})
	}
}

func TestCreatedID_ErrorResponseType(t *testing.T) {
	t.Parallel()

	res := &generated.BadRequestResponse{Message: generated.NewOptString("nope")}
	_, diags := errs.CreatedID(res)

	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "Bad Request: nope")
}

func TestCreatedID_UnexpectedType(t *testing.T) {
	t.Parallel()

	_, diags := errs.CreatedID("not a response")

	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Unexpected Response")
}

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	assert.True(t, errs.IsNotFound(&generated.NotFoundResponse{}))
	assert.False(t, errs.IsNotFound(&generated.CreatedResponse{}))
	assert.False(t, errs.IsNotFound("nope"))
}

func TestResponseDiagnostics(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		res        any
		wantErr    bool
		wantDetail string
	}{
		{"bad request", &generated.BadRequestResponse{Message: generated.NewOptString("b")}, true, "Bad Request: b"},
		{"unauthorized", &generated.UnauthorizedResponse{Message: generated.NewOptString("u")}, true, "Unauthorized: u"},
		{"forbidden", &generated.ForbiddenResponse{Message: generated.NewOptString("f")}, true, "Forbidden: f"},
		{"not found", &generated.NotFoundResponse{Message: generated.NewOptString("n")}, true, "Not Found: n"},
		{"internal", &generated.InternalServerErrorResponse{Message: generated.NewOptString("i")}, true, "Internal Server Error: i"},
		{"unprocessable", &generated.UnprocessableEntityResponse{Message: generated.NewOptString("e")}, true, "Unprocessable Entity: e"},
		{"no message uses fallback", &generated.NotFoundResponse{}, true, "Not Found: no message"},
		// Silence on success is the contract here: a delete that answers 200 OK
		// must not be reported as a failure. A caller that has already decided
		// the response is unusable uses UnexpectedResponse; see its test below.
		{"non-error type is empty", &generated.CreatedResponse{}, false, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			diags := errs.ResponseDiagnostics("op failed", tc.res)
			require.Equal(t, tc.wantErr, diags.HasError())
			if tc.wantErr {
				assert.Equal(t, "op failed", diags.Errors()[0].Summary())
				assert.Contains(t, diags.Errors()[0].Detail(), tc.wantDetail)
			}
		})
	}
}

// UnexpectedResponse can never return empty, which is the whole point of it
// existing separately: its callers have already decided the response is
// unusable and `return` straight after, so an empty result there leaves the
// operation looking successful with no state behind it.
func TestUnexpectedResponse(t *testing.T) {
	t.Parallel()

	t.Run("renders a modeled error's own message", func(t *testing.T) {
		t.Parallel()
		diags := errs.UnexpectedResponse("op failed", &generated.BadRequestResponse{
			Message: generated.NewOptString("b"),
		})
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "Bad Request: b")
	})

	for _, tc := range []struct {
		name string
		res  any
	}{
		{"success type", &generated.CreatedResponse{}},
		{"nil", nil},
	} {
		t.Run(tc.name+" still errors", func(t *testing.T) {
			t.Parallel()
			diags := errs.UnexpectedResponse("op failed", tc.res)
			require.True(t, diags.HasError(), "UnexpectedResponse must never return empty")
			assert.Equal(t, "op failed", diags.Errors()[0].Summary())
			assert.Contains(t, diags.Errors()[0].Detail(), "unexpected API response")
		})
	}
}
