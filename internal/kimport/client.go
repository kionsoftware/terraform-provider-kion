// Package kimport enumerates a live Kion install into Terraform import blocks.
//
// It reads through arbitrary paths supplied by the embedded import manifest
// rather than the typed SDK, because the manifest's unit of work is a path, not
// an SDK method. Auth and transport mirror internal/provider/provider.go.
package kimport

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"terraform-provider-kion/internal/paging"
)

const (
	// pageSize and the walk's page cap live in internal/paging, so this tool and
	// the provider's own raw list reads ask for the same thing.
	pageSize  = paging.DefaultPageSize
	bodyLimit = 512 // max bytes to include in error messages
)

// Lister is the read seam the enumerators depend on, so they can be tested
// without a server.
type Lister interface {
	List(ctx context.Context, path string) ([]map[string]any, error)
	// ListUnder returns the records at data.<key> of a single GET.
	ListUnder(ctx context.Context, path, key string) ([]map[string]any, error)
}

// StatusError is a non-2xx response. Callers need the code, not just the text:
// a 404 from a child collection means the parent has none of that resource,
// which is an expected answer, while a 502 is a real failure.
type StatusError struct {
	Path   string
	Status int
	Body   string // truncated snippet, may be empty
}

func (e *StatusError) Error() string {
	line := fmt.Sprintf("%d %s", e.Status, http.StatusText(e.Status))
	if e.Body != "" {
		return fmt.Sprintf("GET %s: %s: %s", e.Path, line, e.Body)
	}
	return fmt.Sprintf("GET %s: %s", e.Path, line)
}

// Client is a minimal raw-HTTP Kion client.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
	// backoff is the wait before each retry of a transient failure; its length
	// is the retry count.
	backoff []time.Duration
	// authRejected is set once a 401 outlasts its retries: the key is bad, so
	// later 401s fail at once instead of each waiting out the backoff.
	authRejected atomic.Bool
}

// defaultBackoff retries a transient failure twice, after 1s and then 3s.
var defaultBackoff = []time.Duration{time.Second, 3 * time.Second}

// RetryBackoff replaces the waits before each retry (tests pass zeros).
func (c *Client) RetryBackoff(waits ...time.Duration) *Client {
	c.backoff = waits
	return c
}

// retryable reports whether a status is worth another attempt. Kion sometimes
// answers a valid key with a spurious 401; 429 and gateway errors pass. A 500
// is not retried: Kion returns it for deterministic failures.
func retryable(status int) bool {
	switch status {
	case http.StatusUnauthorized, http.StatusTooManyRequests,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}

// NewClient builds a Client. apiPrefix (e.g. "/api") is appended to baseURL,
// tolerating a leading slash either way and never doubling a prefix baseURL
// already ends with. Hosted installs serve their API under "/api" -- pass
// that so callers can hand this a bare install URL unmodified. Pass "" for
// an install that serves the API at the root (e.g. an app hit directly on
// localhost). This mirrors kion-env-copy's KION_API_PREFIX convention.
func NewClient(baseURL, apiKey string, skipSSL bool, apiPrefix string) *Client {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		defaultTransport = &http.Transport{}
	}
	transport := defaultTransport.Clone()
	if skipSSL {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested
	}
	return &Client{
		baseURL: joinAPIPrefix(baseURL, apiPrefix),
		apiKey:  apiKey,
		http:    &http.Client{Transport: transport, Timeout: 60 * time.Second},
		backoff: defaultBackoff,
	}
}

// joinAPIPrefix appends prefix to baseURL. Both sides are normalized of
// leading/trailing slashes before joining, and if baseURL already ends with
// prefix (e.g. a caller passes --url https://host/api with the default
// --api-prefix /api), the prefix is not doubled. An empty (post-trim) prefix
// leaves baseURL untouched.
func joinAPIPrefix(baseURL, prefix string) string {
	base := strings.TrimSuffix(baseURL, "/")
	trimmedPrefix := strings.Trim(prefix, "/")
	if trimmedPrefix == "" {
		return base
	}
	if base == trimmedPrefix || strings.HasSuffix(base, "/"+trimmedPrefix) {
		return base
	}
	return base + "/" + trimmedPrefix
}

// get fetches one page, retrying a transient failure (see retryable) with the
// same request each time.
func (c *Client) get(ctx context.Context, path string, page int) (json.RawMessage, error) {
	for attempt := 0; ; attempt++ {
		body, err := c.getOnce(ctx, path, page)
		var se *StatusError
		if err == nil || !errors.As(err, &se) || !retryable(se.Status) || attempt == len(c.backoff) ||
			(se.Status == http.StatusUnauthorized && c.authRejected.Load()) {
			if se != nil && se.Status == http.StatusUnauthorized {
				c.authRejected.Store(true)
			}
			return body, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.backoff[attempt]):
		}
	}
}

func (c *Client) getOnce(ctx context.Context, path string, page int) (json.RawMessage, error) {
	// A path may carry its own filter (kion_billing_rule's rule_type_ids).
	path, rawQuery, _ := strings.Cut(path, "?")
	q, err := url.ParseQuery(rawQuery)
	if err != nil {
		return nil, fmt.Errorf("list path %q: %w", path, err)
	}
	q.Set("page", fmt.Sprint(page))
	q.Set("count", fmt.Sprint(pageSize))
	u := c.baseURL + path + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // close errors are not actionable

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodySnippet := truncateBody(resp)
		return nil, &StatusError{Path: path, Status: resp.StatusCode, Body: bodySnippet}
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GET %s: read body: %w", path, err)
	}
	var body json.RawMessage
	if err := json.Unmarshal(buf, &body); err != nil {
		if looksLikeHTML(buf) {
			return nil, fmt.Errorf(
				"GET %s: got an HTML page instead of JSON (likely a missing or wrong --api-prefix -- "+
					"the install may serve its API under a different prefix, or at the root if you are "+
					"hitting it directly): %s", path, snippetOf(buf))
		}
		return nil, fmt.Errorf("GET %s: decode: %w", path, err)
	}
	return body, nil
}

// looksLikeHTML reports whether buf, after leading whitespace, starts with
// "<" -- the shape of an HTML error/login page a web app returns with HTTP
// 200 for an unrecognized path, which a plain JSON-decode error would
// otherwise report as an opaque "invalid character '<'" message.
func looksLikeHTML(buf []byte) bool {
	trimmed := bytes.TrimLeft(buf, " \t\r\n")
	return len(trimmed) > 0 && trimmed[0] == '<'
}

// truncateBody reads up to bodyLimit bytes from resp.Body and returns a single-line snippet.
func truncateBody(resp *http.Response) string {
	buf, _ := io.ReadAll(io.LimitReader(resp.Body, int64(bodyLimit)+1)) //nolint:errcheck // body read errors are acceptable here
	return snippetOf(buf)
}

// snippetOf renders buf as a single-line, length-capped preview for error messages.
func snippetOf(buf []byte) string {
	if len(buf) == 0 {
		return ""
	}
	truncated := len(buf) > bodyLimit
	if truncated {
		buf = buf[:bodyLimit]
	}
	// A byte-count truncation can land mid-rune, splitting a multi-byte UTF-8
	// character and turning the tail into invalid UTF-8 that would otherwise
	// flow straight into a customer-facing error message.
	snippet := strings.ToValidUTF8(string(buf), "")
	// Remove newlines and tabs for single-line output
	snippet = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r == '\r' {
			return ' '
		}
		return r
	}, snippet)
	if truncated {
		snippet += "…"
	}
	return strings.TrimSpace(snippet)
}

// unwrap pulls records out of a bare list or an {items|data,total} envelope,
// and reports the envelope's total (-1 when there is none).
func unwrap(body json.RawMessage) ([]map[string]any, int, error) {
	var list []map[string]any
	if err := json.Unmarshal(body, &list); err == nil {
		return list, -1, nil
	}

	var env struct {
		Items json.RawMessage `json:"items"`
		Data  json.RawMessage `json:"data"`
		Total *int            `json:"total"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, -1, err
	}

	total := -1
	isEnvelope := false
	if env.Total != nil {
		total = *env.Total
		isEnvelope = true
	}

	// Handle items field (may be null or empty)
	if len(env.Items) > 0 {
		var records []map[string]any
		if err := json.Unmarshal(env.Items, &records); err == nil {
			return records, total, nil
		}
		// A singleton endpoint returns one object rather than a list.
		var single map[string]any
		if err := json.Unmarshal(env.Items, &single); err == nil {
			return []map[string]any{single}, total, nil
		}
	}

	// Handle data field (may be null or empty)
	if len(env.Data) > 0 {
		var records []map[string]any
		if err := json.Unmarshal(env.Data, &records); err == nil {
			return records, total, nil
		}
		// A doubly-nested envelope: {"status":200,"data":{"items":[...],"total":N,...}}.
		// Some endpoints (e.g. /v4/billing-source, /v3/label, /beta/scope) wrap
		// their items/total envelope inside "data" instead of putting it at the
		// top level. Detect that shape structurally and unwrap the inner
		// envelope recursively, rather than falling through to the
		// single-object branch below and returning the envelope itself as one
		// bogus record.
		if isNestedEnvelope(env.Data) {
			return unwrap(env.Data)
		}
		// A named-collection envelope:
		// {"status":200,"data":{"dashboards":[{...}],"hidden_dashboard_count":0}}.
		// /v1/dashboards wraps its list under a descriptive key, alongside
		// sibling scalar/object metadata, instead of using "items" or a bare
		// array. Detect that shape structurally by counting the data object's
		// keys whose value is itself a JSON array: exactly one such key means
		// that array is the records and every other key (scalar or object) is
		// metadata to ignore. Zero array-valued keys, or two or more, fall
		// through unchanged to the singleton branch below, so a genuine
		// single-object record (e.g. {"data":{"id":1,"name":"x"}}) is still
		// returned as one record.
		if arr, ok := soleArrayValue(env.Data); ok {
			var records []map[string]any
			if err := json.Unmarshal(arr, &records); err == nil {
				return records, total, nil
			}
		}
		// A singleton endpoint returns one object rather than a list.
		var single map[string]any
		if err := json.Unmarshal(env.Data, &single); err == nil {
			return []map[string]any{single}, total, nil
		}
	}

	// If we detected an envelope (total field present), return empty list.
	// Never return the envelope itself as a record.
	if isEnvelope {
		return []map[string]any{}, total, nil
	}

	// A bare singleton object with no envelope.
	var single map[string]any
	if err := json.Unmarshal(body, &single); err == nil && len(single) > 0 {
		return []map[string]any{single}, -1, nil
	}
	return nil, total, nil
}

// isNestedEnvelope reports whether raw is a JSON object carrying an "items"
// key, meaning it is itself an envelope (e.g. the inner object of
// {"status":200,"data":{"pagination":{...},"total":17,"items":[...]}}) rather
// than a genuine single-object record. The key is checked for presence, not
// truthiness, so {"items":null} still counts -- that's an empty envelope, not
// a record named "items".
func isNestedEnvelope(raw json.RawMessage) bool {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return false
	}
	_, ok := obj["items"]
	return ok
}

// soleArrayValue reports whether raw is a JSON object with exactly one
// array-valued key -- regardless of how many other, non-array keys it also
// has -- returning that array's raw bytes when so. Used to detect a
// named-collection envelope like {"dashboards":[...],"hidden_count":0} that
// wraps a list under a descriptive key, alongside sibling scalar/object
// metadata, rather than using "items" or a bare array.
func soleArrayValue(raw json.RawMessage) (json.RawMessage, bool) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, false
	}
	var found json.RawMessage
	arrayKeys := 0
	for _, v := range obj {
		trimmed := bytes.TrimLeft(v, " \t\r\n")
		if len(trimmed) > 0 && trimmed[0] == '[' {
			arrayKeys++
			found = v
		}
	}
	if arrayKeys != 1 {
		return nil, false
	}
	return found, true
}

// List GETs path, unwrapping and paging as needed.
//
// The walk itself is internal/paging; this supplies the unwrapping, which is
// specific to the shapes the import tool meets (bare arrays, named collections,
// doubly-nested envelopes) and to the padding records some collections return.
// dropBlanks runs per page, BEFORE paging compares what it holds against the
// reported total: counting padding as records made page 1 look complete and
// silently lost every record only reachable on page 2.
func (c *Client) List(ctx context.Context, path string) ([]map[string]any, error) {
	return paging.All(ctx, "GET "+path, func(ctx context.Context, page int) (paging.Page[map[string]any], error) {
		body, err := c.get(ctx, path, page)
		if err != nil {
			return paging.Page[map[string]any]{}, err
		}
		records, total, err := unwrap(body)
		if err != nil {
			if page == 1 {
				return paging.Page[map[string]any]{}, fmt.Errorf("GET %s: %w", path, err)
			}
			return paging.Page[map[string]any]{}, fmt.Errorf("GET %s (page %d): %w", path, page, err)
		}
		return paging.Page[map[string]any]{Items: dropBlanks(records), Total: total}, nil
	})
}

// ListUnder GETs path once and returns the records under data.<key>. It is for
// a response whose data object holds several arrays, where List's structural
// unwrapping would pick the wrong one or return the object itself. A null or
// absent key is no records.
func (c *Client) ListUnder(ctx context.Context, path, key string) ([]map[string]any, error) {
	body, err := c.get(ctx, path, 1)
	if err != nil {
		return nil, err
	}
	var env struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("GET %s: %w", path, err)
	}
	raw, ok := env.Data[key]
	if !ok || string(bytes.TrimSpace(raw)) == "null" {
		return []map[string]any{}, nil
	}
	var records []map[string]any
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, fmt.Errorf("GET %s: data.%s: %w", path, key, err)
	}
	return dropBlanks(records), nil
}

// dropBlanks removes zero-valued padding records.
//
// Some collections pad every page out to `total` rather than to the page size:
// GET /v4/compliance/program/5/family?count=100 returns 110 items on page 1 --
// the 100 real ones plus 10 all-zero fillers -- and 110 again on page 2, that
// time 10 real and 100 fillers. Counting the padding as records made page 1
// look complete (`len(records) < total` was already false), paging stopped, and
// the 10 real records only reachable on page 2 were never fetched. They then
// surfaced as "10 record(s) skipped: no id", which reads like junk being
// discarded rather than real records going missing.
//
// A record whose every value is a zero value carries no id, so it could never
// have produced an import block; dropping it loses nothing and lets the page
// count reflect what was actually retrieved.
func dropBlanks(records []map[string]any) []map[string]any {
	out := records[:0]
	for _, rec := range records {
		if !isBlankRecord(rec) {
			out = append(out, rec)
		}
	}
	return out
}

// isBlankRecord reports whether every value in rec is a zero value.
func isBlankRecord(rec map[string]any) bool {
	if len(rec) == 0 {
		return true
	}
	for _, v := range rec {
		switch t := v.(type) {
		case nil:
		case string:
			if t != "" {
				return false
			}
		case bool:
			if t {
				return false
			}
		case float64:
			if t != 0 {
				return false
			}
		case []any:
			if len(t) != 0 {
				return false
			}
		case map[string]any:
			if !isBlankRecord(t) {
				return false
			}
		default:
			return false
		}
	}
	return true
}
