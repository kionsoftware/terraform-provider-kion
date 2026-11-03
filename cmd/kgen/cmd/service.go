package cmd

import (
	"terraform-provider-kion/internal/kgen/servicepkg"

	"github.com/spf13/cobra"
)

var (
	serviceName          string
	serviceSnakeName     string
	serviceClearComments bool
	serviceForce         bool
	serviceIncludeTags   bool
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Create a complete service package (resource + data source + registration)",
	RunE: func(_ *cobra.Command, _ []string) error {
		return servicepkg.Create(serviceName, serviceSnakeName, !serviceClearComments, serviceForce, serviceIncludeTags)
	},
}

func init() {
	addScaffoldFlags(serviceCmd, &serviceName, &serviceSnakeName, &serviceClearComments, &serviceForce, &serviceIncludeTags)
	rootCmd.AddCommand(serviceCmd)
}
