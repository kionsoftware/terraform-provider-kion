//go:build kgendocs

// Like the examples command, this imports internal/provider (→ every service
// package) to read compiled resource schemas, so it lives behind the `kgendocs`
// build tag. Build with `-tags kgendocs`.
package cmd

import (
	"terraform-provider-kion/internal/kgen/module"

	"github.com/spf13/cobra"
)

var (
	moduleOutDir          string
	moduleProviderVersion string
	moduleFilterResource  string
	moduleForce           bool
)

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Generate a Terraform module per resource from the provider schema",
	Long: `Generate a standalone Terraform module for each registered resource.

Each module is emitted from the compiled Plugin Framework schema, so its
variable names, types, requiredness and descriptions cannot drift from the
provider. Output per module: main.tf, variables.tf, outputs.tf, versions.tf,
README.md (terraform-docs markers), examples/complete and a plan-only test.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return module.Generate(moduleOutDir, moduleProviderVersion, moduleFilterResource, moduleForce)
	},
}

func init() {
	moduleCmd.Flags().StringVar(&moduleOutDir, "out", "", "output directory (default: <project root>/modules)")
	moduleCmd.Flags().StringVar(&moduleProviderVersion, "provider-version", module.DefaultProviderConstraint, "provider version constraint written to versions.tf")
	moduleCmd.Flags().StringVar(&moduleFilterResource, "resource", "", "generate only for a specific resource (e.g., kion_ou)")
	moduleCmd.Flags().BoolVarP(&moduleForce, "force", "f", false, "force creation, overwriting existing modules")
	rootCmd.AddCommand(moduleCmd)
}
