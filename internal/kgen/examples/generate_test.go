package examples

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/kfs/mocks"
)

// swapFS installs a fresh MockFS in place of the package-level fsw seam for the
// duration of the test, restoring the original on cleanup.
func swapFS(t *testing.T) *mocks.MockFS {
	t.Helper()
	m := mocks.NewMockFS(t)
	orig := fsw
	fsw = m
	t.Cleanup(func() { fsw = orig })
	return m
}

// statFindsGoMod makes findProjectRoot succeed: any path ending in "go.mod"
// reports as existing (so the project root resolves to the current dir), while
// every other Stat reports not-exist (so skip-if-exists checks proceed to
// write). It returns the resolved go.mod parent directory captured on first hit.
func statFindsGoMod(m *mocks.MockFS) {
	m.EXPECT().Stat(mock.Anything).RunAndReturn(func(name string) (fs.FileInfo, error) {
		if strings.HasSuffix(name, "go.mod") {
			return nil, nil
		}
		return nil, os.ErrNotExist
	}).Maybe()
}

// capturedWrites records every WriteFile the generator performs.
type capturedWrites struct {
	mu    sync.Mutex
	files []string
}

func (c *capturedWrites) add(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files = append(c.files, name)
}

func (c *capturedWrites) list() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.files))
	copy(out, c.files)
	return out
}

func TestGenerate_FilterResource_WritesOnlyMatching(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)
	statFindsGoMod(m)

	var mkdirs capturedWrites
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).RunAndReturn(
		func(path string, _ fs.FileMode) error {
			mkdirs.add(path)
			return nil
		}).Maybe()

	var writes capturedWrites
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(name string, _ []byte, _ fs.FileMode) error {
			writes.add(name)
			return nil
		}).Maybe()

	err := Generate(false, "kion_ou_note")
	require.NoError(t, err)

	files := writes.list()
	require.NotEmpty(t, files, "expected at least one file written for the filtered resource")

	// Every written file must belong to the ou_note resource or data source.
	for _, f := range files {
		require.Contains(t, f, "ou_note", "unexpected file written for a non-matching resource: %s", f)
	}

	// Both the resource and the data source for ou_note should be generated.
	var sawResource, sawDataSource bool
	for _, f := range files {
		if strings.HasSuffix(f, filepath.Join("resources", "kion_ou_note", "resource.tf")) {
			sawResource = true
		}
		if strings.HasSuffix(f, filepath.Join("data-sources", "kion_ou_note", "data-source.tf")) {
			sawDataSource = true
		}
	}
	require.True(t, sawResource, "expected resource.tf for kion_ou_note, got %v", files)
	require.True(t, sawDataSource, "expected data-source.tf for kion_ou_note, got %v", files)

	// Directories created should mirror the written files.
	require.NotEmpty(t, mkdirs.list())
}

func TestGenerate_NoMatch_ReturnsError(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)
	statFindsGoMod(m)

	// No resource matches, so nothing is written and Generate must error.
	err := Generate(false, "kion_does_not_exist_anywhere")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no resource or data source found")
}

func TestGenerate_SkipsExistingWhenNotForced(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	// go.mod resolves the project root; the output .tf files "already exist".
	m.EXPECT().Stat(mock.Anything).RunAndReturn(func(name string) (fs.FileInfo, error) {
		if strings.HasSuffix(name, "go.mod") {
			return nil, nil
		}
		// Existing output file -> skip path (no MkdirAll/WriteFile expected).
		return nil, nil
	}).Maybe()

	// If the generator tried to write while files exist and force=false, these
	// unexpected calls would fail the mock. We intentionally register no
	// MkdirAll/WriteFile expectations.
	err := Generate(false, "kion_ou_note")
	require.NoError(t, err)
}

func TestGenerate_ForceOverwritesExisting(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	// Even when the output files exist, force=true must still write them.
	// findProjectRoot needs Stat(go.mod); with force=true the skip-check Stat is
	// not called, so only go.mod stats occur.
	m.EXPECT().Stat(mock.Anything).RunAndReturn(func(name string) (fs.FileInfo, error) {
		if strings.HasSuffix(name, "go.mod") {
			return nil, nil
		}
		return nil, nil
	}).Maybe()

	var writes capturedWrites
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(nil).Maybe()
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(name string, _ []byte, _ fs.FileMode) error {
			writes.add(name)
			return nil
		}).Maybe()

	err := Generate(true, "kion_ou_note")
	require.NoError(t, err)
	require.NotEmpty(t, writes.list(), "force=true should overwrite even existing files")
}

func TestGenerate_MkdirAllErrorPropagates(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)
	statFindsGoMod(m)

	sentinel := errors.New("mkdir boom")
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(sentinel).Maybe()
	// WriteFile should never be reached, but tolerate it in case of ordering.
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	err := Generate(true, "kion_ou_note")
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Contains(t, err.Error(), "creating directory")
}

func TestGenerate_WriteFileErrorPropagates(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)
	statFindsGoMod(m)

	sentinel := errors.New("write boom")
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(nil).Maybe()
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).Return(sentinel).Maybe()

	err := Generate(true, "kion_ou_note")
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Contains(t, err.Error(), "writing")
}

func TestGenerate_ProjectRootNotFoundPropagates(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	// No go.mod anywhere -> findProjectRoot walks to filesystem root and fails.
	m.EXPECT().Stat(mock.Anything).Return(nil, os.ErrNotExist).Maybe()

	err := Generate(false, "kion_ou_note")
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not find project root")
}

func TestGenerate_AllResources_WritesManyFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)
	statFindsGoMod(m)

	var writes capturedWrites
	m.EXPECT().MkdirAll(mock.Anything, mock.Anything).Return(nil).Maybe()
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(name string, data []byte, _ fs.FileMode) error {
			require.NotEmpty(t, data, "written .tf content should not be empty for %s", name)
			writes.add(name)
			return nil
		}).Maybe()

	err := Generate(false, "")
	require.NoError(t, err)

	files := writes.list()
	// The provider registers dozens of resources and data sources; sanity-check
	// that generating everything produces a substantial, well-formed set.
	require.Greater(t, len(files), 10, "expected many example files, got %d", len(files))

	var sawResources, sawDataSources bool
	for _, f := range files {
		if strings.Contains(f, filepath.Join("examples", "resources")) && strings.HasSuffix(f, "resource.tf") {
			sawResources = true
		}
		if strings.Contains(f, filepath.Join("examples", "data-sources")) && strings.HasSuffix(f, "data-source.tf") {
			sawDataSources = true
		}
	}
	require.True(t, sawResources, "expected resource examples in output")
	require.True(t, sawDataSources, "expected data-source examples in output")
}
