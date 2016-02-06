package main

import (
	"fmt"
	"log"
	"os"
	"shorturl/shorturl"
)

const usage = "Usage:\n  shorturl <uid>"

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

	uid := os.Args[1]
	s, err := shorturl.GetByUID(db, uid)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(s)
}
