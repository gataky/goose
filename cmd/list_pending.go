package cmd

import (
	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(listPendingCmd)
}

var listPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List all pending migrations",
	Long: `To get a list of migrations that are ready to be ran run "goose list pending"

6a8f40ecd57b264da0d0492af62b577f626bfbe1 20201023_040000_d_o_solv-20201023_040000_d
76499a490b9c0006100d963e6006f72cf56c6826 20201023_050000_e_o_solv-20201023_050000_e
9ebb39681a4428cc5693ea2d926e5f73711ce9a4 20201023_060000_f_o_solv-20201023_060000_f
cc7eff6ea9e68da4265bc834afda28f9a9db05a8 20201023_070000_g_o_solv-20201023_070000_g

The output order will be from the oldest migration that hasn't been ran to the newest.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listMigrations(lib.Up)
	},
}
