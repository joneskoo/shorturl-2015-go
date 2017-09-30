package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/joneskoo/shorturl-go/database"
	"github.com/joneskoo/shorturl-go/handlers"
)

var allowedURLSchemes = []string{"http", "https", "ftp", "ftps", "feed", "gopher", "magnet", "spotify"}

// globals
var (
	db            *database.Database
	secure        bool
	csrfStateFile = "csrf.secret"
	listenAddr    = "127.0.0.1:39284"
)

func main() {
	flag.BoolVar(&secure, "secure", false, "set secure (HTTPS) flag in cookies")
	flag.StringVar(&database.ConnString, "connstring", database.ConnString, "PostgreSQL connection string")
	flag.StringVar(&csrfStateFile, "csrf-file", csrfStateFile, "file to store CSRF secret in")
	flag.StringVar(&listenAddr, "listen", listenAddr, "listen on [host]:port")
	flag.StringVar(&database.Domain, "domain", database.Domain, "domain name")
	flag.Parse()
	log.Printf("Starting server, os.Args=%s", strings.Join(os.Args, " "))

	var err error
	if db, err = database.New(); err != nil {
		log.Fatalf("Connecting to database: %v", err)
	}

	log.Print("Listening on http://", listenAddr)

	secret, err := csrfSecret()
	if err != nil {
		log.Fatalf("Loading CSRF: %v", err)
	}
	CSRF := csrf.Protect(secret, csrf.Secure(secure))

	h := handlers.New(db, secure)
	if err := http.ListenAndServe(listenAddr, CSRF(h)); err != nil {
		log.Fatal(err)
	}
}

func csrfSecret() ([]byte, error) {
	if _, err := os.Stat(csrfStateFile); os.IsNotExist(err) {
		randBytes := make([]byte, 32)
		if _, err := rand.Read(randBytes); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(csrfStateFile, randBytes, 0600); err != nil {
			return nil, fmt.Errorf("failed to write csrf to file: %v", err)
		}
	}
	csrfSecret, err := ioutil.ReadFile(csrfStateFile)
	if len(csrfSecret) != 32 {
		return nil, fmt.Errorf("CSRF secret file must be 32 bytes, got %d", len(csrfSecret))
	}
	return csrfSecret, err
}
