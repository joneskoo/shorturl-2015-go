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
	db            *shorturl.DB
	secure        bool
	csrfStateFile = "csrf.secret"
	listenAddr    = "127.0.0.1:39284"
)

func main() {
	flag.BoolVar(&secure, "secure", false, "use https URLs and set secure flag in cookies")
	connstring := flag.String("connstring", "user=joneskoo dbname=joneskoo sslmode=disable", "PostgreSQL connection string")
	flag.StringVar(&csrfStateFile, "csrf-file", csrfStateFile, "file to store CSRF secret in")
	flag.StringVar(&listenAddr, "listen", listenAddr, "listen on [host]:port")
	flag.Parse()
	log.Printf("Starting server, os.Args=%s", strings.Join(os.Args, " "))

	var err error
	if db, err = shorturl.Open(*connstring); err != nil {
		log.Fatalf("Connecting to database: %v", err)
	}

	log.Print("Listening on http://", listenAddr)

	h := shorturl.Handler(db, secure)
	if err := http.ListenAndServe(listenAddr, h); err != nil {
		log.Fatal(err)
	}
}
