package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"terraform-provider-kion/internal/kgen/crud"
	"terraform-provider-kion/internal/kgen/versions"
)

var (
	crudConfig          string
	crudConfigOverrides string
	crudSDK             string
	crudSDKVersion      string
	crudCrudOverrides   string
	crudArchetypes      string
	crudTestValues      string
	crudVersionSupport  string
	crudForce           bool
	crudResource        string
	crudStrict          bool
)

var crudCmd = &cobra.Command{
	Use:   "crud",
	Short: "Generate CRUD methods, data sources, tests, and sweepers for service packages",
	RunE: func(cmd *cobra.Command, _ []string) error {
		n, err := crud.Generate(crud.Options{
			Config:          crudConfig,
			ConfigOverrides: crudConfigOverrides,
			SDKDir:          crudSDK,
			SDKVersion:      crudSDKVersion,
			CrudOverrides:   crudCrudOverrides,
			Archetypes:      crudArchetypes,
			TestValues:      crudTestValues,
			VersionSupport:  crudVersionSupport,
			Force:           crudForce,
			OnlyResource:    crudResource,
			Strict:          crudStrict,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "generated CRUD for %d resource(s)\n", n)
		return nil
	},
}

func init() {
	crudCmd.Flags().StringVar(&crudConfig, "config", "codegen/generator_config.yaml", "path to generator config")
	crudCmd.Flags().StringVar(&crudConfigOverrides, "config-overrides", "codegen/config_overrides.yaml", "path to config overrides")
	crudCmd.Flags().StringVar(&crudSDK, "sdk", "../kion-sdk-go", "path to kion-sdk-go")
	crudCmd.Flags().StringVar(&crudSDKVersion, "sdk-version", versions.CurrentSDK(), "SDK version subpackage")
	crudCmd.Flags().StringVar(&crudCrudOverrides, "crud-overrides", "codegen/crud_overrides.yaml", "path to CRUD overrides")
	crudCmd.Flags().StringVar(&crudArchetypes, "archetypes", "codegen/crud_archetypes.yaml", "path to compound-key/parent-read archetype declarations")
	crudCmd.Flags().StringVar(&crudTestValues, "test-values", "codegen/test_values.yaml", "path to acceptance-test sample values")
	crudCmd.Flags().StringVar(&crudVersionSupport, "version-support", "codegen/version_support.yaml", "path to version-support gates")
	crudCmd.Flags().BoolVarP(&crudForce, "force", "f", false, "overwrite existing files")
	crudCmd.Flags().StringVar(&crudResource, "resource", "", "generate only this resource")
	crudCmd.Flags().BoolVar(&crudStrict, "strict", false, "exit non-zero if any data source was downgraded from list+filter to id-only")
	rootCmd.AddCommand(crudCmd)
}
