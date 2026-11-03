package conns_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"terraform-provider-kion/internal/conns"

	"github.com/stretchr/testify/require"
)

func TestDetectVersion_ParsesVersionOn200(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/version", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":200,"data":"3.16.1"}`)); err != nil {
			t.Errorf("write body: %v", err)
		}
	}))
	defer srv.Close()

	// APIURL carries the /api suffix in production; DetectVersion appends /version.
	c := &conns.KionClient{APIURL: srv.URL + "/api", HTTPClient: srv.Client()}
	err := c.DetectVersion(t.Context())

	require.NoError(t, err)
	require.True(t, c.VersionDetected)
	require.Equal(t, "3.16.1", c.Version.String())
	require.True(t, c.Version.AtLeast(conns.MustParseKionVersion("3.16.0")))
}

func TestDetectVersion_ErrorsOnNon200(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := &conns.KionClient{APIURL: srv.URL + "/api", HTTPClient: srv.Client()}
	err := c.DetectVersion(t.Context())

	require.Error(t, err)
	require.ErrorContains(t, err, "status 503")
	require.False(t, c.VersionDetected)
}
