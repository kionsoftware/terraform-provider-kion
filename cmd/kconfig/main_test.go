package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
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
	code := run([]string{"gen", "-nonexistent"}, &out, &errb)
	assert.Equal(t, 2, code)
}

func TestRun_genError(t *testing.T) {
	t.Parallel()

	// A nonexistent spec makes config.Gen fail while reading operations.
	var out, errb bytes.Buffer
	code := run([]string{"gen", "--spec", "/nonexistent-spec.json"}, &out, &errb)
	assert.Equal(t, 1, code)
	assert.Contains(t, errb.String(), "error:")
}

func TestRun_genWriteError(t *testing.T) {
	t.Parallel()

	var out, errb bytes.Buffer
	code := run([]string{"gen", "--write", "--spec", "/nonexistent-spec.json"}, &out, &errb)
	assert.Equal(t, 1, code)
	assert.Contains(t, errb.String(), "error:")
}

func TestRun_checkError(t *testing.T) {
	t.Parallel()

	var out, errb bytes.Buffer
	code := run([]string{"check", "--spec", "/nonexistent-spec.json"}, &out, &errb)
	assert.Equal(t, 1, code)
	assert.Contains(t, errb.String(), "error:")
}
