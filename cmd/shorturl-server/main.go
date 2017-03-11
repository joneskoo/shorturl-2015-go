package main

import (
	"crypto/rand"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"

	shorturl "github.com/joneskoo/shorturl-go"
)

func main() {
	db, err := shorturl.ConnectToDatabase()
	if err != nil {
		log.Fatal(err)
	}

	csrfSecretFile := "csrf.secret"
	if _, err := os.Stat(csrfSecretFile); os.IsNotExist(err) {
		randBytes := make([]byte, 32)
		_, err := rand.Read(randBytes)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(csrfSecretFile, randBytes, 0600)
	}
	csrfSecret, err := ioutil.ReadFile(csrfSecretFile)
	// FIXME: use secure cookie (default)
	CSRF := csrf.Protect([]byte(csrfSecret), csrf.Secure(false))

	if err != nil {
		log.Fatal(err)
	}
	if len(csrfSecret) != 32 {
		panic("CSRF secret file must be 32 bytes")
	}

	addr := "0.0.0.0:39284"
	log.Print("Listening on http://", addr)
	view := shorturl.NewView(db)

	r := mux.NewRouter()
	r.HandleFunc("/favicon.ico", view.FaviconHandler)
	r.HandleFunc("/index.html", view.Index)
	r.HandleFunc("/", view.Index)
	r.HandleFunc("/{key}", view.Redirect)
	r.HandleFunc("/add/", view.Add)
	r.Handle("/p/{key}", http.StripPrefix("/p/", http.HandlerFunc(view.Preview)))
	r.HandleFunc("/static/style.css", view.Static)
	r.HandleFunc("/always-preview/enable", setAlwaysPreview)
	r.HandleFunc("/always-preview/disable", unsetAlwaysPreview)
	if err := http.ListenAndServe(addr, CSRF(r)); err != nil {
		log.Fatal(err)
	}
}

// setAlwaysPreview sets the preview cookie which forces plain
// shorturls to show preview page instead.
func setAlwaysPreview(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:   "preview",
		Value:  "true",
		Path:   "/",
		MaxAge: 86400 * 365 * 10, // 10 years
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

// unsetAlwaysPreview sets the preview cookie which forces plain
// shorturls to show preview page instead.
func unsetAlwaysPreview(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:   "preview",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}
