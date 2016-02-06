package main

import (
	"fmt"
	"log"
	"os"
	"shorturl/shorturl"
)

const usage = "Usage:\n  shorturl-add <url>"

// Look up shorturl and print on console
func main() {
	// Check command has exactly one argument or print usage
	if args := os.Args[1:]; len(args) != 1 {
		log.Fatal(usage)
	}

	db, err := shorturl.ConnectToDatabase()
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	url := os.Args[1]
	host := "::1"

	s, err := shorturl.Add(db, url, host)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
}
