package cmd

import (
	"fmt"
	"sort"

	"terraform-provider-kion/internal/kgen/migrate"

	"github.com/spf13/cobra"
)

var (
	migrateOldSnap string
	migrateNewSnap string
	migrateOnly    string
)

var migrateDiffCmd = &cobra.Command{
	Use:   "migrate-diff",
	Short: "Print the old→new provider attribute delta per resource (seeds codegen/state_upgrades.yaml)",
	Long: "Diffs the two committed `terraform providers schema -json` snapshots and prints, per " +
		"old resource type, the attributes dropped / added / type-changed and whether the id type " +
		"changed. Use it to author and drift-check codegen/state_upgrades.yaml.",
	RunE: func(_ *cobra.Command, _ []string) error {
		oldS, err := migrate.LoadSchema(orDefault(migrateOldSnap, "codegen/schema_snapshots/old.json"))
		if err != nil {
			return err
		}
		newS, err := migrate.LoadSchema(orDefault(migrateNewSnap, "codegen/schema_snapshots/new.json"))
		if err != nil {
			return err
		}
		deltas := migrate.Diff(oldS, newS)
		types := make([]string, 0, len(deltas))
		for t := range deltas {
			if migrateOnly == "" || t == migrateOnly {
				types = append(types, t)
			}
		}
		sort.Strings(types)
		for _, t := range types {
			d := deltas[t]
			if len(d.Dropped) == 0 && len(d.Added) == 0 && len(d.TypeChanged) == 0 && !d.IDTypeChanged {
				fmt.Printf("%s: (no change)\n", t)
				continue
			}
			fmt.Printf("%s:\n", t)
			if d.IDTypeChanged {
				fmt.Printf("  id: type changed (string→number)\n")
			}
			if len(d.TypeChanged) > 0 {
				fmt.Printf("  type-changed: %v\n", d.TypeChanged)
			}
			if len(d.Dropped) > 0 {
				fmt.Printf("  dropped:      %v\n", d.Dropped)
			}
			if len(d.Added) > 0 {
				fmt.Printf("  added:        %v\n", d.Added)
			}
		}
		return nil
	},
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func init() {
	migrateDiffCmd.Flags().StringVar(&migrateOldSnap, "old", "", "old provider schema snapshot (default codegen/schema_snapshots/old.json)")
	migrateDiffCmd.Flags().StringVar(&migrateNewSnap, "new", "", "new provider schema snapshot (default codegen/schema_snapshots/new.json)")
	migrateDiffCmd.Flags().StringVar(&migrateOnly, "resource", "", "limit to a single resource type (e.g. kion_user_group)")
	rootCmd.AddCommand(migrateDiffCmd)
}
