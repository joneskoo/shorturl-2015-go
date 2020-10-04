package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joneskoo/shorturl-go"
)

var allowedURLSchemes = []string{"http", "https", "ftp", "ftps", "feed", "gopher", "magnet", "spotify"}

// globals
var (
	db            *shorturl.Database
	secure        bool
	csrfStateFile = "csrf.secret"
	listenAddr    = "127.0.0.1:39284"
)

func main() {
	flag.BoolVar(&secure, "secure", false, "set secure (HTTPS) flag in cookies")
	flag.StringVar(&shorturl.ConnString, "connstring", shorturl.ConnString, "PostgreSQL connection string")
	flag.StringVar(&csrfStateFile, "csrf-file", csrfStateFile, "file to store CSRF secret in")
	flag.StringVar(&listenAddr, "listen", listenAddr, "listen on [host]:port")
	flag.StringVar(&shorturl.Domain, "domain", shorturl.Domain, "domain name")
	flag.Parse()
	log.Printf("Starting server, os.Args=%s", strings.Join(os.Args, " "))

	var err error
	if db, err = shorturl.New(); err != nil {
		log.Fatalf("Connecting to database: %v", err)
	}

	log.Print("Listening on http://", listenAddr)

	h := shorturl.NewHandlers(db, secure)
	if err := http.ListenAndServe(listenAddr, h); err != nil {
		log.Fatal(err)
	}
}
