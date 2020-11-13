package cmd

import (
	"fmt"
	"sort"

	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the last batch of migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := lib.NewDatabase()
		if err != nil {
			return err
		}

		markers, err := db.GetLastNMarkers(2)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		migrations := lib.NewMigrations()
		sort.Sort(sort.Reverse(migrations))

		err = migrations.Slice(markers)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if err := migrations.Execute2(lib.DOWN, db); err != nil {
			return err
		}
		return nil
	},
}
