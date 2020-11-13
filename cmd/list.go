package cmd

import (
	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List pending or executed migrations",
}

func listingData() (lib.Migrations, error) {
	migrations := lib.NewMigrations()

	db, err := lib.NewDatabase()
	if err != nil {
		return nil, err
	}

	currentMigration, err := db.GetHashForMarkerN(1)
	if err != nil {
		return nil, err
	}

	if err = migrations.Range(currentMigration, -1, lib.UP); err != nil {
		return nil, err
	}
	return migrations, nil
}
