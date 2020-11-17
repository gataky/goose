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
		SilenceUsage: true,
	}
	db         *DB
	batch      *BatchInfo
	migrations Migrations
)

func init() {
	cobra.OnInitialize(initConfig)
	cobra.EnableCommandSorting = false
	rootCmd.SetHelpCommand(aboutCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(redoCmd)

	listCmd.AddCommand(listExecutedCmd)
	listCmd.AddCommand(listPendingCmd)

	makeCmd.Flags().StringVarP(&templateType, "template", "t", "schema", `The 
	template to use to make your migration scripts. These templates are defined 
	in the .goose.yaml file.`)
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

	batch, err = db.LastBatch()
	if err != nil && err.Error() != "sql: no rows in result set" {
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

func runMigration(args []string, direction int) error {

	var steps int
	if len(args) > 0 {
		steps, _ = strconv.Atoi(args[0])
	}

	if err := migrations.Slice(batch.Hash, steps, direction); err != nil {
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
	Long: `Goose is a SQL migration tool that applies migrations relative to the 
	last migration file ran.  You can apply N number of migration up or down to 
	your database and goose will keep track of everything for you.

Terms:
* migration: a single script to run against the database
* batch    : a group of migrations to run against the database
* marker   : a marker of the migrations that were last in their batch
* hash     : the git hash of those added files

Goose keeps track of the migrations ran in a table called goosey in your 
database. To initialize this table you'll need to run "goose init" which will 
create a table that has six important fields:
* created_at : The time a migration was made. This information is encoded in the 
		       directory name
* merged_at  : The time a migration was committed with the repository
* executed_at: The time a migration was ran against the database
* hash       : The commit hash of when these files were added to the repository
* author     : The author of a migration. This information is encoded in the 
			   directory name
* marker     : An indicator if this migration is the last in its batch

To make a migration run "goose make (firstname) (lastname) (message)" this will 
create directory with the format yyyymmdd_hhmmss_firstname_lastname_message and 
in that directory two files, up.sql and down.sql, will be created where you'll 
write your SQL scripts. Other files added do these directories will be ignored 
by goose.

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

Goose will apply migrations in the order that they were added to the 
repository. To determine this order goose runs "git log 
--pretty='format:%Cred%H|%aD' --name-status --diff-filter=A --reverse" 
which produces

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
	Long: `Assuming we're starting with a new database and we want to apply the 
	first three migrations we could run "goose up 3" which will run the first 
	three migrations.  In the example case: a, b and c

+------------------------------------------+----------+----------+
| hash                                     | author   | marker   |
|------------------------------------------+----------+----------|
| e965f4511fce6ae61e1cfdcf174f61cfd4fe920b | a o      | False    |
| cac4966fa648df678b9f59117d085b40d647ef19 | b o      | False    |
| e0ca0a9d0afe2d168ed09efe2f859f76bcfd109f | c o      | True     |
+------------------------------------------+----------+----------+

This is referred to as a batch and the last migration in this batch is marked 
as true to indicate it's the last one.  More on this when we get to rollbacks 
and redos.

If you want to apply all the migration you can run "goose up" and that will run 
every migration that's remaining.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigration(args, Up)
	},
	Args: stepValidator,
}

var downCmd = &cobra.Command{
	Use:   "down [steps]",
	Short: "Run one or more down migrations",
	Long: `Assuming we're starting with a database that has migrations already 
	in it (a, b and c)

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

If you want to remove all the migration you can run "goose down" and that will 
undo every migration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigration(args, Down)
	},
	Args: stepValidator,
}

type templateValues struct {
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
			viper.GetString("migration-path"), migration)
		author := fmt.Sprintf("%s %s", args[0], args[1])

		templates := viper.GetStringMap(
			"templates")[templateType].(map[string]interface{})

		if err := os.Mkdir(directory, 0777); err != nil {
			return err
		}

		values := templateValues{
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

func script(
	templates map[string]interface{}, values templateValues, fname string,
) error {

	t, err := template.New("new").Parse(templates[fname].(string))
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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List pending or executed migrations",
}

func listMigrations(direction int) error {

	if err := migrations.Slice(batch.Batch, -1, direction); err != nil {
		return err
	}

	for _, d := range migrations {
		fmt.Println(d.Hash, d.Path)
	}
	return nil
}

var listExecutedCmd = &cobra.Command{
	Use:   "executed",
	Short: "List all executed migrations",
	Long: `To get a list of migrations that have already ran, run "goose 
	list executed"

e0ca0a9d0afe2d168ed09efe2f859f76bcfd109f 20201023_030000_c_o_solv-c
cac4966fa648df678b9f59117d085b40d647ef19 20201023_020000_b_o_solv-b
e965f4511fce6ae61e1cfdcf174f61cfd4fe920b 20201023_010000_a_o_solv-a

The output order will be from most recent migration ran to the oldest.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listMigrations(Down)
	},
}

var listPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List all pending migrations",
	Long: `To get a list of migrations that are ready to be ran run "goose 
	list pending"

6a8f40ecd57b264da0d0492af62b577f626bfbe1 20201023_040000_d_o_solv-d
76499a490b9c0006100d963e6006f72cf56c6826 20201023_050000_e_o_solv-e
9ebb39681a4428cc5693ea2d926e5f73711ce9a4 20201023_060000_f_o_solv-f
cc7eff6ea9e68da4265bc834afda28f9a9db05a8 20201023_070000_g_o_solv-g

The output order will be from the oldest migration that hasn't been ran to the 
newest.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listMigrations(Up)
	},
}

var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "Rollback to the last marker and reapply to the current marker",
	Long: `If you want to rollback and reapply that batch, "goose redo" will 
	do that for you.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		sort.Sort(sort.Reverse(migrations))

		if err := migrations.Slice(batch.Hash, batch.Steps, Down); err != nil {
			return err
		}

		if err := migrations.Execute(Down, db, batch.Exclude); err != nil {
			return err
		}

		sort.Sort(migrations)

		if err := migrations.Execute(Up, db, batch.Exclude); err != nil {
			return err
		}
		return nil
	},
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback to the last marker",
	Long: `Rollbacks will rollback a batch of migrations using the marker 
	talked about above. For exampe, here we have two batches:  
	* 1: a, b and c 
	* 2: d, e and f
Where f and c are markers indicating the last migration ran in their batch.

+------------------------------------------+----------+----------+
| hash                                     | author   | marker   |
|------------------------------------------+----------+----------|
| e965f4511fce6ae61e1cfdcf174f61cfd4fe920b | a o      | False    |
| cac4966fa648df678b9f59117d085b40d647ef19 | b o      | False    |
| e0ca0a9d0afe2d168ed09efe2f859f76bcfd109f | c o      | True     |
| 6a8f40ecd57b264da0d0492af62b577f626bfbe1 | d o      | False    |
| 76499a490b9c0006100d963e6006f72cf56c6826 | e o      | False    |
| 9ebb39681a4428cc5693ea2d926e5f73711ce9a4 | f o      | True     |
+------------------------------------------+----------+----------+

To rollback to c run "goose rollback" which will put us in this state

+------------------------------------------+----------+----------+
| hash                                     | author   | marker   |
|------------------------------------------+----------+----------|
| e965f4511fce6ae61e1cfdcf174f61cfd4fe920b | a o      | False    |
| cac4966fa648df678b9f59117d085b40d647ef19 | b o      | False    |
| e0ca0a9d0afe2d168ed09efe2f859f76bcfd109f | c o      | True     |
+------------------------------------------+----------+----------+ `,
	RunE: func(cmd *cobra.Command, args []string) error {

		sort.Sort(sort.Reverse(migrations))

		err := migrations.Slice(batch.Hash, batch.Steps, Down)
		if err != nil {
			return err
		}

		if err := migrations.Execute(Down, db, batch.Exclude); err != nil {
			return err
		}
		return nil
	},
}
