package lib

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/spf13/viper"
)

type DB struct {
	*sql.DB
}

const (
	pg_UNDEFINED_TABLE = "42P01"
	UNDEFINED_TABLE    = "table does not exist"
)

func NewDatabase() (*DB, error) {
	url := viper.GetString("database-url")
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return &DB{db}, db.Ping()
}

func (db DB) GetLastNMarkers(limit int) (map[string]bool, error) {
	rows, err := db.Query(`
		SELECT hash FROM goosey 
			WHERE marker = TRUE
			ORDER BY executed_at DESC
			LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hashes := make(map[string]bool, limit)
	for rows.Next() {
		var hash string
		if err = rows.Scan(&hash); err != nil {
			return nil, err
		}
		hashes[hash] = true
	}
	if len(hashes) == 0 {
		return hashes, fmt.Errorf("no markers found")
	}
	return hashes, rows.Err()
}

func (db DB) GetHashForMarkerN(offset int) (string, error) {
	var hash string
	err := db.QueryRow(`
		SELECT hash FROM goosey 
			WHERE marker = TRUE
			ORDER BY executed_at DESC
			OFFSET $1
			LIMIT 1
	`, offset-1).Scan(&hash)
	if err != nil && err.Error() == "sql: no rows in result set" {
		return "", nil
	} else if err == nil {
		return hash, nil
	}

	pgerr, ok := err.(*pq.Error)
	if !ok {
		return "", err
	} else if pgerr.Code == pg_UNDEFINED_TABLE {
		return "", fmt.Errorf(UNDEFINED_TABLE)
	}
	return hash, err
}

func (db DB) ListMigrations() (Migrations, error) {
	rows, err := db.Query("SELECT created_at, hash, marker FROM goosey")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	migrations := make(Migrations, 0, 10)
	for rows.Next() {
		var (
			hash       string
			marker     bool
			executedAt time.Time
		)
		err := rows.Scan(&executedAt, &hash, &marker)
		if err != nil {
			log.Fatal(err)
		}
		migrations = append(migrations, &Migration{
			Hash:   hash,
			Marker: marker,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return migrations, err
}

func (db DB) SetLastMigration(script Script, marker bool) error {
	_, err := db.Exec(`
		INSERT INTO goosey (
			hash, author, marker, merged_at, created_at
		) VALUES ($1, $2, $3, $4, $5)
	`, script.Hash, script.Author, marker, script.MergedDate, script.CreateDate)
	return err
}

func (db DB) DelLastMigration(hash string, marker bool) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if _, err = tx.Exec("DELETE FROM goosey WHERE hash = $1", hash); err != nil {
		return err
	}
	if marker == true {
		if _, err = tx.Exec(`
			UPDATE goosey SET marker = TRUE WHERE id = (
				SELECT id FROM goosey ORDER BY executed_at DESC LIMIT 1
			)
		`); err != nil {
			return err
		}
	}
	return err
}

func (db DB) RunScript(script string) error {
	_, err := db.Exec(script)
	return err
}

func (db DB) CreateMigrationTable() error {
	_, err := db.Exec(`
		CREATE TABLE goosey (
			id SERIAL   PRIMARY KEY,
			created_at  TIMESTAMPTZ,
			merged_at   TIMESTAMPTZ,
			executed_at TIMESTAMPTZ DEFAULT NOW(),
			hash        TEXT,
			author      TEXT,
			marker      BOOLEAN
		);
	`)
	return err
}
