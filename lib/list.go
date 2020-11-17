package lib

import (
	"fmt"

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

	if err := migrations.Slice(batch.Batch, -1, direction); err != nil {
		return err
	}

	for _, d := range migrations {
		fmt.Println(d.Hash, d.Path)
	}
	return nil
}
