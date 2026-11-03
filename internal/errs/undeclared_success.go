package errs

import (
	"errors"

	"github.com/ogen-go/ogen/validate"
)

// IsUndeclaredSuccess reports whether err is nothing but the SDK refusing to
// decode a response the server considered a success.
//
// The SDK is ogen-generated from the OpenAPI spec, so its decoder has an arm per
// status code the spec declares and rejects anything else. Where the spec
// under-declares a success, a write that succeeded comes back as
// `decode response: unexpected status code: 200` and the provider reports an
// error for a change the server already made -- state and reality diverge,
// which is worse than a clean failure.
//
// Four PATCH operations are in that position today. Each annotates its response
// in the API spec as `201: OKResponse`, which is self-contradictory (OKResponse is
// the 200 model) and the server duly answers 200:
//
//	PatchAppConfig, PatchAppRole, PatchOUCloudAccessRole, PatchProjectCloudAccessRole
//
// The real fix is upstream: correct those four annotations to 200, regenerate
// the spec and the SDK, and bump the replace in go.mod. Until that lands, the
// generated Update bodies treat an undeclared 2xx as the success it is and take
// their state from the read-back that follows, which is authoritative either
// way. Nothing else is tolerated: a status the spec does not declare and the
// server does not consider a success still fails.
//
// It is written to retire itself. Once the spec declares 200, the decoder
// returns a response instead of an error and this never fires again.
func IsUndeclaredSuccess(err error) bool {
	var unexpected *validate.UnexpectedStatusCodeError
	if !errors.As(err, &unexpected) {
		return false
	}
	return unexpected.StatusCode >= 200 && unexpected.StatusCode < 300
}
