package conns_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/conns"
)

func TestDetectFinancialMode_ReadsBudgetMode(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{"budget mode on", `{"status":200,"data":{"budget_mode":true,"allocation_mode":false}}`, true},
		{"budget mode off", `{"status":200,"data":{"budget_mode":false,"allocation_mode":false}}`, false},
		// An install that omits the field is a spend-plan install as far as the
		// create endpoint is concerned, and the mode still counts as detected.
		{"field absent", `{"status":200,"data":{}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/api/v3/app-config", r.URL.Path)
				if _, err := w.Write([]byte(tc.body)); err != nil {
					t.Errorf("write body: %v", err)
				}
			}))
			defer srv.Close()

			c := &conns.KionClient{APIURL: srv.URL + "/api", HTTPClient: srv.Client()}
			require.NoError(t, c.DetectFinancialMode(t.Context()))
			require.True(t, c.FinancialModeDetected)
			require.Equal(t, tc.want, c.BudgetMode)
		})
	}
}

// Reading app-config needs a global settings permission, so a 403 is an ordinary
// outcome for a restricted credential. It must leave the mode undetected rather
// than claim spend-plan mode, which would send every create to the wrong endpoint.
func TestDetectFinancialMode_UndetectedOnForbidden(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		if _, err := w.Write([]byte(`{"status":403,"message":"Insufficient permission to view settings."}`)); err != nil {
			t.Errorf("write body: %v", err)
		}
	}))
	defer srv.Close()

	c := &conns.KionClient{APIURL: srv.URL + "/api", HTTPClient: srv.Client()}
	err := c.DetectFinancialMode(t.Context())

	require.Error(t, err)
	require.False(t, c.FinancialModeDetected)
	require.False(t, c.BudgetMode)
}

func TestDetectFinancialMode_UndetectedOnGarbage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`<html>not json</html>`)); err != nil {
			t.Errorf("write body: %v", err)
		}
	}))
	defer srv.Close()

	c := &conns.KionClient{APIURL: srv.URL + "/api", HTTPClient: srv.Client()}
	err := c.DetectFinancialMode(t.Context())

	require.ErrorContains(t, err, "decoding app-config response")
	require.False(t, c.FinancialModeDetected)
}
