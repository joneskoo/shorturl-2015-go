package shorturl

import (
	"database/sql"
	"strconv"

	_ "github.com/lib/pq" // postgresql driver
)

type DB struct {
	*sql.DB
}

// SQL
const (
	sqlByID  = "SELECT url, host, ts FROM shorturl WHERE id = $1"
	sqlByURL = "SELECT id, host, ts FROM shorturl WHERE url = $1"
)

// Open creates a database configured from command line flags.
func Open(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// Get retrieves short url from database by short id
func (db *DB) Get(shortCode string) (*Shorturl, error) {
	var err error
	id, err := strconv.ParseInt(shortCode, idBase, 32)
	if err != nil {
		return nil, ErrNotFound
	}
	s := &Shorturl{ID: id}
	err = db.QueryRow(sqlByID, s.ID).Scan(&s.URL, &s.Host, &s.Added)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return s, err
}
