package crud

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var hardSDK = regexp.MustCompile(`kion-sdk-go/generated/v\d+_\d+`)

// A template that names a version outlives the bump that should have moved it.
func TestTemplatesDoNotHardCodeSDKVersion(t *testing.T) {
	require.NoError(t, filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		require.NoError(t, err)
		if d.IsDir() || !strings.HasSuffix(p, ".gtpl") {
			return nil
		}
		b, rerr := os.ReadFile(p)
		require.NoError(t, rerr)
		require.Falsef(t, hardSDK.Match(b), "%s hard-codes an SDK version; use {{.SDKImport}}", p)
		return nil
	}))
}

func TestBuildIndexRequiresVersion(t *testing.T) {
	_, err := buildIndex(nil, t.TempDir(), "")
	require.Error(t, err)
}
