package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(makeCmd)
	makeCmd.Flags().StringVarP(&templateType, "template", "t", "schema", "The template to use to make your migration scripts. These templates are defined in the .goose.yaml file.")
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
		migration := fmt.Sprintf("%s_%s_%s_%s", timestamp, args[0], args[1], args[2])
		directory := filepath.Join(viper.GetString("migration-path"), migration)
		author := fmt.Sprintf("%s %s", args[0], args[1])

		templates := viper.GetStringMap("templates")[templateType].(map[string]interface{})

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

func script(templates map[string]interface{}, values templateValues, fname string) error {

	t, err := template.New("new").Parse(templates[fname].(string))
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(values.Directory, fmt.Sprintf("%s.sql", fname)))
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, values)
}
