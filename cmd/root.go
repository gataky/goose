package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(pendingCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(redoCmd)
}

var rootCmd = &cobra.Command{
	Use:   "goose",
	Short: "A SQL migration tool for Solv",
	Long: `Goose is a SQL migration tool built for Solv.  Goose allows you
to run up and down migrations on a database.`,
}

func Execute() error {
	return rootCmd.Execute()
}

var makeCmd = &cobra.Command{
	Use:   "make (first_name) (last_name) (message)",
	Short: "Make a new migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("make")
		return nil
	},
	Args: cobra.ExactArgs(3),
}

var pendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("pending")
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all executed migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("list")
		return nil
	},
}

var upCmd = &cobra.Command{
	Use:   "up [N]",
	Short: "Run one or more up migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("up")
		return nil
	},
	Args: cobra.MaximumNArgs(1),
}

var downCmd = &cobra.Command{
	Use:   "down [N]",
	Short: "Run one or more down migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("down")
		return nil
	},
	Args: cobra.MaximumNArgs(1),
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the last batch of migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("rollback")
		return nil
	},
}

var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "Rollback last batch and perform all migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("redo")
		return nil
	},
}

/*
a
b
c  -
d
e  -
f
g <-
h
i
*/
