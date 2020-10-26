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
	Short: "Initializes a migration table in the database called goosey",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := lib.NewDatabase("postgres://user:pass@localhost:5432/db?sslmode=disable")
		if err != nil {
			return err
		}
		err = db.CreateMigrationTable()
		if err == nil {
			fmt.Println("initialized successfully")
		}
		return err
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

		var (
			steps int
			err   error
		)
		if len(args) == 1 {
			steps, err = strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid number %s", args[0])
			}
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

		if steps == 0 {
			steps = len(migrations)
		}

		currentMigration, err := db.GetHashForMarkerN(1)
		if err != nil {
			return err
		}

		start, stop, err := migrations.FindMigrationRangeUp(currentMigration, steps)
		if err != nil {
			return err
		}

		err = migrations.Execute(start, stop, lib.Up, db)
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
		var (
			steps int
			err   error
		)
		if len(args) == 1 {
			steps, err = strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid number %s", args[0])
			}
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

		if steps == 0 {
			steps = len(migrations)
		}

		currentMigration, err := db.GetHashForMarkerN(1)
		if err != nil {
			return err
		}

		start, stop, err := migrations.FindMigrationRangeDown(currentMigration, steps)
		if err != nil {
			return err
		}
		err = migrations.Execute(start, stop, lib.Down, db)
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
