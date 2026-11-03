package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertScaffoldFlags verifies the shared set of flags registered by
// addScaffoldFlags is present on a scaffold subcommand, along with the
// expected shorthands.
func assertScaffoldFlags(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	cases := []struct {
		name      string
		shorthand string
	}{
		{"snakename", "s"},
		{"clear-comments", "c"},
		{"name", "n"},
		{"force", "f"},
		{"include-tags", "t"},
	}
	for _, tc := range cases {
		f := cmd.Flags().Lookup(tc.name)
		require.NotNilf(t, f, "%s command missing --%s flag", cmd.Name(), tc.name)
		assert.Equalf(t, tc.shorthand, f.Shorthand, "--%s shorthand mismatch on %s", tc.name, cmd.Name())
	}
}

// TestScaffoldCommands_Metadata covers resource, datasource, and service,
// which share the addScaffoldFlags helper. It asserts Use, non-empty Short,
// a wired RunE, and the full flag set, without ever invoking RunE (which
// would write files).
func TestScaffoldCommands_Metadata(t *testing.T) {
	scaffolds := []struct {
		use string
		cmd *cobra.Command
	}{
		{"resource", resourceCmd},
		{"datasource", datasourceCmd},
		{"service", serviceCmd},
	}

	for _, s := range scaffolds {
		t.Run(s.use, func(t *testing.T) {
			assert.Equal(t, s.use, s.cmd.Use)
			assert.NotEmpty(t, s.cmd.Short, "short description should be set")
			assert.NotNil(t, s.cmd.RunE, "RunE should be wired")
			assertScaffoldFlags(t, s.cmd)
		})
	}
}

// TestScaffoldCommands_RegisteredOnRoot confirms each scaffold command is the
// same instance that is attached to rootCmd (guards against a subcommand var
// existing but never being added in init()).
func TestScaffoldCommands_RegisteredOnRoot(t *testing.T) {
	for _, name := range []string{"resource", "datasource", "service"} {
		assert.Same(t, mustFindScaffold(name), findSubcommand(rootCmd, name),
			"%s command registered on root should be the package-level var", name)
	}
}

func mustFindScaffold(name string) *cobra.Command {
	switch name {
	case "resource":
		return resourceCmd
	case "datasource":
		return datasourceCmd
	case "service":
		return serviceCmd
	}
	return nil
}
