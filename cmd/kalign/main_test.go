package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kalign"
)

func TestRun_usage(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{nil, {"bogus"}} {
		var out, errb bytes.Buffer
		code := run(args, &out, &errb)
		assert.Equal(t, 2, code)
		assert.Contains(t, errb.String(), "usage:")
	}
}

func TestRun_badFlag(t *testing.T) {
	t.Parallel()

	var out, errb bytes.Buffer
	code := run([]string{"check", "-nonexistent"}, &out, &errb)
	assert.Equal(t, 2, code)
}

func TestRun_checkError(t *testing.T) {
	t.Parallel()

	// A nonexistent SDK dir makes kalign fail while parsing the SDK types.
	var out, errb bytes.Buffer
	code := run([]string{"check", "-sdk", "/nonexistent-sdk-dir"}, &out, &errb)
	assert.Equal(t, 1, code)
	assert.Contains(t, errb.String(), "error:")
}

func TestGate(t *testing.T) {
	t.Parallel()

	f := []kalign.Finding{{Service: "a", Attr: "x", Kind: kalign.KindType}}
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.yaml")

	var errb bytes.Buffer
	assert.Equal(t, 1, gate(f, path, false, &errb), "missing baseline must fail")

	assert.Equal(t, 0, gate(f, path, true, &errb), "-update writes the baseline")
	assert.Equal(t, 1, gate(f, path, false, &errb), "a fresh TODO entry must fail until triaged")

	require.NoError(t, os.WriteFile(path, []byte("a.x:type: \"accepted\"\n"), 0o644))
	assert.Equal(t, 0, gate(f, path, false, &errb), "baselined finding passes")
	assert.Equal(t, 1, gate(nil, path, false, &errb), "stale entry fails")
	assert.Equal(t, 1, gate(append(f, kalign.Finding{Service: "b", Attr: "y", Kind: kalign.KindFlex}), path, false, &errb), "new finding fails")

	assert.Equal(t, 1, gate(f, "", false, &errb), "no baseline: any finding fails")
	assert.Equal(t, 0, gate(nil, "", false, &errb))
}

func TestRun_sdkResolveError(t *testing.T) {
	orig := resolveSDKDir
	t.Cleanup(func() { resolveSDKDir = orig })
	resolveSDKDir = func() (string, error) { return "", errors.New("no module") }

	var out, errb bytes.Buffer
	assert.Equal(t, 1, run([]string{"check"}, &out, &errb))
	assert.Contains(t, errb.String(), "no module")
}

func TestRun_genError(t *testing.T) {
	t.Parallel()

	var out, errb bytes.Buffer
	code := run([]string{"gen", "-sdk", "/nonexistent-sdk-dir"}, &out, &errb)
	assert.Equal(t, 1, code)
	assert.Contains(t, errb.String(), "error:")
}
