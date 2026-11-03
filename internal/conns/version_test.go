package conns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseKionVersion(t *testing.T) {
	tests := []struct {
		in                  string
		major, minor, patch int
		wantErr             bool
	}{
		{"3.16.1", 3, 16, 1, false},
		{"3.16", 3, 16, 0, false},
		{" 3.14.2 ", 3, 14, 2, false},
		{"3.16.1-rc2", 3, 16, 1, false},
		{"garbage", 0, 0, 0, true},
		{"3", 0, 0, 0, true},
	}
	for _, tt := range tests {
		got, err := ParseKionVersion(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseKionVersion(%q): expected error, got none", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseKionVersion(%q): unexpected error %v", tt.in, err)
			continue
		}
		if got.Major != tt.major || got.Minor != tt.minor || got.Patch != tt.patch {
			t.Errorf("ParseKionVersion(%q) = %d.%d.%d, want %d.%d.%d",
				tt.in, got.Major, got.Minor, got.Patch, tt.major, tt.minor, tt.patch)
		}
	}
}

func TestKionVersionAtLeast(t *testing.T) {
	tests := []struct {
		have, min string
		want      bool
	}{
		{"3.16.1", "3.16.0", true},
		{"3.16.0", "3.16.0", true},
		{"3.16.0", "3.16.1", false},
		{"3.15.9", "3.16.0", false},
		{"4.0.0", "3.16.0", true},
		{"3.14.2", "3.16.0", false},
		{"3.17.0", "3.16.0", true},
	}
	for _, tt := range tests {
		have := MustParseKionVersion(tt.have)
		minVer := MustParseKionVersion(tt.min)
		if got := have.AtLeast(minVer); got != tt.want {
			t.Errorf("%s.AtLeast(%s) = %v, want %v", tt.have, tt.min, got, tt.want)
		}
	}
}

func TestDetectVersion(t *testing.T) {
	// Emulate GET /api/version → {"status":200,"data":"3.16.1"} exactly as the
	// real endpoint responds.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, werr := w.Write([]byte(`{"status":200,"data":"3.16.1"}`)); werr != nil {
			t.Errorf("writing version response: %v", werr)
		}
	}))
	defer srv.Close()

	// APIURL carries the "/api" suffix, mirroring how Configure builds it.
	c := &KionClient{APIURL: srv.URL + "/api", HTTPClient: srv.Client()}
	if err := c.DetectVersion(context.Background()); err != nil {
		t.Fatalf("DetectVersion: %v", err)
	}
	if !c.VersionDetected {
		t.Fatal("VersionDetected = false, want true")
	}
	if c.Version.Major != 3 || c.Version.Minor != 16 || c.Version.Patch != 1 {
		t.Fatalf("Version = %s, want 3.16.1", c.Version)
	}
}

func TestDetectVersion_failureIsNonFatal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &KionClient{APIURL: srv.URL + "/api", HTTPClient: srv.Client()}
	if err := c.DetectVersion(context.Background()); err == nil {
		t.Fatal("expected an error on 500 response")
	}
	if c.VersionDetected {
		t.Fatal("VersionDetected should stay false after a failed detect")
	}
}
