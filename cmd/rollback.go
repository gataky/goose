package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "!Rollback the last batch of migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("rollback")
		return nil
	},
}
