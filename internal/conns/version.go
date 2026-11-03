package conns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// KionVersion is a parsed Kion release version (e.g. 3.16.1). It is populated
// once at provider Configure time via DetectVersion and used to gate resources
// that require a minimum Kion version.
type KionVersion struct {
	Major, Minor, Patch int
	Raw                 string
}

// ParseKionVersion parses a version string like "3.16.1" (patch optional, and
// trailing pre-release/build suffixes such as "1-rc2" are tolerated).
func ParseKionVersion(s string) (KionVersion, error) {
	v := KionVersion{Raw: strings.TrimSpace(s)}
	parts := strings.SplitN(v.Raw, ".", 3)
	if len(parts) < 2 {
		return v, fmt.Errorf("invalid Kion version %q: want at least MAJOR.MINOR", s)
	}

	var err error
	if v.Major, err = strconv.Atoi(parts[0]); err != nil {
		return v, fmt.Errorf("invalid major version in %q: %w", s, err)
	}
	if v.Minor, err = strconv.Atoi(parts[1]); err != nil {
		return v, fmt.Errorf("invalid minor version in %q: %w", s, err)
	}
	if len(parts) == 3 {
		patch := parts[2]
		if i := strings.IndexAny(patch, "-+"); i >= 0 {
			patch = patch[:i]
		}
		// Best-effort: a non-numeric patch leaves Patch at 0.
		if p, perr := strconv.Atoi(patch); perr == nil {
			v.Patch = p
		}
	}
	return v, nil
}

// MustParseKionVersion parses a version string and panics on failure. Use it
// for compile-time-constant minimums in resource code, e.g.
// conns.MustParseKionVersion("3.16.0").
func MustParseKionVersion(s string) KionVersion {
	v, err := ParseKionVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}

// AtLeast reports whether v is greater than or equal to minVer, comparing major,
// then minor, then patch.
func (v KionVersion) AtLeast(minVer KionVersion) bool {
	switch {
	case v.Major != minVer.Major:
		return v.Major > minVer.Major
	case v.Minor != minVer.Minor:
		return v.Minor > minVer.Minor
	default:
		return v.Patch >= minVer.Patch
	}
}

// String renders the version, preferring the raw form the server reported.
func (v KionVersion) String() string {
	if v.Raw != "" {
		return v.Raw
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// versionResponse matches GET /api/version: {"status":200,"data":"3.16.1"}.
type versionResponse struct {
	Status int    `json:"status"`
	Data   string `json:"data"`
}

// DetectVersion queries GET {APIURL}/version and records the result on the
// client. APIURL already carries the "/api" suffix, so the effective path is
// e.g. https://host/api/version.
//
// It is best-effort: on any failure the client's version is left undetected
// and the error is returned for the caller to log. Detection failure is not
// fatal, RequireMinKionVersion degrades to a warning so the API can enforce
// support on its own.
func (c *KionClient) DetectVersion(ctx context.Context) (err error) {
	url := strings.TrimRight(c.APIURL, "/") + "/version"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("building version request: %w", err)
	}
	// The version endpoint is unauthenticated, but send credentials if we have
	// them in case an install fronts it with auth.
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	} else if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("querying %s: %w", url, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("version endpoint %s returned status %d", url, resp.StatusCode)
	}

	var vr versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return fmt.Errorf("decoding version response: %w", err)
	}

	v, err := ParseKionVersion(vr.Data)
	if err != nil {
		return err
	}

	c.Version = v
	c.VersionDetected = true
	return nil
}
