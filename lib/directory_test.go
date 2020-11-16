package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var migrationDirectory string

func init() {
	setup()
}

func setup() {

	tempDir, err := ioutil.TempDir(os.TempDir(), "goosey-*")
	if err != nil {
		log.Fatal(err)
	}

	migrationDirectory = tempDir

	if err = executeCommand("git", []string{"init"}, tempDir); err != nil {
		log.Fatal(err)
	}

	templates := []struct {
		name string
		sql  string
	}{
		{"up.sql", "CREATE TABLE %s (id SERIAL)"},
		{"down.sql", "DROP TABLE %s"},
	}

	for i, m := range []string{"a", "b", "c", "d", "e", "f", "g"} {
		path := filepath.Join(
			tempDir,
			fmt.Sprintf("2020010%d_120000_%s_%s_%s", i, m, m, m),
		)
		if err = os.Mkdir(path, 0777); err != nil {
			log.Fatal(err)
		}

		for _, template := range templates {
			err = makeSqlScript(
				path,
				template.name,
				fmt.Sprintf(template.sql, m),
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		if err = executeCommand(
			"git",
			[]string{"add", "."},
			tempDir,
		); err != nil {
			log.Fatal(err)
		}

		if err = executeCommand(
			"git",
			[]string{"commit", "-am", fmt.Sprintf("+%s", m)},
			tempDir,
		); err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Millisecond * 250)

	}
}

func executeCommand(command string, args []string, directory string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = directory
	return cmd.Run()
}

func makeSqlScript(path, name, script string) error {
	ufile, err := os.Create(filepath.Join(path, name))
	if err != nil {
		log.Fatal(err)
	}
	_, err = ufile.WriteString(script)
	return err
}

func Test_Initial(t *testing.T) {

	migrations := new(Migrations).List(migrationDirectory)

	assert.Equal(t, 7, len(migrations))
	order := []string{"a", "b", "c", "d", "e", "f", "g"}
	for i, migration := range migrations {
		parts := strings.Split(migration.Path, "_")
		message := parts[len(parts)-1]
		assert.Equal(t, order[i], message)
	}

}

var parseAutherFromPathTests = []struct {
	path     string
	expected string
	hasErr   bool
	err      error
}{
	{
		"date_time_john_zoidberg_message",
		"john zoidberg",
		false,
		nil,
	},
	{
		"date_time_john__message",
		"john",
		false,
		nil,
	},
	{
		"date_time__zoidberg_message",
		"zoidberg",
		false,
		nil,
	},
	{
		"date_zoidberg_message",
		"",
		true,
		fmt.Errorf("invalid directory structure"),
	},
}

func Test_parseAutherFromPath(t *testing.T) {
	for _, tt := range parseAutherFromPathTests {
		t.Run(tt.path, func(t *testing.T) {
			author, err := parseAuthorFromPath(tt.path)
			if tt.hasErr {
				assert.Equal(t, tt.expected, author)
				assert.Equal(t, tt.err, err)
			} else {
				assert.Equal(t, tt.expected, author)
				assert.NoError(t, err)
			}
		})
	}
}

var parseTimeFromPathTests = []struct {
	path     string
	expected time.Time
	hasErr   bool
	err      error
}{
	{
		"20200101_120000",
		time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
		false,
		nil,
	},
	{
		"foo",
		time.Time{},
		true,
		fmt.Errorf("invalid directory name foo"),
	},
}

func Test_parseTimeFromPath(t *testing.T) {
	for _, tt := range parseTimeFromPathTests {
		t.Run(tt.path, func(t *testing.T) {
			timestamp, err := parseTimeFromPath(tt.path)
			if tt.hasErr {
				assert.Equal(t, tt.expected, time.Time{})
				assert.Equal(t, tt.err, err)
			} else {
				assert.Equal(t, tt.expected, timestamp)
			}

		})
	}
}

var parseTimeFromCommitTests = []struct {
	input    string
	expected time.Time
	hasErr   bool
}{
	{
		"Sat, 14 Nov 2020 13:03:03 -0800",
		time.Date(2020, 11, 14, 13, 3, 3, 0, time.Local),
		false,
	},
	{
		"foo",
		time.Time{},
		true,
	},
}

func Test_parseTimeFromCommit(t *testing.T) {
	for _, tt := range parseTimeFromCommitTests {
		t.Run(tt.input, func(t *testing.T) {

			timestamp, err := parseTimeFromCommit(tt.input)

			if tt.hasErr {
				assert.Equal(t, tt.expected, time.Time{})
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expected, timestamp)
			}

		})
	}
}

var boundryTests = []struct {
	name     string
	items    int
	start    int
	steps    int
	expected int
}{
	{"step less than bound", 10, 5, 2, 7},
	{"step to bound", 10, 5, 5, 10},
	{"step over bound", 10, 5, 6, 10},
	{"no steps", 10, 5, -1, 10},
}

func Test_boundry(t *testing.T) {
	for _, tt := range boundryTests {

		t.Run(tt.name, func(t *testing.T) {
			actual := boundary(tt.items, tt.start, tt.steps)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

var sliceUpTests = []struct {
	hash     string
	steps    int
	expected Migrations
}{
	{
		"", 0,
		Migrations{
			&Migration{Index: 0, Hash: "a"},
			&Migration{Index: 0, Hash: "b"},
			&Migration{Index: 0, Hash: "c"},
			&Migration{Index: 0, Hash: "d"},
			&Migration{Index: 0, Hash: "e"},
			&Migration{Index: 0, Hash: "f"},
			&Migration{Index: 0, Hash: "g"},
		},
	},
	{
		"", 1,
		Migrations{
			&Migration{Index: 0, Hash: "a"},
		},
	},
	{
		"", 100,
		Migrations{
			&Migration{Index: 0, Hash: "a"},
			&Migration{Index: 0, Hash: "b"},
			&Migration{Index: 0, Hash: "c"},
			&Migration{Index: 0, Hash: "d"},
			&Migration{Index: 0, Hash: "e"},
			&Migration{Index: 0, Hash: "f"},
			&Migration{Index: 0, Hash: "g"},
		},
	},
	{
		"d", 2,
		Migrations{
			&Migration{Index: 0, Hash: "e"},
			&Migration{Index: 0, Hash: "f"},
		},
	},
	{
		"d", 0,
		Migrations{
			&Migration{Index: 0, Hash: "e"},
			&Migration{Index: 0, Hash: "f"},
			&Migration{Index: 0, Hash: "g"},
		},
	},
}

func Test_SliceUp(t *testing.T) {

	for _, tt := range sliceUpTests {
		t.Run(fmt.Sprintf("%s %d", tt.hash, tt.steps), func(t *testing.T) {
			migrations := Migrations{}
			for i, hash := range []string{"a", "b", "c", "d", "e", "f", "g"} {
				migrations = append(migrations, &Migration{
					Index: i,
					Hash:  hash,
				})
			}

			err := migrations.Slice(tt.hash, tt.steps, Up)
			assert.NoError(t, err)

			assert.Equal(t, len(tt.expected), len(migrations))

			for i, m := range migrations {
				assert.Equal(t, tt.expected[i].Hash, m.Hash)
			}
		})
	}
}

var sliceDownTests = []struct {
	hash     string
	steps    int
	expected Migrations
}{
	{
		"g", 100,
		Migrations{
			&Migration{Index: 0, Hash: "g"},
			&Migration{Index: 1, Hash: "f"},
			&Migration{Index: 2, Hash: "e"},
			&Migration{Index: 3, Hash: "d"},
			&Migration{Index: 4, Hash: "c"},
			&Migration{Index: 5, Hash: "b"},
			&Migration{Index: 6, Hash: "a"},
		},
	},
	{
		"d", 2,
		Migrations{
			&Migration{Index: 0, Hash: "d"},
			&Migration{Index: 1, Hash: "c"},
		},
	},
	{
		"d", 0,
		Migrations{
			&Migration{Index: 0, Hash: "d"},
			&Migration{Index: 1, Hash: "c"},
			&Migration{Index: 2, Hash: "b"},
			&Migration{Index: 3, Hash: "a"},
		},
	},
}

func Test_SliceDown(t *testing.T) {

	for _, tt := range sliceDownTests {
		t.Run(fmt.Sprintf("%s %d", tt.hash, tt.steps), func(t *testing.T) {
			migrations := Migrations{}
			for i, hash := range []string{"a", "b", "c", "d", "e", "f", "g"} {
				migrations = append(migrations, &Migration{
					Index: i,
					Hash:  hash,
				})
			}

			err := migrations.Slice(tt.hash, tt.steps, Down)
			assert.NoError(t, err)

			assert.Equal(t, len(tt.expected), len(migrations))

			for i, m := range migrations {
				assert.Equal(t, tt.expected[i].Hash, m.Hash)
			}
		})
	}
}
