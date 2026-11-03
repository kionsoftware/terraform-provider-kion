// Package errs provides helper functions for extracting errors from SDK responses.
package errs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
)

// CreatedID extracts the record ID from a CreatedResponse returned by POST
// endpoints. Returns an error diagnostic if the response is not a CreatedResponse
// or if the record ID is not set.
func CreatedID(res any) (int64, diag.Diagnostics) {
	var diags diag.Diagnostics

	cr, ok := res.(*generated.CreatedResponse)
	if !ok {
		diags.Append(ResponseDiagnostics("extracting created ID", res)...)
		if !diags.HasError() {
			diags.AddError("Unexpected Response", fmt.Sprintf("expected *CreatedResponse, got %T", res))
		}
		return 0, diags
	}

	if !cr.RecordID.IsSet() {
		diags.AddError("Missing Record ID", "the created response did not contain a record ID")
		return 0, diags
	}

	return RawCreatedID(int64(cr.RecordID.Value))
}

// RawCreatedID validates a record ID decoded from a create response by hand,
// for the endpoints served over raw HTTP rather than through the SDK.
//
// Zero is rejected: Kion issues ids from 1, so a zero is either a response with
// no record id in it at all or one that explicitly reported a non-id. Writing it
// to state produces a resource that exists in Kion and can never be addressed —
// refresh reads id 0 and finds nothing, destroy deletes id 0 — so every apply
// leaks a record (#71). Failing is the only outcome that says so.
func RawCreatedID(recordID int64) (int64, diag.Diagnostics) {
	var diags diag.Diagnostics

	if recordID <= 0 {
		diags.AddError(
			"Missing Record ID",
			fmt.Sprintf("the create response reported record ID %d, which is not a usable Kion id; "+
				"the record may have been created but Terraform cannot manage it", recordID),
		)
		return 0, diags
	}

	return recordID, diags
}

// RawCreatedStringID rejects an empty id, which Kion never issues and
// Terraform could never address.
func RawCreatedStringID(id string) (string, diag.Diagnostics) {
	if id == "" {
		return "", diag.Diagnostics{diag.NewErrorDiagnostic("creating resource", "the create response did not include an id")}
	}
	return id, nil
}

// IsInternalServerError reports whether the response is a 500. Some reads
// answer a missing record with one; see missing_via_list.
func IsInternalServerError(res any) bool {
	_, ok := res.(*generated.InternalServerErrorResponse)
	return ok
}

// IsNotFound returns true if the response is a *NotFoundResponse.
func IsNotFound(res any) bool {
	_, ok := res.(*generated.NotFoundResponse)
	return ok
}

// ResponseDiagnostics converts error response types (BadRequest, Unauthorized,
// Forbidden, NotFound, InternalServerError) into Terraform diagnostics.
// Returns empty diagnostics for success response types.
//
// Callers pass ANY response here, so silence on success is the contract: a
// delete that answers 200 OK must not be reported as a failure. A caller that
// has ALREADY established the response is unusable wants UnexpectedResponse
// instead, which is never silent.
func ResponseDiagnostics(summary string, res any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch r := res.(type) {
	case *generated.BadRequestResponse:
		diags.AddError(summary, fmt.Sprintf("Bad Request: %s", r.Message.Or("no message")))
	case *generated.UnauthorizedResponse:
		diags.AddError(summary, fmt.Sprintf("Unauthorized: %s", r.Message.Or("no message")))
	case *generated.ForbiddenResponse:
		diags.AddError(summary, fmt.Sprintf("Forbidden: %s", r.Message.Or("no message")))
	case *generated.NotFoundResponse:
		diags.AddError(summary, fmt.Sprintf("Not Found: %s", r.Message.Or("no message")))
	case *generated.InternalServerErrorResponse:
		diags.AddError(summary, fmt.Sprintf("Internal Server Error: %s", r.Message.Or("no message")))
	case *generated.UnprocessableEntityResponse:
		diags.AddError(summary, fmt.Sprintf("Unprocessable Entity: %s", r.Message.Or("no message")))
	}

	return diags
}

// UnexpectedResponse reports a response the caller has already determined it
// cannot use. It ALWAYS produces an error.
//
// The distinction from ResponseDiagnostics matters because of what the callers
// do next. A create that cannot find its record in the response does:
//
//	created, ok := out.(*generated.XResponse)
//	if !ok || !created.Data.Set {
//		resp.Diagnostics.Append(…)
//		return
//	}
//
// With ResponseDiagnostics, a response that is neither a modeled error nor the
// expected success type produced NO diagnostic, so that returned with neither an
// error nor any state:
//
//	Missing Resource State After Create
//	The Terraform Provider unexpectedly returned no resource state after having
//	no errors in the resource creation.
//
// The record is created and then orphaned -- Terraform never learns its id, so
// nothing will ever address or destroy it. kion_scope is how this surfaced: the
// spec models its create response as {data: …} while Kion answers {status,
// record_id}, so the typed response carried no data at all.
//
// It still renders a modeled error's own message when there is one; the point
// is only that it cannot return empty.
func UnexpectedResponse(summary string, res any) diag.Diagnostics {
	if diags := ResponseDiagnostics(summary, res); diags.HasError() {
		return diags
	}
	var diags diag.Diagnostics
	// Naming the Go type is what makes a spec/server mismatch diagnosable.
	diags.AddError(summary, fmt.Sprintf(
		"unexpected API response %T: it did not carry the record this operation needed. "+
			"This usually means the OpenAPI spec models the response differently from what the server sends.", res))
	return diags
}
