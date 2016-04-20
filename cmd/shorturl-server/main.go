package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"io/ioutil"
	"crypto/rand"
	"github.com/gorilla/mux"
	"github.com/gorilla/csrf"

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

	csrfSecretFile := path.Join(contentRoot, "csrf.secret")
	if _, err := os.Stat(csrfSecretFile); os.IsNotExist(err) {
		randBytes := make([]byte, 32)
		_, err := rand.Read(randBytes)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(csrfSecretFile, randBytes, 0600)
	}
	csrfSecret, err := ioutil.ReadFile(csrfSecretFile)
	if err != nil {
		log.Fatal(err)
	}
	if len(csrfSecret) != 32 {
		panic("CSRF secret file must be 32 bytes")
	}

	addr := "[::1]:8000"
	log.Print("Listening on", addr)
	view := shorturl.NewView(contentRoot, db)
	
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Redirect yx.fi/xxx to target
		if req.URL.Path != "/" {
			view.Redirect(w, req)
			return
		}
		// For root URI /, serve index page
		view.Index(w, req)
	})
	r.HandleFunc("/add/", view.Add)
	r.Handle("/p/{key}", http.StripPrefix("/p/", http.HandlerFunc(view.Preview)))

	// Static files
	staticHandler := http.FileServer(http.Dir(path.Join(contentRoot, "static")))
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", staticHandler))

	r.HandleFunc("/always-preview/enable", setAlwaysPreview)
	r.HandleFunc("/always-preview/disable", unsetAlwaysPreview)
	CSRF := csrf.Protect([]byte(csrfSecret), csrf.Secure(false))
	http.ListenAndServe(addr, CSRF(r))
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