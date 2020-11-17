package lib

import (
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(listExecutedCmd)
}

var listExecutedCmd = &cobra.Command{
	Use:   "executed",
	Short: "List all executed migrations",
	Long: `To get a list of migrations that have already ran, run "goose list executed"

e0ca0a9d0afe2d168ed09efe2f859f76bcfd109f 20201023_030000_c_o_solv-20201023_030000_c
cac4966fa648df678b9f59117d085b40d647ef19 20201023_020000_b_o_solv-20201023_020000_b
e965f4511fce6ae61e1cfdcf174f61cfd4fe920b 20201023_010000_a_o_solv-20201023_010000_a

The output order will be from most recent migration ran to the oldest.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listMigrations(Down)
	},
}
