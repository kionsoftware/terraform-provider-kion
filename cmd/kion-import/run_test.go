package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultProviderConstraint(t *testing.T) {
	t.Parallel()
	assert.Equal(t, ">= 1.5.0", defaultProviderConstraint("1.5.0"))
	assert.Equal(t, ">= 1.5.0", defaultProviderConstraint("v1.5.0"))
	// A source build carries no release version; fall back to the first release.
	assert.Equal(t, ">= 1.0.0", defaultProviderConstraint("dev"))
	assert.Equal(t, ">= 1.0.0", defaultProviderConstraint(""))
}

// A rejected API key used to exit 0, write an imports.tf of nothing but
// "Not imported" comments, and print the next terraform steps.
func TestRunFailsWhenEveryReadFails(t *testing.T) { //nolint:paralleltest // run reads package-level flags
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"status":401,"message":"Unauthorized"}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	out := filepath.Join(t.TempDir(), "imports.tf")
	setFlags(t, srv.URL, out)

	for _, probe := range []bool{false, true} {
		flagProbe = probe
		err := run(nil, nil)
		require.Error(t, err, "probe=%v", probe)
		assert.Contains(t, err.Error(), "every read failed")
		_, statErr := os.Stat(out)
		assert.True(t, os.IsNotExist(statErr), "probe=%v: %s was written", probe, out)
	}
}

func TestRunWritesImportsWhenAReadSucceeds(t *testing.T) { //nolint:paralleltest // run reads package-level flags
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{"status":200,"data":[]}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	out := filepath.Join(t.TempDir(), "imports.tf")
	setFlags(t, srv.URL, out)

	require.NoError(t, run(nil, nil))
	body, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(body), `version = ">= 1.0.0"`)
}

func TestHelpDoesNotPrintTheAPIKey(t *testing.T) { //nolint:paralleltest // sets an env var
	t.Setenv("KION_APIKEY", "app_1_secretvalue")
	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"-h"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, buf.String(), "--api-key")
	assert.NotContains(t, buf.String(), "secretvalue")
}

func setFlags(t *testing.T, url, out string) {
	t.Helper()
	url0, key0, out0, pv0, prefix0, inc0, probe0 := flagURL, flagAPIKey, flagOut, flagProviderVersion, flagAPIPrefix, flagInclude, flagProbe
	t.Cleanup(func() {
		flagURL, flagAPIKey, flagOut, flagProviderVersion = url0, key0, out0, pv0
		flagAPIPrefix, flagInclude, flagProbe = prefix0, inc0, probe0
		retryBackoff = nil
	})
	retryBackoff = []time.Duration{0, 0}
	flagURL, flagAPIKey, flagOut = url, "test-key", out
	flagProviderVersion = defaultProviderConstraint(version)
	flagAPIPrefix = ""
	flagInclude = []string{"kion_ou", "kion_label"}
	flagProbe = false
}
