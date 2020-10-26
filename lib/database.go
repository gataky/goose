package lib

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Datastore interface {
	DeleteLastMigration(string) error
	GetHashForMarkerN(offset int) error
	GetLatestMigration() (string, error)
	GetListOfMigrations() ([]Migration, error)
	RunScript(path string) error
	SetLatestMigration(string, string, bool) error
}

type DB struct {
	*sql.DB
}

const (
	pg_UNDEFINED_TABLE = "42P01"
	UNDEFINED_TABLE    = "table does not exist"
)

func NewDatabase(url string) (*DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return &DB{db}, db.Ping()
}

func (db DB) GetHashForMarkerN(offset int) (string, error) {
	var hash string

	err := db.QueryRow(`
		SELECT hash FROM goosey 
			ORDER BY created_at DESC
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

func (db DB) GetListOfMigrations() (Migrations, error) {
	rows, err := db.Query("SELECT created_at, name, hash, marker FROM goosey")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	migrations := make(Migrations, 0, 10)
	for rows.Next() {
		var (
			name       string
			hash       string
			marker     bool
			executedAt time.Time
		)
		err := rows.Scan(&executedAt, &name, &hash, &marker)
		if err != nil {
			log.Fatal(err)
		}
		migrations = append(migrations, &Migration{
			Date:   executedAt,
			Path:   name,
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

func (db DB) SetLatestMigration(script Script, marker bool) error {
	_, err := db.Exec(`
		INSERT INTO goosey (
			name, hash, author, marker
		) VALUES ($1, $2, $3, $4)
	`, script.Name, script.Hash, script.Author, marker)
	return err
}

func (db DB) DeleteLastMigration(hash string) error {
	_, err := db.Exec("DELETE FROM goosey WHERE hash = $1", hash)
	return err
}

func (db DB) RunScript(script string) error {
	_, err := db.Exec(script)
	return err
}

func (db DB) CreateMigrationTable() error {
	_, err := db.Exec(`
		CREATE TABLE goosey (
			id SERIAL  PRIMARY KEY,
			created_at TIMESTAMP DEFAULT NOW(),
			name 	   TEXT,
			hash       TEXT,
			author     TEXT,
			marker     BOOLEAN
		);
	`)
	return err
}
