// Like importmanifest.go and unlike module.go, this command reads generated
// JSON rather than compiled provider schemas, so it carries NO kgendocs build
// tag and its tests run in the ordinary CI unit job.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"terraform-provider-kion/internal/kgen/importmodules"
	"terraform-provider-kion/internal/kgen/module"
)

var (
	importModulesIn         string
	importModulesOut        string
	importModulesManifest   string
	importModulesModulesDir string
	importModulesForce      bool
)

var importModulesCmd = &cobra.Command{
	Use:   "import-modules",
	Short: "Rewrite generated resource blocks into module calls",
	// A dangling-reference failure is not a usage mistake, and printing the
	// whole flag list under it buries the one line that matters.
	SilenceUsage: true,
	Long: `Turn the configuration ` + "`terraform plan -generate-config-out`" + ` writes --
bare ` + "`resource \"kion_*\" \"<label>\"`" + ` blocks -- into ` + "`module`" + ` calls against the
modules under modules/terraform-kion-*.

This is the step after an import. kion-import enumerates an install into import
blocks, Terraform writes configuration from the provider's own Read, and this
rewrites that configuration to go through the modules instead of the resources.

The mapping comes from modules/module_manifest.json rather than from variables.tf,
because a module variable is not always named after the provider attribute: kgen
module renames anything colliding with a module block meta-argument, so
kion_cloud_rule's source becomes cloud_rule_source and kion_compliance_program's
version becomes compliance_program_version. Guessing those wrong does not fail
cleanly -- Terraform reads a stray source = "aws" as the module's own source
address and goes looking for a module there.

A resource block whose type has no manifest entry is left byte-for-byte alone,
as is every non-resource block. An attribute with no module variable is dropped
and reported rather than guessed at, so a rewrite never silently changes what is
under management.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		manifest, err := importmodules.LoadManifest(importModulesManifest)
		if err != nil {
			return fmt.Errorf("%w\n\nRun `make modules` if %s is missing",
				err, importModulesManifest)
		}

		src, err := os.ReadFile(importModulesIn)
		if err != nil {
			return err
		}

		rewritten, untouched, err := importmodules.CountKnownBlocks(src, manifest)
		if err != nil {
			return err
		}

		out, warnings, err := importmodules.Rewrite(src, manifest, importModulesModulesDir)
		if err != nil {
			return err
		}

		dangling, err := importmodules.FindDanglingRefs(src, manifest)
		if err != nil {
			return err
		}

		// Checked before writing: a rewrite that clobbers the only copy of a
		// generated configuration is not recoverable by re-running anything
		// cheap -- it costs another plan against a live install.
		if importModulesOut != importModulesIn {
			if _, err := os.Stat(importModulesOut); err == nil && !importModulesForce {
				return fmt.Errorf("%s already exists (use --force to overwrite)", importModulesOut)
			}
		}
		if err := os.WriteFile(importModulesOut, out, 0o644); err != nil { //nolint:gosec // generated config, not a secret
			return err
		}

		w := cmd.OutOrStdout()
		for _, warn := range warnings {
			// Warnings go to stdout with the summary, not stderr: they are part
			// of the result -- "here is what was dropped" -- and a caller
			// redirecting output wants them in the same place as the counts.
			if _, err := fmt.Fprintf(w, "  dropped: %s\n", warn); err != nil {
				return err
			}
		}
		for _, ref := range dangling {
			if _, err := fmt.Fprintf(w, "  dangling: %s\n", ref); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "Wrote %s: %d block(s) rewritten, %d left untouched, %d attribute(s) dropped\n",
			importModulesOut, rewritten, untouched, len(warnings)); err != nil {
			return err
		}
		if len(dangling) > 0 {
			// Loud, and last, because the written file does not load as-is:
			// Terraform rejects a reference to a resource this rewrite removed.
			// Reported rather than repaired -- the replacement depends on the
			// module exposing a matching output, which the manifest has no
			// record of.
			return fmt.Errorf("%d reference(s) point at resources this rewrite converted to modules; "+
				"%s needs those hand-edited to module.<label>.<output> before Terraform will load it",
				len(dangling), importModulesOut)
		}
		return nil
	},
}

func init() {
	importModulesCmd.Flags().StringVar(&importModulesIn, "in", "generated.tf",
		"configuration to rewrite, as written by terraform plan -generate-config-out")
	importModulesCmd.Flags().StringVar(&importModulesOut, "out", "modules.tf",
		"where to write the rewritten configuration; pass the same path as --in to rewrite in place")
	importModulesCmd.Flags().StringVar(&importModulesManifest, "manifest",
		filepath.Join("modules", module.ManifestFileName),
		"module manifest written by make modules")
	importModulesCmd.Flags().StringVar(&importModulesModulesDir, "modules-dir", "./modules",
		"source prefix written into each module block; must stay a local path (a leading ./ is load-bearing)")
	importModulesCmd.Flags().BoolVarP(&importModulesForce, "force", "f", false,
		"overwrite --out if it already exists")
	rootCmd.AddCommand(importModulesCmd)
}
