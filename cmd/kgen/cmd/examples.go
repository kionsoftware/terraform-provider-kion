//go:build kgendocs

// The examples command imports internal/provider (→ every service package) to
// read compiled resource schemas, so it is behind the `kgendocs` build tag. This
// keeps the default kgen binary free of the service packages, so scaffolding and
// schema generation still work when internal/service is empty or not compiling
// (e.g. regenerating it from scratch). Build docs commands with `-tags kgendocs`.
package cmd

import (
	"terraform-provider-kion/internal/kgen/examples"

	"github.com/spf13/cobra"
)

var (
	examplesFilterResource string
	examplesForce          bool
)

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Generate example .tf files from resource/data source schemas",
	RunE: func(_ *cobra.Command, _ []string) error {
		return examples.Generate(examplesForce, examplesFilterResource)
	},
}

func init() {
	examplesCmd.Flags().StringVar(&examplesFilterResource, "resource", "", "generate only for a specific resource (e.g., kion_aws_account)")
	examplesCmd.Flags().BoolVarP(&examplesForce, "force", "f", false, "force creation, overwriting existing files")
	rootCmd.AddCommand(examplesCmd)
}
