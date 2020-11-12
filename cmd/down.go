package cmd

import (
	"strconv"

	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down [steps]",
	Short: "Run one or more down migrations",
	RunE: func(cmd *cobra.Command, args []string) error {

		var steps int
		if len(args) > 0 {
			steps, _ = strconv.Atoi(args[0])
		}

		db, err := lib.NewDatabase()
		if err != nil {
			return err
		}

		migrations := lib.NewMigrations()

		currentMigration, err := db.GetHashForMarkerN(1)
		if err != nil {
			return err
		}

		start, stop, err := migrations.FindMigrationRangeDown(currentMigration, steps)
		if err != nil {
			return err
		}

		if err := migrations.Execute(start, stop, lib.DOWN, db); err != nil {
			return err
		}
		return nil
	},
	Args: StepValidator,
}
