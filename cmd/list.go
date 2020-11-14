package cmd

import (
	"fmt"

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

func listMigrations(direction lib.Direction) error {
	migrations := lib.NewMigrations()

	db, err := lib.NewDatabase()
	if err != nil {
		return err
	}

	currentMigration, err := db.GetHashForMarkerN(1)
	if err != nil {
		return err
	}

	if err = migrations.Range(currentMigration, -1, direction); err != nil {
		return err
	}

	for _, d := range migrations {
		fmt.Println(d.Hash, d.Path)
	}
	return nil
}
