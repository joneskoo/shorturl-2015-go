package database

import (
	"database/sql"
	"strconv"

	_ "github.com/lib/pq" // postgresql driver
)

// ConnString is the default database connection string
var ConnString = "user=joneskoo dbname=joneskoo sslmode=disable"

// SQL
const (
	sqlByID   = "SELECT url, host, ts FROM shorturl WHERE id = $1"
	sqlByURL  = "SELECT id, host, ts FROM shorturl WHERE url = $1"
	sqlInsert = "INSERT INTO shorturl(url, host, cookie) VALUES ($1, $2, $3) RETURNING id, ts"
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

// GetByUID retrieves short url from database by base-36 id
func (db *Database) GetByUID(uid string) (*Shorturl, error) {
	var err error
	s := new(Shorturl)
	s.ID, err = strconv.ParseInt(uid, idBase, 64)
	if err != nil {
		return nil, ErrNotFound
	}
	err = db.QueryRow(sqlByID, s.ID).Scan(&s.URL, &s.Host, &s.Added)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return s, err
}

// GetByURL retrieves short url from database by base-36 id
func (db *Database) GetByURL(url string) (s Shorturl, err error) {
	s = Shorturl{URL: url}
	err = db.QueryRow(sqlByURL, s.URL).Scan(&s.ID, &s.Host, &s.Added)
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}
	return
}

// List retrieves short URLs from database
func (db *Database) List() (shorturls chan Shorturl, err error) {
	shorturls = make(chan Shorturl)
	// Query shorturl from database
	rows, err := db.Query("SELECT id, url, host, ts FROM shorturl")
	if err != nil {
		return
	}
	go func() {
		defer close(shorturls)
		defer rows.Close()
		for rows.Next() {
			s := Shorturl{}
			err := rows.Scan(&s.ID, &s.URL, &s.Host, &s.Added)
			if err != nil {
				return
			}
			shorturls <- s
		}
	}()
	return
}

// Add Short URL to database and return Shorturl object
func (db *Database) Add(url, host, clientid string) (s Shorturl, err error) {
	s, err = db.GetByURL(url)
	switch err {
	case ErrNotFound:
		// Normal case: did not exist, so add it
		s = Shorturl{URL: url, Host: host}
		err = db.QueryRow(sqlInsert, url, host, clientid).Scan(&s.ID, &s.Added)
		return s, err
	case nil:
		// No error, exists, re-use old one
		return s, nil
	default:
		// Other error, return error
		return Shorturl{}, err
	}
}
