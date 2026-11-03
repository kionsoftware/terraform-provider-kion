package crud

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// An unknown plan value (a Computed list left out of configuration) must not
// reach the API as []: for an OptNil array, [] is a value ("no scopes"),
// while unset keeps the server's. A null plan value is still sent as [], so
// an Optional list the practitioner removed is cleared.
func TestEntityOptNilArrayUnknownIsOmitted(t *testing.T) {
	spec := fixtureSpec{
		Name: "widget",
		Dir:  filepath.Join("testdata", "optnilarray"),
		Ops: resOps{
			Create: &opRef{Method: "POST", Path: "/v3/widget"},
			Read:   &opRef{Method: "GET", Path: "/v3/widget/{id}"},
			Update: &opRef{Method: "PUT", Path: "/v3/widget/{id}"},
			Delete: &opRef{Method: "DELETE", Path: "/v3/widget/{id}"},
		},
	}
	src := generateParsed(t, spec)["widget.go"]
	want := "Scopes: generated.OptNilStringArray{Value: scopes, Set: !plan.Scopes.IsUnknown()}"
	for _, fn := range []string{"Create", "Update"} {
		body := funcBody(t, src, fn)
		require.Contains(t, body, want, "%s body:\n%s", fn, body)
		require.NotContains(t, body, "Value: scopes, Set: true}", "%s still sends an unknown list as []", fn)
	}
}
