package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(makeCmd)
}

var makeCmd = &cobra.Command{
	Use:   "make (first_name) (last_name) (message)",
	Short: "!Make a new migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("make")
		return nil
	},
	Args: cobra.ExactArgs(3),
}
