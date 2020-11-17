package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [commit hash]",
	Short: "Initializes a migration table in the database called goosey",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		start := ""
		if len(args) == 1 {
			start = args[0]
		}

		if err := db.InitGoosey(start); err != nil {
			return err
		}
		fmt.Println("initialized successfully")
		return nil
	},
}
