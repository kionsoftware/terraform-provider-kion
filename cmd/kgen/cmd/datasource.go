package cmd

import (
	"terraform-provider-kion/internal/kgen/datasource"

	"github.com/spf13/cobra"
)

var (
	datasourceName          string
	datasourceSnakeName     string
	datasourceClearComments bool
	datasourceForce         bool
	datasourceIncludeTags   bool
)

var datasourceCmd = &cobra.Command{
	Use:   "datasource",
	Short: "Create scaffolding for a data source",
	RunE: func(_ *cobra.Command, _ []string) error {
		return datasource.Create(datasourceName, datasourceSnakeName, !datasourceClearComments, datasourceForce, datasourceIncludeTags)
	},
}

func init() {
	addScaffoldFlags(datasourceCmd, &datasourceName, &datasourceSnakeName, &datasourceClearComments, &datasourceForce, &datasourceIncludeTags)
	rootCmd.AddCommand(datasourceCmd)
}
