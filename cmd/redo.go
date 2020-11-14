package cmd

import (
	"sort"

	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(redoCmd)
}

var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "Rollback to the last marker and reapply to the current marker",
	Long:  `If you want to rollback and reapply that batch, "goose redo" will do that for you.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := lib.NewDatabase()
		if err != nil {
			return err
		}

		markers, err := db.GetLastNMarkers(2)
		if err != nil {
			return err
		}

		migrations := lib.NewMigrations()
		sort.Sort(sort.Reverse(migrations))

		if err = migrations.Slice(markers); err != nil {
			return err
		}

		if err := migrations.Execute(lib.DOWN, db); err != nil {
			return err
		}

		sort.Sort(migrations)

		if err := migrations.Execute(lib.UP, db); err != nil {
			return err
		}
		return nil
	},
}
