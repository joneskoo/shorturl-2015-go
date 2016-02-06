package main

import (
	"log"
	"net/http"
	"os"
	"shorturl/shorturl"
	"shorturl/views"
)

func main() {
	var err error
	db, err := shorturl.ConnectToDatabase()
	if err != nil {
		log.Fatal(err)
	}

	contentRoot := "."
	if len(os.Args) >= 2 {
		contentRoot = os.Args[1]
	}

	addr := "[::1]:8000"
	log.Print("Listening on", addr)
	view := views.NewView(contentRoot, db)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Redirect yx.fi/xxx to target
		if req.URL.Path != "/" {
			view.Redirect(w, req)
			return
		}
		// For root URI /, serve index page
		view.Index(w, req)
	})
	http.Handle("/p/", http.StripPrefix("/p", http.HandlerFunc(view.Preview)))
	http.Handle("/add/", http.HandlerFunc(view.Add))
	http.Handle("/static/", http.FileServer(http.Dir(contentRoot)))
	http.ListenAndServe(addr, nil)
}
