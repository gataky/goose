package lib

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Direction int

const (
	DOWN Direction = iota
	UP
)

var timeRegex *regexp.Regexp

func init() {
	if exp, err := regexp.Compile(`(\d{8}_\d{6})`); err != nil {
		log.Fatal(err)
	} else {
		timeRegex = exp
	}
}

// Migration has all the relative information needed to run up and down scripts.
type Migration struct {
	Index int

	// Path is the system path of where the migration scripts are located
	Path string

	// Hash is the git hash id
	Hash string

	// MergedDate is the date the script was merged
	MergedDate time.Time

	Up   Script
	Down Script

	// Marker indicates if the migration is a stopping point in a batch migration
	Marker   bool
	Migrated bool
}

type Migrations []*Migration

func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m Migrations) Less(i, j int) bool { return m[i].Index < m[j].Index }

// NewMigrations creates a list of Migrations
func NewMigrations() Migrations {
	path := viper.GetString("migration-path")
	migrations := new(Migrations).List(path)
	return migrations
}

// List returns a sorted list of Migrations in assending order based on time.
func (migrations Migrations) List(path string) Migrations {

	cmd := exec.Command("git", "log", "--pretty=format:%H|%aD", "--name-status", "--diff-filter=A", "--reverse")
	cmd.Dir = path

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)

	index := 0
	for scanner.Scan() {
		hash_date := strings.Split(scanner.Text(), "|")
		hash := hash_date[0]
		date := hash_date[1]
		merged_timestamp, _ := parseTimeFromCommit(date)

		scanner.Scan()
		next := strings.Split(scanner.Text(), "\t")[1]
		dir := filepath.Dir(next)

		created_timestamp, _ := parseTimeFromPath(dir)
		author, _ := parseAuthorFromPath(dir)

		migrations = append(migrations, &Migration{
			Index:      index,
			Path:       dir,
			Hash:       hash,
			MergedDate: merged_timestamp,
			Up: Script{
				Hash:       hash,
				Path:       filepath.Join(path, dir, "up.sql"),
				MergedDate: merged_timestamp,
				CreateDate: created_timestamp,
				Author:     author,
				direction:  UP,
			},
			Down: Script{
				Hash:       hash,
				Path:       filepath.Join(path, dir, "down.sql"),
				MergedDate: merged_timestamp,
				CreateDate: created_timestamp,
				Author:     author,
				direction:  DOWN,
			},
		})
		index += 1

		// move to the next block
		for scanner.Scan() {
			next := scanner.Text()
			if next == "" {
				break
			}
		}
	}
	return migrations
}

func parseAuthorFromPath(path string) (string, error) {
	parts := strings.Split(path, "_")
	if len(parts) <= 4 {
		return "", fmt.Errorf("invalid directory structure")
	}
	return fmt.Sprintf("%s %s", parts[2], parts[3]), nil
}

func parseTimeFromPath(path string) (time.Time, error) {
	match := timeRegex.FindAllString(path, 1)
	if len(match) == 0 {
		return time.Time{}, fmt.Errorf("invalid directory name %s", path)
	}
	return time.Parse("20060102_030405", match[0])
}

func parseTimeFromCommit(timestamp string) (time.Time, error) {
	return time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", timestamp)
}

func (migrations *Migrations) Slice(markers map[string]bool) error {
	indices := make([]int, 0, 2)
	for i, m := range *migrations {
		if _, found := markers[m.Hash]; found {
			indices = append(indices, i)
		}
	}

	if len(markers) == 1 {
		indices = append(indices, len(*migrations))
	}
	*migrations = (*migrations)[indices[0]:indices[1]]
	return nil
}

func (migrations *Migrations) Range(hash string, steps int, direction Direction) error {
	if steps == 0 {
		steps = len(*migrations)
	}

	if hash == "" && direction == UP {
		start := 0
		stop := boundry(len(*migrations), 0, steps)
		*migrations = (*migrations)[start:stop]
		return nil
	} else if hash == "" && direction == DOWN {
		return fmt.Errorf("no starting point found, nothing to do")
	}

	if direction == DOWN {
		sort.Sort(sort.Reverse(*migrations))
	}

	for index, migration := range *migrations {
		if migration.Hash == hash {
			start := index + int(direction)
			stop := boundry(len(*migrations), index+int(direction), steps)
			*migrations = (*migrations)[start:stop]
			return nil
		}
	}

	return fmt.Errorf("can not find index for %s", hash)
}

func boundry(items, start, steps int) int {
	if start+steps >= items || steps == -1 {
		return items
	}
	return start + steps
}

// Execute will execute the scripts in the range of start and stop
func (migrations Migrations) Execute(direction Direction, db *DB) error {
	var err error
	last := len(migrations) - 1
	for i, migration := range migrations {
		if direction == UP {
			err = migration.Up.Execute(db, i == last)
		} else {
			err = migration.Down.Execute(db, i == last)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Script is the specific up or down script.
type Script struct {
	// Hash is the git commit hash for this migration
	Hash string

	// Path is the absolute path of the migration script
	Path string

	// MergedDate is the date the migration was committed to the repo
	MergedDate time.Time

	// CreateDate is the date the migration was created with the make command.
	// This is the date the is part of the directory where the scripts reside
	CreateDate time.Time

	Author    string
	direction Direction
}

func (s Script) Execute(db *DB, isLastMigration bool) error {
	script, err := ioutil.ReadFile(s.Path)
	if err != nil {
		return err
	}

	err = db.RunScript(string(script))
	if err != nil {
		return fmt.Errorf("failed to execute script %s %s: %s", s.Hash, s.Path, err)
	}

	if s.direction == UP {
		err = db.SetLastMigration(s, isLastMigration)
	} else {
		err = db.DelLastMigration(s.Hash, isLastMigration)
	}
	return err
}
