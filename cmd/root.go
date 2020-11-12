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
	Short: "A SQL migration tool for Solv",
	Long: `Goose is a SQL migration tool built for Solv.  Goose allows you
to run up and down migrations on a database.`,
}

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
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("test")

		migrations := lib.NewMigrations()
		fmt.Println(migrations[0])
		return nil
	},
}

func StepValidator(cmd *cobra.Command, args []string) error {
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
