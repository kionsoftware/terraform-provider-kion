package errs_test

import (
	"testing"

	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"

	"terraform-provider-kion/internal/errs"
)

func TestIsEmptyRecord(t *testing.T) {
	empty := &generated.BillingSourceGovcloudResponse{Data: generated.NewOptBillingSourceGovcloud(generated.BillingSourceGovcloud{})}
	if !errs.IsEmptyRecord(empty) {
		t.Error("an all-zero record is how the route answers a deleted one")
	}
	full := &generated.BillingSourceGovcloudResponse{Data: generated.NewOptBillingSourceGovcloud(generated.BillingSourceGovcloud{ID: generated.NewOptUint64(5)})}
	if errs.IsEmptyRecord(full) {
		t.Error("a record with an id is not empty")
	}
	if errs.IsEmptyRecord(&generated.NotFoundResponse{}) || errs.IsEmptyRecord(nil) {
		t.Error("only a success payload with data can be an empty record")
	}
}
