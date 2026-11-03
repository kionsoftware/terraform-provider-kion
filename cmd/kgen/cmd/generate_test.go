//go:build kgendocs

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExamplesCommand_Metadata asserts wiring for the examples subcommand
// without running RunE (which generates .tf files).
func TestExamplesCommand_Metadata(t *testing.T) {
	assert.Equal(t, "examples", examplesCmd.Use)
	assert.NotEmpty(t, examplesCmd.Short)
	assert.NotNil(t, examplesCmd.RunE)

	resourceFlag := examplesCmd.Flags().Lookup("resource")
	require.NotNil(t, resourceFlag, "examples should expose --resource")
	assert.Empty(t, resourceFlag.Shorthand, "--resource has no shorthand")

	forceFlag := examplesCmd.Flags().Lookup("force")
	require.NotNil(t, forceFlag, "examples should expose --force")
	assert.Equal(t, "f", forceFlag.Shorthand)

	assert.Same(t, examplesCmd, findSubcommand(rootCmd, "examples"))
}

// TestAllSubcommands_HaveShortDescriptions is a blanket check that every
// registered subcommand carries a non-empty Short (help output hygiene).
func TestAllSubcommands_HaveShortDescriptions(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		// Skip cobra's auto-injected help/completion commands.
		if c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		assert.NotEmptyf(t, c.Short, "subcommand %q should have a Short description", c.Name())
	}
}

// TestForceFlag_DefaultsToFalse verifies the destructive --force flag defaults
// off across every command that exposes it.
func TestForceFlag_DefaultsToFalse(t *testing.T) {
	cmds := []*cobra.Command{resourceCmd, datasourceCmd, serviceCmd, examplesCmd}
	for _, c := range cmds {
		f := c.Flags().Lookup("force")
		require.NotNilf(t, f, "%s should have --force", c.Name())
		assert.Equalf(t, "false", f.DefValue, "%s --force should default to false", c.Name())
	}
}

// TestTestsCommand_Removed guards against the retired schema-driven test
// generator returning: kgen crud owns every acceptance-test file.
func TestTestsCommand_Removed(t *testing.T) {
	assert.Nil(t, findSubcommand(rootCmd, "tests"))
}
