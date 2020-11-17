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

type BatchInfo struct {
	// The hash where the marker is true for the batch
	Batch string

	// Number of steps to the next batch
	Steps int

	// Hash to exclude. Used in cases where the database was initialized
	// with a hash.
	Exclude string

	// The starting point for the batch
	Hash string
}

/*
 * LastBatch will return the last marker and the number of steps to the
 * marker before that.
 */
func (db DB) LastBatch() (*BatchInfo, error) {
	var batch string
	var hash string
	var steps int
	var rowid int
	rows, err := db.Query(`
		( SELECT 
			a.batch, a.hash, b.steps, a.id
		FROM goosey AS a JOIN ( 
			SELECT 
				MAX (id) id, batch, COUNT(batch) steps
			FROM goosey GROUP BY batch
		) AS b 
			ON a.id = b.id
			ORDER BY a.executed_at DESC LIMIT 1 )
		UNION ALL
		( SELECT 
			a.batch, a.hash, b.steps, a.id
		FROM goosey AS a JOIN ( 
			SELECT 
				MAX (id) id, batch, COUNT(batch) steps
			FROM goosey GROUP BY batch
		) AS b 
			ON a.id = b.id
			ORDER BY a.executed_at ASC LIMIT 1 )
		ORDER BY id DESC;
	`)
	if err != nil {
		return &BatchInfo{}, err
	}
	batches := make([]*BatchInfo, 0, 2)
	for i := 0; rows.Next(); i++ {
		if err = rows.Scan(&batch, &hash, &steps, &rowid); err != nil {
			return nil, err
		}
		batches = append(batches, &BatchInfo{
			Batch: batch,
			Hash:  hash,
			Steps: steps,
		})
	}
	if len(batches) > 0 && batches[1].Batch == "" {
		batches[0].Exclude = batches[1].Hash
		return batches[0], err
	} else if len(batches) > 0 {
		return batches[0], err
	}
	return &BatchInfo{}, err
}

/*
 * InsertLastMigration inserts a row into goosey with information related to
 * the migration afte
 */
func (db DB) InsertLastMigration(script Script) error {
	_, err := db.Exec(`
		INSERT INTO goosey (
			created_at, merged_at, hash, author, batch
		) VALUES ($1, $2, $3, $4, $5)
	`, script.CreateDate, script.MergedDate, script.Hash, script.Author, script.Batch)
	return err
}

/*
 * DeleteLastMigration deletes a row from goosey.  If marker is true then the
 * last row in the table will have its marker column set to true.
 */
func (db DB) DeleteLastMigration(hash string) error {
	if _, err := db.Exec(`
			DELETE FROM goosey WHERE hash = $1
		`, hash); err != nil {
		return err
	}
	return nil
}

/*
 * RunScript executes a string of sql
 */
func (db DB) RunScript(script string) error {
	_, err := db.Exec(script)
	return err
}

/*
 * InitGoosey initializes the goosey table.
 */
func (db DB) InitGoosey(start string) error {

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

	_, err = tx.Exec(`
		CREATE TABLE goosey (
			id          SERIAL PRIMARY KEY,
			created_at  TIMESTAMPTZ,
			merged_at   TIMESTAMPTZ,
			executed_at TIMESTAMPTZ DEFAULT NOW(),
			hash        TEXT,
			author      TEXT,
			batch       TEXT
		);
	`)
	if err != nil {
		return err
	}

	if len(start) > 0 {
		_, err = tx.Exec(`
			INSERT INTO goosey 
				(hash, batch) 
			VALUES ($1, $2)
		`, start, "")
	}
	return err
}
