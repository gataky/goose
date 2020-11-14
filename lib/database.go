package lib

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

type DB struct {
	*sql.DB
}

func NewDatabase() (*DB, error) {
	url := viper.GetString("database-url")
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return &DB{db}, db.Ping()
}

// LastMarker will return the lats marker and the number of steps to the marker before that.
func (db DB) LastMarker() (string, int, error) {
	rows, err := db.Query(`
	SELECT hash FROM goosey WHERE 
		id <= ( SELECT id FROM goosey 
				WHERE marker = true 
				ORDER BY executed_at DESC 
				LIMIT 1 OFFSET 0 )
	AND
		id > COALESCE((SELECT id FROM goosey 
					   WHERE marker = true 
					   ORDER BY executed_at DESC 
					   LIMIT 1 OFFSET 1 ), 0)
	ORDER BY executed_at DESC
	`)

	if err != nil {
		return "", 0, err
	}
	defer rows.Close()

	hashes := make([]string, 0, 1)
	for rows.Next() {
		var hash string
		if err = rows.Scan(&hash); err != nil {
			return "", 0, err
		}
		hashes = append(hashes, hash)
	}
	var hash string
	if len(hashes) > 0 {
		hash = hashes[0]
	}
	return hash, len(hashes), nil
}

// InsertLastMigration inserts a row into goosey with information related to the migration afte
func (db DB) InsertLastMigration(script Script, marker bool) error {
	_, err := db.Exec(`
		INSERT INTO goosey (
			created_at, merged_at, hash, author, marker
		) VALUES ($1, $2, $3, $4, $5)
	`, script.CreateDate, script.MergedDate, script.Hash, script.Author, marker)
	return err
}

// DeleteLastMigration deletes a row from goosey.  If marker is true then the last row in the
// table will have its marker column set to true.
func (db DB) DeleteLastMigration(hash string, marker bool) error {
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

// RunScript executes a string of sql
func (db DB) RunScript(script string) error {
	_, err := db.Exec(script)
	return err
}

// InitGoosey initializes the goosey table.
func (db DB) InitGoosey() error {
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
