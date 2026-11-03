package conns

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
)

// HTTPDoer is the subset of *http.Client used for raw API calls not covered by
// the generated SDK. It exists so request construction can be unit-tested with a
// mock and so callers can inject an httptest-backed client. A *http.Client
// satisfies it.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// KionClient wraps the generated SDK client and is stored as the provider's
// Meta value. Resources and data sources access it via Meta().
type KionClient struct {
	Client *generated.Client

	// APIURL is the API root the raw helpers and DetectVersion build paths
	// under, the base URL with the provider's resolved `apipath` already
	// applied (https://host/api by default, or whatever `apipath` set, which
	// may be nothing). Raw paths are appended to it verbatim, so it must NOT
	// have its own suffix re-added here; doing so ignored `apipath` and
	// produced /api/api/… against a default configuration.
	APIURL     string
	HTTPClient HTTPDoer
	APIKey     string
	AuthToken  string

	// Version is the Kion release the connected instance reported via
	// GET /api/version, populated at Configure time by DetectVersion.
	// VersionDetected is false if detection failed (Version is then zero).
	Version         KionVersion
	VersionDetected bool

	// BudgetMode is whether the instance funds projects with budgets rather
	// than spend plans, read from GET /v3/app-config at Configure time by
	// DetectFinancialMode. FinancialModeDetected is false if that read failed,
	// which it may for a token without the global settings permission.
	BudgetMode            bool
	FinancialModeDetected bool
}

// RawDelete performs a raw HTTP DELETE against the Kion API for endpoints not
// covered by the generated SDK (e.g., DELETE /v2/ou/{id}).
func (c *KionClient) RawDelete(ctx context.Context, path string) (err error) {
	url := strings.TrimRight(c.APIURL, "/") + path

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
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
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned status %d for DELETE %s", resp.StatusCode, path)
	}
	return nil
}

// RawHTTPError is returned by the raw request helpers on a >= 400 status. It
// carries the status code so callers can distinguish e.g. 404 (gone) for Read.
type RawHTTPError struct {
	Method, Path string
	StatusCode   int
	Body         []byte
}

func (e *RawHTTPError) Error() string {
	return fmt.Sprintf("API returned status %d for %s %s: %s", e.StatusCode, e.Method, e.Path, string(e.Body))
}

// IsRawNotFound reports whether err is a RawHTTPError with a 404 status.
func IsRawNotFound(err error) bool {
	var e *RawHTTPError
	return errors.As(err, &e) && e.StatusCode == http.StatusNotFound
}

// rawRequest performs a raw HTTP call against the Kion API for endpoints the
// generated SDK does not cover, returning the response body.
func (c *KionClient) rawRequest(ctx context.Context, method, path string, body []byte) (_ []byte, err error) {
	url := strings.TrimRight(c.APIURL, "/") + path

	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rdr)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
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
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	respBody, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		return nil, fmt.Errorf("reading response: %w", rerr)
	}
	if resp.StatusCode >= 400 {
		return respBody, &RawHTTPError{Method: method, Path: path, StatusCode: resp.StatusCode, Body: respBody}
	}
	return respBody, nil
}

// RawGet performs a raw HTTP GET for an endpoint not covered by the SDK.
func (c *KionClient) RawGet(ctx context.Context, path string) ([]byte, error) {
	return c.rawRequest(ctx, http.MethodGet, path, nil)
}

// RawPost performs a raw HTTP POST with a JSON body.
func (c *KionClient) RawPost(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.rawRequest(ctx, http.MethodPost, path, body)
}

// RawPatch performs a raw HTTP PATCH with a JSON body.
func (c *KionClient) RawPatch(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.rawRequest(ctx, http.MethodPatch, path, body)
}

// RawPut performs a raw HTTP PUT with a JSON body.
func (c *KionClient) RawPut(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.rawRequest(ctx, http.MethodPut, path, body)
}
