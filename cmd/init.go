package cmd

import (
	"fmt"

	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a migration table in the database called goosey",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := lib.NewDatabase()
		if err != nil {
			return err
		}
		err = db.CreateMigrationTable()
		if err == nil {
			fmt.Println("initialized successfully")
		}
		return err
	},
}
