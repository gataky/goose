package lib

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

const (
	upScriptName   = "up.sql"
	downScriptName = "down.sql"
	Down           = 0
	Up             = 1
)

var (
	timeRegex   *regexp.Regexp
	timePattern = "20060102_030405"
)

func init() {
	var err error
	timeRegex, err = regexp.Compile(`(\d{8}_\d{6})`)
	if err != nil {
		log.Fatal(err)
	}
}

type Migrations []*Migration

func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m Migrations) Less(i, j int) bool { return m[i].Date.Before(m[j].Date) }

// NewMigrations creates a list of Migrations
func NewMigrations(path string) Migrations {
	return new(Migrations).List(path)
}

// List returns a sorted list of Migrations in assending order based on time.
func (ds Migrations) List(path string) Migrations {
	mapping := make(map[string]*Migration, 10)

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			_, name := filepath.Split(path)
			hash := fmt.Sprintf("%x", md5.Sum([]byte(name)))

			date, err := getTimeFromPath(path)
			if err != nil {
				fmt.Printf("skipping: %s\n", err)
				return nil
			}
			mapping[name] = &Migration{Path: path, Hash: hash, Date: date}
		} else {
			name := filepath.Dir(path)
			_, name = filepath.Split(name)
			hash := fmt.Sprintf("%x", md5.Sum([]byte(name)))

			script := Script{Path: path, Hash: hash}

			migration := mapping[name]
			if info.Name() == upScriptName {
				migration.Up = script
			} else if info.Name() == downScriptName {
				migration.Down = script
			} else {
				log.Printf("skipping: %s\n", path)
			}
		}
		return nil
	})

	migrations := make(Migrations, 0, len(mapping))
	for _, v := range mapping {
		migrations = append(migrations, v)
	}
	sort.Sort(migrations)
	return migrations
}

func getTimeFromPath(path string) (time.Time, error) {
	match := timeRegex.FindAllString(path, 1)
	if len(match) == 0 {
		return time.Time{}, fmt.Errorf("invalid directory name %s", path)
	}
	return time.Parse(timePattern, match[0])
}

func (ds *Migrations) Reconcile(db *DB) error {
	ranMigrations, err := db.GetListOfMigrations()
	if err != nil {
		return err
	}

	ranMigrationMap := make(map[string]*Migration)
	for _, migration := range ranMigrations {
		ranMigrationMap[migration.Hash] = migration
	}

	for _, migration := range *ds {
		if _, ok := ranMigrationMap[migration.Hash]; ok {
			(*migration).Migrated = true
		}
	}

	return nil
}

// FindMigrationRangeDown finds the range of migration scripts that are requested to run for the
// up process.
func (ds *Migrations) FindMigrationRangeUp(hash string, steps int) (int, int, error) {
	if hash == "" {
		return 0, boundry(len(*ds), 0, steps), nil
	}

	for index, migration := range *ds {
		if migration.Hash == hash {
			return index + 1, boundry(len(*ds), index+1, steps), nil
		}
	}

	return -1, -1, fmt.Errorf("can't find index for %s", hash)
}

func boundry(items, start, steps int) int {
	if start+steps >= items || steps == -1 {
		return items
	}
	return start + steps
}

// FindMigrationRangeDown finds the range of migration scripts that are requested to run for the
// down process. The list will also be in reverse order after calling this method.
func (ds *Migrations) FindMigrationRangeDown(hash string, steps int) (int, int, error) {
	if hash == "" {
		return -1, -1, fmt.Errorf("no starting point found")
	}
	sort.Sort(sort.Reverse(*ds))

	for index, migration := range *ds {
		if migration.Hash == hash {
			return index, boundry(len(*ds), index, steps), nil
		}
	}

	return -1, -1, fmt.Errorf("can't find index for %s", hash)
}

// Execute will execute the scripts in the range of start and stop
func (ds Migrations) Execute(start, stop int) error {
	for _, migration := range ds[start:stop] {
		fmt.Println(migration.Hash, migration.Path, migration.Migrated)
	}
	return nil
}

func (ds Migrations) FindWarnings(currentMigration string) Migrations {
	warnings := make(Migrations, 0, 10)
	currentMigrationFound := false
	for _, migration := range ds {
		if migration.Hash == currentMigration {
			currentMigrationFound = true
		}
		if migration.Migrated == false && currentMigrationFound == false {
			warnings = append(warnings, migration)
		}
	}
	return warnings
}

// Migration has all the relative information needed to run up and down scripts.
type Migration struct {
	Path     string
	Hash     string
	Date     time.Time
	Up       Script
	Down     Script
	Warning  bool
	Marker   bool
	Migrated bool
}

// Script is the specific up or down script.
type Script struct {
	Path string
	Hash string
}
