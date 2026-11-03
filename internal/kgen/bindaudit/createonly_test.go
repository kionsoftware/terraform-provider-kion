package bindaudit_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/bindaudit"
)

// A create_only value comes from the create response, not the config. When the
// attribute is settable, Create discards the practitioner's value, so it is
// reported; a Computed-only one (token) is outside the audit. note is bound
// nowhere and reported for both ops.
func TestCreateOnlyAudit(t *testing.T) {
	t.Parallel()
	findings, err := bindaudit.Run(filepath.Join("testdata", "createonly"))
	require.NoError(t, err)

	var keys []string
	for _, f := range findings {
		keys = append(keys, f.Key())
	}
	require.ElementsMatch(t, []string{"widget.note:create", "widget.note:update", "widget.secret:create"}, keys)
}
