package cmd

import (
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
		return runMigration(args, lib.UP)
	},
	Args: stepValidator,
}
