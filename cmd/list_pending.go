package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(listPendingCmd)
}

var listPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		migrations, start, err := listingData()
		if err != nil {
			return err
		}

		for _, d := range migrations[start:] {
			fmt.Println(d.Hash, d.Path)
		}
		return nil
	},
}
