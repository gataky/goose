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
	GetLatestMigration() (string, error)
	GetListOfMigrations() ([]Migration, error)
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

	err = db.Ping()
	conn := &DB{db}

	_, err = conn.GetLatestMigration()
	if err != nil {
		fmt.Println(err)
	}
	return conn, nil
}

func (db DB) GetLatestMigration() (string, error) {
	var hash string
	err := db.QueryRow("SELECT hash FROM goosey ORDER BY created_at DESC LIMIT 1").Scan(&hash)
	if pgerr, ok := err.(*pq.Error); ok {
		if pgerr.Code == pg_UNDEFINED_TABLE {
			return "", fmt.Errorf(UNDEFINED_TABLE)
		}
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

func (db DB) SetLatestMigration(hash, name string, marker bool) error {
	return nil
}
