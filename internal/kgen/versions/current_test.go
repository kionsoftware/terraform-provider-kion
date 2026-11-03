package versions

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrentSDKIsNewestTracked(t *testing.T) {
	require.Equal(t, TrackedVersions[len(TrackedVersions)-1].Dir, CurrentSDK())
}

// The Makefile's SDK_VERSION feeds refresh-spec's path; it must name the same
// sub-package the generator imports.
func TestMakefileSDKVersionMatches(t *testing.T) {
	b, err := os.ReadFile("../../../Makefile")
	require.NoError(t, err)
	m := regexp.MustCompile(`(?m)^SDK_VERSION \?= (\S+)$`).FindSubmatch(b)
	require.NotNil(t, m, "Makefile has no SDK_VERSION line")
	require.Equal(t, CurrentSDK(), string(m[1]))
}
