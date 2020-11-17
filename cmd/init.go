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
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		start := ""
		if len(args) == 1 {
			start = args[0]
		}

		db, err := lib.NewDatabase()
		if err != nil {
			return err
		}
		err = db.InitGoosey(start)
		if err == nil {
			fmt.Println("initialized successfully")
		}
		return err
	},
}
