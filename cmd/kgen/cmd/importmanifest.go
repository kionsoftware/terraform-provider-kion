// Unlike module.go and examples.go, this command reads codegen YAML rather than
// compiled provider schemas, so it carries NO kgendocs build tag -- which lets
// its drift test run in the ordinary CI unit job.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"terraform-provider-kion/internal/kgen/importmanifest"
	"terraform-provider-kion/internal/kgen/kfs"
)

var importManifestRoot string

var importManifestCmd = &cobra.Command{
	Use:   "import-manifest",
	Short: "Generate codegen/import_manifest.json for the kion-import tool",
	Long: `Derive, from codegen/generator_config.yaml + crud_archetypes.yaml + the
schema snapshot, everything an enumerator needs to walk a live Kion install and
emit Terraform import blocks: the list endpoint per resource, the read shape,
and the import-id format that matches the ImportState the crud templates
generate.

The output is embedded into kion-import, so regenerate it whenever an archetype
or a read path changes.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		root := importManifestRoot
		if root == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			root = cwd
		}
		m, err := importmanifest.Generate(kfs.OS{}, root)
		if err != nil {
			return err
		}
		readable := 0
		for _, r := range m.Resources {
			if r.Readable {
				readable++
			}
		}
		fmt.Printf("Wrote %s: %d resources, %d readable\n",
			importmanifest.OutputPath, len(m.Resources), readable)
		return nil
	},
}

func init() {
	importManifestCmd.Flags().StringVar(&importManifestRoot, "root", "",
		"project root (default: current directory)")
	rootCmd.AddCommand(importManifestCmd)
}
