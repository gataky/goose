package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/sir-wiggles/goose/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(testCmd)
}

var rootCmd = &cobra.Command{
	Use:   "goose",
	Short: "A SQL migration tool.",
	Long: `Goose is a SQL migration tool that applies migrations relative to the last migration 
file ran.  You can apply N number of migration up or down to your database and goose will keep 
track of everything for you.

Goose will apply migrations in the order that they were added to the repository.
`,
}

// Execute will run cobra cli
func Execute() error {
	return rootCmd.Execute()
}

func initConfig() {
	viper.AddConfigPath("$HOME/") // call multiple times to add many search paths
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.SetConfigName(".goose") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		migrations := lib.NewMigrations()
		fmt.Println(migrations[0])
		return nil
	},
}

func stepValidator(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return errors.New("Invalid number of arguments for this command, max 1")
	}
	if len(args) == 1 {
		_, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("Invalid number %s", args[0])
		}
	}
	return nil
}

func runMigration(args []string, direction lib.Direction) error {

	var steps int
	if len(args) > 0 {
		steps, _ = strconv.Atoi(args[0])
	}

	db, err := lib.NewDatabase()
	if err != nil {
		return err
	}

	migrations := lib.NewMigrations()

	currentMigration, err := db.GetHashForMarkerN(1)
	if err != nil {
		return err
	}

	if err = migrations.Range(currentMigration, steps, direction); err != nil {
		return err
	}

	if err := migrations.Execute(direction, db); err != nil {
		return err
	}
	return nil
}
