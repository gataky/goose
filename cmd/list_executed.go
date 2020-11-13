package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(listExecutedCmd)
}

var listExecutedCmd = &cobra.Command{
	Use:   "executed",
	Short: "List all executed migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		migrations, start, err := listingData()
		if err != nil {
			return err
		}

		for _, d := range migrations[:start] {
			fmt.Println(d.Hash, d.Path)
		}
		return nil
	},
}
