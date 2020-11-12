package cmd

import (
	"strconv"

	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up [steps]",
	Short: "Run one or more up migrations",
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

		start, stop, err := migrations.FindMigrationRangeUp(currentMigration, steps)
		if err != nil {
			return err
		}

		if err := migrations.Execute(start, stop, lib.UP, db); err != nil {
			return err
		}
		return nil
	},
	Args: StepValidator,
}
