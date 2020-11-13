package cmd

import (
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
		return runMigration(args, lib.DOWN)
	},
	Args: stepValidator,
}
