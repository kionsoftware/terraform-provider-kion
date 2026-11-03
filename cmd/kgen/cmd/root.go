// Package cmd implements the kgen CLI commands.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kgen [resource|datasource|service]",
	Short: "Create scaffolding for the Terraform Kion Provider",
}

// addScaffoldFlags registers the flags used by the resource, datasource, and
// service scaffold subcommands. Keeping this in a helper avoids polluting
// unrelated subcommands (e.g. examples, tests) with irrelevant flags.
func addScaffoldFlags(cmd *cobra.Command, name, snakeName *string, clearComments, force, includeTags *bool) {
	cmd.Flags().StringVarP(snakeName, "snakename", "s", "", "explicitly give name in snake case (e.g., cloud_rule)")
	cmd.Flags().BoolVarP(clearComments, "clear-comments", "c", false, "do not include instructional comments in source")
	cmd.Flags().StringVarP(name, "name", "n", "", "name of the entity")
	cmd.Flags().BoolVarP(force, "force", "f", false, "force creation, overwriting existing files")
	cmd.Flags().BoolVarP(includeTags, "include-tags", "t", false, "Indicate that this resource has tags and the code for tagging should be generated")
}

// Execute runs the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
