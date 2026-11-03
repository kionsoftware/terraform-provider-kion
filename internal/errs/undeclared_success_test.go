package errs_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ogen-go/ogen/validate"
	"github.com/stretchr/testify/assert"

	"terraform-provider-kion/internal/errs"
)

// undecodable builds the error an ogen-generated decoder returns for a status
// the spec does not declare, the same way the SDK does.
func undecodable(status int) error {
	return validate.UnexpectedStatusCodeWithResponse(&http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(`{"status":200,"message":""}`)),
	})
}

func TestIsUndeclaredSuccess(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		// The live case: PatchOUCloudAccessRole declares 201, the server answers
		// 200, and the write already landed.
		{"200 the spec does not declare", undecodable(200), true},
		{"204 the spec does not declare", undecodable(204), true},
		{"299 is still 2xx", undecodable(299), true},
		// Anything the server did not treat as a success stays an error, however
		// the spec describes it.
		{"300 is not a success", undecodable(300), false},
		{"404 is not a success", undecodable(404), false},
		{"500 is not a success", undecodable(500), false},
		// Only this one decode failure is tolerated; every other error still is one.
		{"nil", nil, false},
		{"unrelated error", errors.New("connection refused"), false},
		{"wrong content type", validate.InvalidContentType("text/html"), false},
		// ogen joins the decode error with the body-read error and the SDK wraps
		// the result, so the check has to unwrap rather than type-assert.
		{"wrapped", fmt.Errorf("decode response: %w", undecodable(200)), true},
		{"joined", errors.Join(undecodable(200), errors.New("read body")), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, errs.IsUndeclaredSuccess(tt.err))
		})
	}
}
