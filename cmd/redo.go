package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(redoCmd)
}

var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "!Rollback last batch and perform all migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("redo")
		return nil
	},
}
