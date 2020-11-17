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
	rootCmd.SetHelpCommand(aboutCmd)
}

var rootCmd = &cobra.Command{
	Use:          "goose",
	Short:        "A PostgreSQL migration tool.",
	SilenceUsage: true,
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

	batch, err := db.LastBatch()
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	if err = migrations.Slice(batch.Hash, steps, direction); err != nil {
		return err
	}

	if err := migrations.Execute(direction, db, batch.Exclude); err != nil {
		return err
	}
	return nil
}

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "The internal working of goose",
	Long: `Goose is a SQL migration tool that applies migrations relative to the last migration file ran.  You can apply N number of migration up or down to your database and goose will keep track of everything for you.

Terms:
* migration: a single script to run against the database
* batch    : a group of migrations to run against the database
* marker   : a marker of the migrations that were last in their batch
* hash     : the git hash of those added files

Goose keeps track of the migrations ran in a table called goosey in your database. To initialize this table you'll need to run "goose init" which will create a table that has six important fields:
* created_at : The time a migration was made. This information is encoded in the directory name
* merged_at  : The time a migration was committed with the repository
* executed_at: The time a migration was ran against the database
* hash       : The commit hash of when these files were added to the repository
* author     : The author of a migration. This information is encoded in the directory name
* marker     : An indicator if this migration is the last in its batch

To make a migration run "goose make (firstname) (lastname) (message)" this will create directory with the format yyyymmdd_hhmmss_firstname_lastname_message and in that directory two files, up.sql and down.sql, will be created where you'll write your SQL scripts. Other files added do these directories will be ignored by goose.

The structure of your migration directory will look like 
.
├── 20201023_010000_a_o_solv-20201023_010000_a
│   ├── down.sql
│   └── up.sql
├── 20201023_020000_b_o_solv-20201023_020000_b
│   ├── down.sql
│   └── up.sql
├── 20201023_030000_c_o_solv-20201023_030000_c
│   ├── down.sql
│   └── up.sql
├── 20201023_040000_d_o_solv-20201023_040000_d
│   ├── down.sql
│   └── up.sql
├── 20201023_050000_e_o_solv-20201023_050000_e
│   ├── down.sql
│   └── up.sql
├── 20201023_060000_f_o_solv-20201023_060000_f
│   ├── down.sql
│   └── up.sql
└── 20201023_070000_g_o_solv-20201023_070000_g
    ├── down.sql
    └── up.sql

Goose will apply migrations in the order that they were added to the repository. To determine this order goose runs "git log --pretty='format:%Cred%H|%aD' --name-status --diff-filter=A --reverse" which produces

e965f4511fce6ae61e1cfdcf174f61cfd4fe920b|Wed, 11 Nov 2020 22:53:49 -0800
A       20201023_010000_a_o_solv-20201023_010000_a/down.sql
A       20201023_010000_a_o_solv-20201023_010000_a/up.sql

cac4966fa648df678b9f59117d085b40d647ef19|Wed, 11 Nov 2020 22:54:06 -0800
A       20201023_020000_b_o_solv-20201023_020000_b/down.sql
A       20201023_020000_b_o_solv-20201023_020000_b/up.sql

e0ca0a9d0afe2d168ed09efe2f859f76bcfd109f|Wed, 11 Nov 2020 22:54:20 -0800
A       20201023_030000_c_o_solv-20201023_030000_c/down.sql
A       20201023_030000_c_o_solv-20201023_030000_c/up.sql

6a8f40ecd57b264da0d0492af62b577f626bfbe1|Wed, 11 Nov 2020 22:54:56 -0800
A       20201023_040000_d_o_solv-20201023_040000_d/down.sql
A       20201023_040000_d_o_solv-20201023_040000_d/up.sql

76499a490b9c0006100d963e6006f72cf56c6826|Wed, 11 Nov 2020 22:55:07 -0800
A       20201023_050000_e_o_solv-20201023_050000_e/down.sql
A       20201023_050000_e_o_solv-20201023_050000_e/up.sql

9ebb39681a4428cc5693ea2d926e5f73711ce9a4|Wed, 11 Nov 2020 22:55:17 -0800
A       20201023_060000_f_o_solv-20201023_060000_f/down.sql
A       20201023_060000_f_o_solv-20201023_060000_f/up.sql

cc7eff6ea9e68da4265bc834afda28f9a9db05a8|Wed, 11 Nov 2020 22:55:31 -0800
A       20201023_070000_g_o_solv-20201023_070000_g/down.sql
A       20201023_070000_g_o_solv-20201023_070000_g/up.sql

Goose will parse this output and apply the migration in a top down approach. 
`,
	SilenceUsage: true,
}
