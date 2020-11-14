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
	Long: `Assuming we're starting with a new database and we want to apply the first three migrations we could 
run "goose up 3" which will run the first three migrations.  In the example case: a, b and c

+------------------------------------------+----------+----------+
| hash                                     | author   | marker   |
|------------------------------------------+----------+----------|
| e965f4511fce6ae61e1cfdcf174f61cfd4fe920b | a o      | False    |
| cac4966fa648df678b9f59117d085b40d647ef19 | b o      | False    |
| e0ca0a9d0afe2d168ed09efe2f859f76bcfd109f | c o      | True     |
+------------------------------------------+----------+----------+

This is referred to as a batch and the last migration in this batch is marked as true to indicate
it's the last one.  More on this when we get to rollbacks and redos

If you want to apply all the migration you can run "goose up" and that will run every migration 
that's remaining.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigration(args, lib.UP)
	},
	Args: stepValidator,
}
