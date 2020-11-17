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

func listMigrations(direction int) error {
	migrations := lib.NewMigrations()

	db, err := lib.NewDatabase()
	if err != nil {
		return err
	}

	batch, err := db.LastBatch()
	if err != nil {
		return err
	}

	if err = migrations.Slice(batch.Batch, -1, direction); err != nil {
		return err
	}

	for _, d := range migrations {
		fmt.Println(d.Hash, d.Path)
	}
	return nil
}
