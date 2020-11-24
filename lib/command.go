package lib

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:          "goose",
		Short:        "A PostgreSQL migration tool.",
		SilenceUsage: false,
	}
	db           *DB
	instructions *Instructions
	migrations   Migrations
	err          error
)

func init() {
	cobra.OnInitialize(initConfig)
	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(redoCmd)

	listCmd.AddCommand(listExecutedCmd)
	listCmd.AddCommand(listPendingCmd)

	makeCmd.Flags().StringVarP(&templateType, "template", "t", "schema", `The template to use to make your migration scripts. These templates are defined in the .goose.yaml file.`)
}

func initConfig() {
	viper.AddConfigPath("$HOME/")
	viper.AddConfigPath(".")
	viper.SetConfigName(".goose")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	initDependancies()
}

func initDependancies() {
	var err error
	db, err = NewDatabase()
	if err != nil {
		log.Fatal(err)
	}

	migrations = NewMigrations()
}

// Execute will run cobra cli
func Execute() error {
	return rootCmd.Execute()
}

func stepValidator(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return errors.New("invalid number of arguments for this command")
	}
	if len(args) == 1 {
		_, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("Invalid number %s", args[0])
		}
	}
	return nil
}

var initCmd = &cobra.Command{
	Use:   "init [commit hash]",
	Short: "Initializes a migration table in the database called goosey",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		start := ""
		if len(args) == 1 {
			start = args[0]
		}

		if err := db.InitGoosey(start); err != nil {
			return err
		}
		fmt.Println("initialized successfully")
		return nil
	},
}

var upCmd = &cobra.Command{
	Use:   "up [steps]",
	Short: "Run one or more up migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		instructions := NewInstructions(up, args...)
		err = db.LastBatch(instructions)
		if err != nil && err.Error() != "sql: no rows in result set" {
			log.Fatal(err)
		}

		if err := migrations.Slice(instructions); err != nil {
			if err.Error() == "no marker" {
				return nil
			}
			return err
		}

		if err := migrations.Execute(instructions); err != nil {
			return err
		}
		return nil
	},
	Args: stepValidator,
}

var downCmd = &cobra.Command{
	Use:   "down [steps]",
	Short: "Run one or more down migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		instructions := NewInstructions(down, args...)
		err = db.LastBatch(instructions)
		if err != nil && err.Error() != "sql: no rows in result set" {
			log.Fatal(err)
		}

		if err := migrations.Slice(instructions); err != nil {
			if err.Error() == "no marker" {
				return nil
			}
			return err
		}

		if err := migrations.Execute(instructions); err != nil {
			return err
		}
		return nil
	},
	Args: stepValidator,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List pending or executed migrations",
}

var listExecutedCmd = &cobra.Command{
	Use:   "executed",
	Short: "List all executed migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		instructions = NewInstructions(executed, []string{"10"}...)
		err = db.LastBatch(instructions)
		if err != nil && err.Error() != "sql: no rows in result set" {
			log.Fatal(err)
		}

		if err := migrations.Slice(instructions); err != nil {
			if err.Error() == "no marker" {
				return nil
			}
			return err
		}

		for _, d := range migrations {
			fmt.Println(d.Hash, d.Path)
		}
		return nil
	},
}

var listPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		instructions = NewInstructions(pending)
		err = db.LastBatch(instructions)
		if err != nil && err.Error() != "sql: no rows in result set" {
			log.Fatal(err)
		}

		if err := migrations.Slice(instructions); err != nil {
			if err.Error() == "no marker" {
				return nil
			}
			return err
		}

		for _, d := range migrations {
			fmt.Println(d.Hash, d.Path)
		}
		return nil
	},
}

var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "Rollback to the last marker and reapply to the current marker",
	RunE: func(cmd *cobra.Command, args []string) error {

		instructions := NewInstructions(redo)
		err = db.LastBatch(instructions)
		if err != nil && err.Error() != "sql: no rows in result set" {
			log.Fatal(err)
		}
		sort.Sort(sort.Reverse(migrations))

		if err := migrations.Slice(instructions); err != nil {
			if err.Error() == "no marker" {
				return nil
			}
			return err
		}

		if err := migrations.Execute(instructions); err != nil {
			return err
		}

		sort.Sort(migrations)
		instructions.Direction = Up

		if err := migrations.Execute(instructions); err != nil {
			return err
		}
		return nil
	},
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback to the last marker",
	RunE: func(cmd *cobra.Command, args []string) error {
		instructions = NewInstructions(rollback)
		err = db.LastBatch(instructions)
		if err != nil && err.Error() != "sql: no rows in result set" {
			log.Fatal(err)
		}
		sort.Sort(sort.Reverse(migrations))

		if err := migrations.Slice(instructions); err != nil {
			if err.Error() == "no marker" {
				return nil
			}
		}

		if err := migrations.Execute(instructions); err != nil {
			return err
		}
		return nil
	},
}

type Values struct {
	Migration string
	Author    string
	Directory string
	Timestamp string
}

var templateType string

var makeCmd = &cobra.Command{
	Use:   "make {first_name} {last_name} {message}",
	Short: "Make a new migration",
	RunE: func(cmd *cobra.Command, args []string) error {

		now := time.Now()
		timestamp := now.Format("20060102_150405")
		migration := fmt.Sprintf("%s_%s_%s_%s",
			timestamp, args[0], args[1], args[2])
		directory := filepath.Join(
			viper.GetString("migration-repository"),
			viper.GetString("migration-directory"),
			migration,
		)
		fmt.Println(directory)
		author := fmt.Sprintf("%s %s", args[0], args[1])

		templates := viper.GetStringMap(
			"templates")[templateType].(map[string]interface{})

		if err := os.Mkdir(directory, 0777); err != nil {
			return err
		}

		values := Values{
			Migration: migration,
			Author:    author,
			Directory: directory,
			Timestamp: timestamp,
		}

		if err := script(templates, values, "up"); err != nil {
			return err
		}
		if err := script(templates, values, "down"); err != nil {
			return err
		}

		return nil
	},
	Args: cobra.ExactArgs(3),
}

func script(tmpls map[string]interface{}, values Values, fname string) error {

	t, err := template.New("new").Parse(tmpls[fname].(string))
	if err != nil {
		return err
	}

	f, err := os.Create(
		filepath.Join(values.Directory, fmt.Sprintf("%s.sql", fname)))
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, values)
}
