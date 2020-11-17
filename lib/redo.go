package lib

import (
	"sort"

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

		sort.Sort(sort.Reverse(migrations))

		if err := migrations.Slice(batch.Hash, batch.Steps, Down); err != nil {
			return err
		}

		if err := migrations.Execute(Down, db, batch.Exclude); err != nil {
			return err
		}

		sort.Sort(migrations)

		if err := migrations.Execute(Up, db, batch.Exclude); err != nil {
			return err
		}
		return nil
	},
}
