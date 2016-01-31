package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"shorturl/shorturl"
)

var db *sql.DB

func hello(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Path[1:]
	s, err := shorturl.GetByUID(db, uid)
	if err != nil {
		io.WriteString(w, "Not found\n")
		return
	}
	io.WriteString(w, s.String())
	io.WriteString(w, "\n")
}

func list(w http.ResponseWriter, r *http.Request) {
	shorturls, err := shorturl.List(db)
	if err != nil {
		io.WriteString(w, "Not found\n")
		io.WriteString(w, err.Error())
		return
	}
	for {
		s := <-shorturls
		if s.ID == 0 {
			return
		}
		jsonBytes, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}
		w.Write(jsonBytes)
		fmt.Fprintf(w, "\n")
	}
}

func main() {
	var err error
	db, err = shorturl.ConnectToDatabase()
	if err != nil {
		log.Fatal(err)
	}

	addr := "[::1]:8000"
	log.Print("Listening on", addr)
	http.HandleFunc("/", hello)
	http.ListenAndServe(addr, nil)
}
