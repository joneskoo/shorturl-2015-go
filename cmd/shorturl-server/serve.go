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
	base := views.NewView(contentRoot, db)
	http.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path != "/" {
				views.RedirectView{base}.ServeHTTP(w, req)
				return
			}
			views.IndexView{base}.ServeHTTP(w, req)
		})
	http.Handle("/p/", http.StripPrefix("/p", views.PreviewView{base}))
	http.ListenAndServe(addr, nil)
}
