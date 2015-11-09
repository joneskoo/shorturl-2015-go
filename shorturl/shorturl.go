package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

const usage = "Usage:\n  shorturl <uid>"

// Service configuration
const Domain = "yx.fi"
const IdBase = 36

// Database connection
const connString = "dbname=shorturl sslmode=disable"

// Errors
var ShorturlNotFound = errors.New("Shorturl not found")

// SQL
const getShorturlSql = "SELECT url, host, ts FROM shorturl WHERE id = $1;"

// Short URL database structure
type Shorturl struct {
	Id int64
	URL, Host string
	Added time.Time
}

// Save as new in database
func (s *Shorturl) Save(db *sql.DB) (err error) {
	var userid int
	err = db.QueryRow(`INSERT INTO users(name, favorite_fruit, age)
	VALUES('beatrice', 'starfruit', 93) RETURNING id`).Scan(&userid)
	if err != nil {
		panic(err)
	}

	log.Print("Added shorturl ", s)
	return
}

// Get Short URL from database by base-36 id
func getShorturl(db *sql.DB, uid string) (s Shorturl, err error) {
	s = Shorturl{}

	// Parse uid to id
	s.Id, err = strconv.ParseInt(uid, IdBase, 64)
	if err != nil {
		return
	}

	// Query shorturl from database
	err = db.QueryRow(getShorturlSql, s.Id).Scan(&s.URL, &s.Host, &s.Added)
	return
}

// Convert Short URL id to base-36 string format
func (s Shorturl) Uid() string {
	return strconv.FormatInt(s.Id, IdBase)
}

// Represent Short URL in pretty format
func (s Shorturl) String() string {
	return fmt.Sprintf(
		"http://%s/%s (added %s)\n Target: %s\n IP: %s",
		Domain, s.Uid(), s.Added, s.URL, s.Host)
}

// Look up shorturl and print on console
func main() {
	// Check command has exactly one argument or print usage
	if args := os.Args[1:]; len(args) != 1 {
		log.Fatal(usage)
	}

	db, err := sql.Open("postgres", connString)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	uid := os.Args[1]
	s, err := getShorturl(db, uid)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
}
