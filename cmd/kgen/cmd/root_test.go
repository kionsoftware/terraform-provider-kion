package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findSubcommand returns the child of parent whose Name() (first token of Use)
// matches name, or nil if not registered.
func findSubcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestRootCommand_Metadata(t *testing.T) {
	assert.Equal(t, "kgen [resource|datasource|service]", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short, "root command should have a short description")
	// The root command has no RunE of its own; it only dispatches to subcommands.
	assert.Nil(t, rootCmd.Run, "root should not define a Run")
	assert.Nil(t, rootCmd.RunE, "root should not define a RunE")
}

func TestRootCommand_HasExpectedSubcommands(t *testing.T) {
	// Core commands present in the default (provider-free) build. examples/tests
	// are behind the `kgendocs` build tag, so not asserted here.
	expected := []string{"resource", "datasource", "service", "schemas", "versions", "crud"}

	names := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		names[c.Name()] = true
	}

	for _, name := range expected {
		assert.Truef(t, names[name], "expected subcommand %q to be registered", name)
	}
}

func TestRootCommand_UnknownSubcommandErrors(t *testing.T) {
	// Exercise the dispatch path without invoking any real generation. An
	// unknown subcommand must produce an error rather than silently succeeding.
	// We drive rootCmd directly (rather than Execute(), which calls os.Exit).
	var out, errOut bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	rootCmd.SetArgs([]string{"definitely-not-a-real-command"})
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}
