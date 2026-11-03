package conns_test

import (
	"testing"

	"terraform-provider-kion/internal/conns"
)

func TestPathUnless(t *testing.T) {
	if got := conns.PathUnless(false, "/create", "/import"); got != "/create" {
		t.Errorf("got %q, want /create", got)
	}
	if got := conns.PathUnless(true, "/create", "/import"); got != "/import" {
		t.Errorf("got %q, want /import", got)
	}
}
