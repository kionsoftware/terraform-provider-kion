// Command kmigrate rewrites Terraform .tf configuration from the old (SDKv2) Kion
// provider's attribute names to the new (Plugin Framework) provider's, driven by
// codegen/state_upgrades.yaml (the same mapping the provider's state upgraders
// use). It handles the input-attribute half of migration, attribute renames and
// block-set → id-list restructures. So a customer's config keeps working after
// the provider version bump. State migrates automatically in-provider; see
// docs/guides/upgrading-from-0.3.md.
//
// Usage:
//
//	kmigrate [--check] [--upgrades <path>] <dir-or-file> [<dir-or-file>...]
//	kmigrate --version
//
//	--check      print the changes without writing (dry run)
//	--upgrades   path to state_upgrades.yaml (default: built into the binary)
//	--version    print the release this binary was built from
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"terraform-provider-kion/codegen"
	"terraform-provider-kion/internal/kgen/migrate"
)

// version is the release, set by goreleaser's ldflags; a source build is "dev".
// kmigrate rewrites for one provider release's schema, so which one matters.
var version = "dev"

func main() {
	check := flag.Bool("check", false, "print changes without writing (dry run)")
	upPath := flag.String("upgrades", "", "path to state_upgrades.yaml (default: the copy built into this binary)")
	showVersion := flag.Bool("version", false, "print the release this binary was built from")
	flag.Parse()
	if *showVersion {
		fmt.Println("kmigrate version " + version)
		return
	}
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: kmigrate [--check] [--upgrades <path>] <dir-or-file>...")
		os.Exit(2)
	}

	// Default to the embedded ruleset. The previous default was the
	// repo-relative codegen/state_upgrades.yaml, which does not exist in a
	// customer's Terraform directory, so the first documented command failed
	// for everyone outside a clone.
	var (
		ups map[string]migrate.Transform
		err error
	)
	if *upPath == "" {
		ups, err = migrate.ParseUpgrades(codegen.StateUpgradesYAML, "embedded state_upgrades.yaml")
	} else {
		ups, err = migrate.LoadUpgrades(*upPath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "kmigrate: %v\n", err)
		os.Exit(1)
	}

	var files []string
	for _, arg := range flag.Args() {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "kmigrate: %v\n", err)
			os.Exit(1)
		}
		if info.IsDir() {
			if err := filepath.WalkDir(arg, func(p string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(p, ".tf") {
					files = append(files, p)
				}
				return nil
			}); err != nil {
				fmt.Fprintf(os.Stderr, "kmigrate: %v\n", err)
				os.Exit(1)
			}
		} else {
			files = append(files, arg)
		}
	}

	total := 0
	var manual []string
	// Files kmigrate could not read or parse. Collected rather than fatal: this
	// runs over a practitioner's whole configuration, and exiting on the first
	// bad file left the tree HALF migrated, everything before it rewritten,
	// everything after it untouched, with no record of which was which. A single
	// unparseable file (Terraform itself rejects them too) should not put a
	// configuration in a state neither provider can read. Skipped files are left
	// exactly as they were and listed at the end.
	var skipped []string
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v", f, err))
			continue
		}
		out, changes, actions, err := migrate.RewriteNamedFile(f, src, ups)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v", f, err))
			continue
		}
		for _, a := range actions {
			manual = append(manual, fmt.Sprintf("%s: %s", f, a))
		}
		if len(changes) == 0 {
			continue
		}
		total += len(changes)
		for _, c := range changes {
			fmt.Printf("%s: %s\n", f, c)
		}
		if !*check {
			if err := os.WriteFile(f, out, 0o644); err != nil {
				skipped = append(skipped, fmt.Sprintf("%s: %v", f, err))
				continue
			}
		}
	}

	if *check {
		fmt.Printf("\n%d change(s) would be made (dry run; re-run without --check to apply).\n", total)
	} else {
		fmt.Printf("\n%d change(s) applied. Run `terraform fmt` to normalize.\n", total)
	}

	// Printed after the summary, and counted separately: these are not edits
	// kmigrate made, they are edits it cannot make. Left unresolved they surface
	// as "Missing required argument" on the next plan.
	if len(manual) > 0 {
		fmt.Printf("\n%d block(s) need attention before `terraform plan` will pass:\n", len(manual))
		for _, m := range manual {
			fmt.Printf("  %s\n", m)
		}
	}

	// Last, and on stderr, because this is the one outcome that leaves work
	// undone: these files were NOT rewritten and still hold old-provider syntax.
	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr,
			"\n%d file(s) could not be read or parsed and were left UNCHANGED:\n", len(skipped))
		for _, s := range skipped {
			fmt.Fprintf(os.Stderr, "  %s\n", s)
		}
		fmt.Fprintf(os.Stderr,
			"\nTerraform cannot parse these either. Fix them, then re-run kmigrate.\n"+
				"Re-running is safe: rewrites are idempotent, so files already migrated are left alone.\n")
		os.Exit(1)
	}
}
