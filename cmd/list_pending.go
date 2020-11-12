package cmd

import (
	"fmt"

	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(listPendingCmd)
}

var listPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		migrations := lib.NewMigrations()

		db, err := lib.NewDatabase()
		if err != nil {
			return err
		}

		currentMigration, err := db.GetHashForMarkerN(1)
		if err != nil {
			return err
		}

		start, _, err := migrations.FindMigrationRangeUp(currentMigration, -1)
		if err != nil {
			return err
		}

		for _, d := range migrations[start:] {
			fmt.Println(d.Hash, d.Path)
		}
		return nil
	},
}
