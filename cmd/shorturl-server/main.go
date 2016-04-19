package main

import (
	"log"
	"net/http"
	"os"
	shorturl "github.com/joneskoo/shorturl-go"
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
	view := shorturl.NewView(contentRoot, db)
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
	http.HandleFunc("/always-preview/enable", setAlwaysPreview)
	http.HandleFunc("/always-preview/disable", unsetAlwaysPreview)
	http.ListenAndServe(addr, nil)
}


// setAlwaysPreview sets the preview cookie which forces plain
// shorturls to show preview page instead.
func setAlwaysPreview(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name: "preview",
		Value: "true",
		Path: "/",
		MaxAge: 86400 * 365 * 10, // 10 years
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

// unsetAlwaysPreview sets the preview cookie which forces plain
// shorturls to show preview page instead.
func unsetAlwaysPreview(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name: "preview",
		Value: "",
		Path: "/",
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}