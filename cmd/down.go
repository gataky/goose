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
	Long: `Assuming we're starting with a database that has migrations already in it (a, b and c)

+------------------------------------------+----------+----------+
| hash                                     | author   | marker   |
|------------------------------------------+----------+----------|
| e965f4511fce6ae61e1cfdcf174f61cfd4fe920b | a o      | False    |
| cac4966fa648df678b9f59117d085b40d647ef19 | b o      | False    |
| e0ca0a9d0afe2d168ed09efe2f859f76bcfd109f | c o      | True     |
+------------------------------------------+----------+----------+

Running "goose down 3" will run the last three migrations c, b and a

+------------------------------------------+----------+----------+
| hash                                     | author   | marker   |
|------------------------------------------+----------+----------|
+------------------------------------------+----------+----------+

If you want to remove all the migration you can run "goose down" and that will undo every migration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigration(args, lib.Down)
	},
	Args: stepValidator,
}
