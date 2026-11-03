package cmd

import (
	"fmt"

	"terraform-provider-kion/internal/kgen/versions"

	"github.com/spf13/cobra"
)

var (
	versionsSDK         string
	versionsServiceRoot string
	versionsConfig      string
	versionsOverrides   string
	versionsForce       bool
)

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Generate per-service <name>_version_gen.go version-gate declarations from the SDK",
	Long: "For each resource / data source in the merged generator config whose defining API " +
		"operation (create if present, else read) exists only within a bounded range of tracked " +
		"Kion versions, scans the SDK's per-version generated clients and writes " +
		"internal/service/<name>/<name>_version_gen.go declaring minKionVersion / maxKionVersion " +
		"so the runtime version gate survives schema regeneration.",
	RunE: func(_ *cobra.Command, _ []string) error {
		n, err := versions.Generate(versions.Options{
			SDKDir:      versionsSDK,
			ServiceRoot: versionsServiceRoot,
			ConfigPath:  versionsConfig,
			Overrides:   versionsOverrides,
			Force:       versionsForce,
		})
		if err != nil {
			return err
		}
		fmt.Printf("done: %d version-gate file(s) written. Verify with: go build ./...\n", n)
		return nil
	},
}

func init() {
	versionsCmd.Flags().StringVar(&versionsSDK, "sdk", "", "path to the kion-sdk-go module (default ../kion-sdk-go)")
	versionsCmd.Flags().StringVar(&versionsServiceRoot, "out", "", "output root for <name>/<name>_version_gen.go (default internal/service)")
	versionsCmd.Flags().StringVar(&versionsConfig, "config", "", "generator_config.yaml (default codegen/generator_config.yaml)")
	versionsCmd.Flags().StringVar(&versionsOverrides, "overrides", "", "config_overrides.yaml merged over the config (default codegen/config_overrides.yaml)")
	versionsCmd.Flags().BoolVarP(&versionsForce, "force", "f", false, "overwrite existing _version_gen.go files")
	rootCmd.AddCommand(versionsCmd)
}
