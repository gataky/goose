package cmd

import (
	"fmt"
	"strconv"

	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
)

var (
	migrationPath    = "/home/jeff/Documents/migrations/"
	currentMigration = ""
)

func init() {
	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(pendingCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(redoCmd)
	rootCmd.AddCommand(testCmd)
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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a migration table in the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("init")
		return nil
	},
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
		migrations := lib.NewMigrations(migrationPath)
		start, _, err := migrations.FindMigrationRangeUp(currentMigration, -1)
		if err != nil {
			return err
		}
		for _, d := range migrations[start:] {
			fmt.Println(d.Hash, d.Path)
		}
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all executed migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		migrations := lib.NewMigrations(migrationPath)
		start, _, err := migrations.FindMigrationRangeUp(currentMigration, -1)
		if err != nil {
			return err
		}
		for _, d := range migrations[:start] {
			fmt.Println(d.Hash, d.Path)
		}
		return nil
	},
}

var upCmd = &cobra.Command{
	Use:   "up [steps]",
	Short: "Run one or more up migrations",
	RunE: func(cmd *cobra.Command, args []string) error {

		steps, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid number %s", args[0])
		}

		db, err := lib.NewDatabase("postgres://user:pass@localhost:5432/db?sslmode=disable")
		if err != nil {
			return err
		}

		migrations := lib.NewMigrations(migrationPath)
		err = migrations.Reconcile(db)
		if err != nil {
			return err
		}

		currentMigration, err := db.GetLatestMigration()
		if err != nil {
			return err
		}

		warnings := migrations.FindWarnings(currentMigration)
		if len(warnings) > 0 {
			fmt.Println("\n\n\t+======================== WARNING ========================+")
			fmt.Println("\t| The following scripts have not been migrated but appear |")
			fmt.Println("\t| before the current migration in the database.           |")
			fmt.Println("\t+======================== WARNING ========================+")
			for i, w := range warnings {
				fmt.Println(i, w.Hash, w.Path)
			}
			fmt.Println("\t========================= ======= =========================\n\n")
		}

		start, stop, err := migrations.FindMigrationRangeUp(currentMigration, steps)
		if err != nil {
			return err
		}

		err = migrations.Execute(start, stop)
		if err != nil {
			return err
		}
		return nil
	},
	Args: cobra.MaximumNArgs(1),
}

var downCmd = &cobra.Command{
	Use:   "down [steps]",
	Short: "Run one or more down migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		steps, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid number %s", args[0])
		}

		db, err := lib.NewDatabase("postgres://user:pass@localhost:5432/db?sslmode=disable")
		if err != nil {
			return err
		}

		migrations := lib.NewMigrations(migrationPath)
		err = migrations.Reconcile(db)
		if err != nil {
			return err
		}

		currentMigration, err := db.GetLatestMigration()
		if err != nil {
			return err
		}

		warnings := migrations.FindWarnings(currentMigration)
		if len(warnings) > 0 {
			fmt.Println("\n\n\t+======================== WARNING ========================+")
			fmt.Println("\t| The following scripts have not been migrated but appear |")
			fmt.Println("\t| before the current migration in the database.           |")
			fmt.Println("\t+======================== WARNING ========================+")
			for i, w := range warnings {
				fmt.Println(i, w.Hash, w.Path)
			}
			fmt.Println("\t========================= ======= =========================\n\n")
		}

		start, stop, err := migrations.FindMigrationRangeDown(currentMigration, steps)
		if err != nil {
			return err
		}
		err = migrations.Execute(start, stop)
		if err != nil {
			return err
		}
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

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("test")
		db, err := lib.NewDatabase("postgres://user:pass@localhost:5432/db?sslmode=disable")
		if err != nil {
			return err
		}
		migrations := lib.NewMigrations(migrationPath)
		err = migrations.Reconcile(db)
		if err != nil {
			return err
		}
		return nil
	},
}
