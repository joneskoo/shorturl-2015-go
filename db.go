package shorturl

import (
	"database/sql"
	"strconv"

	_ "github.com/lib/pq" // postgresql driver
)

// ConnString is the default database connection string
var ConnString = "user=joneskoo dbname=joneskoo sslmode=disable"

// SQL
const (
	sqlByID  = "SELECT url, host, ts FROM shorturl WHERE id = $1"
	sqlByURL = "SELECT id, host, ts FROM shorturl WHERE url = $1"
)

type Database struct {
	*sql.DB
}

// New creates a database configured from command line flags.
func New() (*Database, error) {
	db, err := sql.Open("postgres", ConnString)
	if err != nil {
		return nil, err
	}
	return &Database{db}, nil
}

// Get retrieves short url from database by short id
func (db *Database) Get(shortCode string) (*Shorturl, error) {
	var err error
	s := new(Shorturl)
	s.ID, err = strconv.ParseInt(shortCode, idBase, 32)
	if err != nil {
		return nil, ErrNotFound
	}
	err = db.QueryRow(sqlByID, s.ID).Scan(&s.URL, &s.Host, &s.Added)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return s, err
}

// List retrieves a list of short URLs from database.
func (db *Database) List() (<-chan Shorturl, error) {
	// Query shorturl from database
	rows, err := db.Query("SELECT id, url, host, ts FROM shorturl")
	if err != nil {
		return nil, err
	}
	shorturls := make(chan Shorturl)
	go func(ch chan<- Shorturl) {
		defer close(ch)
		defer rows.Close()
		for rows.Next() {
			s := Shorturl{}
			err := rows.Scan(&s.ID, &s.URL, &s.Host, &s.Added)
			if err != nil {
				return
			}
			ch <- s
		}
	}(shorturls)
	return shorturls, nil
}

// getByURL retrieves short url from database by target URL
func (db *Database) getByURL(url string) (s Shorturl, err error) {
	s = Shorturl{URL: url}
	err = db.QueryRow(sqlByURL, s.URL).Scan(&s.ID, &s.Host, &s.Added)
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}
	return
}
