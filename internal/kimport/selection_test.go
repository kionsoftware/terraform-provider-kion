package kimport

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/importmanifest"
)

func testRows() []importmanifest.Resource {
	return []importmanifest.Resource{
		{TFType: "kion_label", Readable: true},
		{TFType: "kion_ou", Readable: true},
		{TFType: "kion_azure_policy", Readable: true},
		{TFType: "kion_aws_iam_policy", AliasOf: "kion_iam_policy"},
	}
}

func names(rs []importmanifest.Resource) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.TFType)
	}
	return out
}

func TestSelectionApply(t *testing.T) {
	t.Run("empty selects everything", func(t *testing.T) {
		got, err := Selection{}.Apply(testRows())
		require.NoError(t, err)
		assert.Len(t, got, 4)
	})

	t.Run("include is the whole set", func(t *testing.T) {
		got, err := Selection{Include: []string{"kion_ou", "kion_label"}}.Apply(testRows())
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"kion_ou", "kion_label"}, names(got))
	})

	t.Run("exclude applies after include", func(t *testing.T) {
		s := Selection{Include: []string{"kion_ou", "kion_label"}, Exclude: []string{"kion_label"}}
		got, err := s.Apply(testRows())
		require.NoError(t, err)
		assert.Equal(t, []string{"kion_ou"}, names(got))
	})

	// A name that matches nothing must stop the run. Silently importing
	// everything because of a typo is the failure this guards.
	t.Run("unknown type is an error", func(t *testing.T) {
		_, err := Selection{Include: []string{"kion_labels"}}.Apply(testRows())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kion_labels")
		assert.Contains(t, err.Error(), "--list-types")
	})

	t.Run("unknown exclude is an error too", func(t *testing.T) {
		_, err := Selection{Exclude: []string{"nope"}}.Apply(testRows())
		require.Error(t, err)
	})
}

func TestSelectionMerge(t *testing.T) {
	s := Selection{Include: []string{"kion_ou"}, Exclude: []string{"kion_label"}}
	m := s.Merge([]string{"kion_azure_policy"}, []string{"kion_aws_iam_policy"})
	assert.Equal(t, []string{"kion_ou", "kion_azure_policy"}, m.Include)
	assert.Equal(t, []string{"kion_label", "kion_aws_iam_policy"}, m.Exclude)
	// Merge must not write through to the receiver's backing arrays.
	assert.Equal(t, []string{"kion_ou"}, s.Include)
}

func TestLoadSelection(t *testing.T) {
	dir := t.TempDir()

	t.Run("absent path is empty", func(t *testing.T) {
		s, err := LoadSelection("")
		require.NoError(t, err)
		assert.Empty(t, s.Include)
	})

	t.Run("reads include and exclude", func(t *testing.T) {
		p := filepath.Join(dir, "sel.yaml")
		require.NoError(t, os.WriteFile(p, []byte("include:\n  - kion_ou\nexclude:\n  - kion_label\n"), 0o600))
		s, err := LoadSelection(p)
		require.NoError(t, err)
		assert.Equal(t, []string{"kion_ou"}, s.Include)
		assert.Equal(t, []string{"kion_label"}, s.Exclude)
	})

	// A misspelled key would otherwise be dropped in silence and the file would
	// appear to do nothing.
	t.Run("unknown key is an error", func(t *testing.T) {
		p := filepath.Join(dir, "bad.yaml")
		require.NoError(t, os.WriteFile(p, []byte("includes:\n  - kion_ou\n"), 0o600))
		_, err := LoadSelection(p)
		require.Error(t, err)
	})

	t.Run("missing file is an error", func(t *testing.T) {
		_, err := LoadSelection(filepath.Join(dir, "nope.yaml"))
		require.Error(t, err)
	})
}
