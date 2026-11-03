package crud

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/kfs"
)

// Companion tests are emitted with nil data; each must render against the
// default companion data, or generation silently skips the resource.
func TestCompanionTestsRenderWithNilData(t *testing.T) {
	sdkVersion = "v3_17"
	for name, files := range companionTestsByName {
		g := &generator{fs: kfs.OS{}}
		_, err := g.emitBespoke(t.TempDir(), name, nil, files, true)
		require.NoError(t, err, name)
	}
}

func TestEmitBespokeFailsOnBrokenTemplate(t *testing.T) {
	sdkVersion = "v3_17"
	g := &generator{fs: kfs.OS{}}
	_, err := g.emitBespoke(t.TempDir(), "thing", nil,
		[]bespokeFile{{tmpl: "package x\n{{.NoSuchField}}", outName: "x.go"}}, true)
	require.Error(t, err)
}

func TestFailuresNameEveryResource(t *testing.T) {
	g := &generator{}
	require.NoError(t, g.failureError())
	g.recordFailure("alpha", errors.New("boom"))
	g.recordFailure("beta", errors.New("bang"))
	err := g.failureError()
	require.Error(t, err)
	require.Contains(t, err.Error(), "alpha")
	require.Contains(t, err.Error(), "boom")
	require.Contains(t, err.Error(), "beta")
}

// Without --force an existing file is a deliberate skip, not a failure.
func TestExistingFileWithoutForceIsNotAFailure(t *testing.T) {
	sdkVersion = "v3_17"
	dir := t.TempDir()
	g := &generator{fs: kfs.OS{}}
	files := []bespokeFile{{tmpl: "package x\n", outName: "x.go"}}
	_, err := g.emitBespoke(dir, "thing", nil, files, true)
	require.NoError(t, err)
	_, err = g.emitBespoke(dir, "thing", nil, files, false)
	require.ErrorIs(t, err, errExists)
	require.False(t, g.recordUnlessExists("thing", err))
	require.NoError(t, g.failureError())
	require.True(t, g.recordUnlessExists("thing", errors.New("boom")))
	require.Error(t, g.failureError())
}
