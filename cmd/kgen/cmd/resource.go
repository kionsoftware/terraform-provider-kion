package cmd

import (
	"terraform-provider-kion/internal/kgen/resource"

	"github.com/spf13/cobra"
)

var (
	resourceName          string
	resourceSnakeName     string
	resourceClearComments bool
	resourceForce         bool
	resourceIncludeTags   bool
)

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Create scaffolding for a resource",
	RunE: func(_ *cobra.Command, _ []string) error {
		return resource.Create(resourceName, resourceSnakeName, !resourceClearComments, resourceForce, resourceIncludeTags)
	},
}

func init() {
	addScaffoldFlags(resourceCmd, &resourceName, &resourceSnakeName, &resourceClearComments, &resourceForce, &resourceIncludeTags)
	rootCmd.AddCommand(resourceCmd)
}
