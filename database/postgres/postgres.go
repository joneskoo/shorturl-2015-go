package postgres

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/joneskoo/shorturl-go/database"
	"github.com/joneskoo/shorturl-go/models"
	_ "github.com/lib/pq" // postgresql driver
)

const idBase = 36

// SQL
const (
	sqlByID   = "SELECT url, host, ts FROM shorturl WHERE id = $1"
	sqlByURL  = "SELECT id, host, ts FROM shorturl WHERE url = $1"
	sqlInsert = "INSERT INTO shorturl(url, host, cookie) VALUES ($1, $2, $3) RETURNING id, ts"
)

type postgres struct {
	*sql.DB
}

// New creates a database configured from command line flags.
func New(conn string) (database.Database, error) {
	db, err := sql.Open("postgres", conn)
	return &postgres{db}, err
}

type shorturl struct {
	id        int64
	targetURL string
	host      string
	added     time.Time
}

func (s shorturl) UID() string {
	return strconv.FormatInt(s.id, idBase)
}

func (s shorturl) TargetURL() string {
	return s.targetURL
}

func (s shorturl) CreatedByIP() string {
	return s.host
}

func (s shorturl) Added() time.Time {
	return s.added
}

func (s shorturl) Shorturl() models.Shorturl {
	return models.Shorturl{
		UID:   s.UID(),
		URL:   s.TargetURL(),
		Added: s.Added(),
		Host:  s.CreatedByIP(),
	}
}

// Get retrieves short url from database by short id
func (db *postgres) Get(shortCode string) (*models.Shorturl, error) {
	var err error
	s := shorturl{}
	s.id, err = strconv.ParseInt(shortCode, idBase, 32)
	if err != nil {
		return nil, database.NotFound
	}
	err = db.QueryRow(sqlByID, s.id).Scan(&s.targetURL, &s.host, &s.added)
	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}
	return &models.Shorturl{
		UID:   s.UID(),
		URL:   s.TargetURL(),
		Added: s.Added(),
		Host:  s.CreatedByIP(),
	}, err
}

// List retrieves a list of short URLs from database.
func (db *postgres) List() (<-chan models.Shorturl, error) {
	// Query shorturl from database
	rows, err := db.Query("SELECT id, url, host, ts FROM shorturl")
	if err != nil {
		return nil, err
	}
	shorturls := make(chan models.Shorturl)
	go func(ch chan<- models.Shorturl) {
		defer close(ch)
		defer rows.Close()
		for rows.Next() {
			s := shorturl{}
			err := rows.Scan(&s.id, &s.targetURL, &s.host, &s.added)
			if err != nil {
				return
			}
			ch <- s.Shorturl()
		}
	}(shorturls)
	return shorturls, nil
}

// Add Short URL to database.
func (db *postgres) Add(url, host, clientid string) (models.Shorturl, error) {
	s, err := db.getByURL(url)
	if err == nil {
		return s.Shorturl(), nil
	}
	if err == database.NotFound {
		// Normal case: did not exist, so add it
		s = shorturl{targetURL: url, host: host}
		err = db.QueryRow(sqlInsert, url, host, clientid).Scan(&s.id, &s.added)
		return s.Shorturl(), err
	}

	return models.Shorturl{}, err
}

// getByURL retrieves short url from database by target URL
func (db *postgres) getByURL(url string) (shorturl, error) {
	s := shorturl{targetURL: url}
	err := db.QueryRow(sqlByURL, s.targetURL).Scan(&s.id, &s.host, &s.added)
	if err == sql.ErrNoRows {
		err = database.NotFound
	}
	return s, err
}
