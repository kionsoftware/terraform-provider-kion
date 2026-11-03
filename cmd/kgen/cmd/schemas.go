package cmd

import (
	"fmt"

	"terraform-provider-kion/internal/kgen/schemas"

	"github.com/spf13/cobra"
)

var (
	schemasConfig    string
	schemasSpec      string
	schemasRenames   string
	schemasOverrides string
	schemasAdditions string
)

var schemasCmd = &cobra.Command{
	Use:   "schemas",
	Short: "Generate resource/data-source schemas (and schema tests) from the OpenAPI spec",
	Long: "Runs HashiCorp's tfplugingen tools against the OpenAPI spec, applies attribute " +
		"renames from the renames sidecar, and writes *_schema_gen.go plus a " +
		"ValidateImplementation unit test into each service package. Requires " +
		"tfplugingen-openapi and tfplugingen-framework on PATH (make install-codegen-tools).",
	RunE: func(_ *cobra.Command, _ []string) error {
		n, err := schemas.Generate(schemas.Options{
			Config:    schemasConfig,
			Spec:      schemasSpec,
			Renames:   schemasRenames,
			Overrides: schemasOverrides,
			Additions: schemasAdditions,
		})
		if err != nil {
			return err
		}
		fmt.Printf("done: %d schema file(s) written. Verify with: go build ./...\n", n)
		return nil
	},
}

func init() {
	schemasCmd.Flags().StringVar(&schemasConfig, "config", "", "tfplugingen-openapi config (default codegen/generator_config.yaml)")
	schemasCmd.Flags().StringVar(&schemasSpec, "spec", "", "OpenAPI 3 spec (default spec/openapi3.json)")
	schemasCmd.Flags().StringVar(&schemasRenames, "renames", "", "attribute-rename sidecar (default codegen/renames.yaml)")
	schemasCmd.Flags().StringVar(&schemasOverrides, "schema-overrides", "", "schema-override sidecar (default codegen/schema_overrides.yaml)")
	schemasCmd.Flags().StringVar(&schemasAdditions, "spec-additions", "", "OpenAPI spec-additions sidecar merged before generation (default codegen/spec_additions.yaml)")
	rootCmd.AddCommand(schemasCmd)
}
