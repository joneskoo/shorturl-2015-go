package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joneskoo/shorturl-go/database"
	"github.com/joneskoo/shorturl-go/yxfi-server/assets"
)

var allowedURLSchemes = []string{"http", "https", "ftp", "ftps", "feed", "gopher", "magnet", "spotify"}

// globals
var (
	db            *database.Database
	secure        bool
	csrfStateFile = "csrf.secret"
	listenAddr    = "127.0.0.1:39284"
)

func main() {
	flag.BoolVar(&secure, "secure", false, "set secure (HTTPS) flag in cookies")
	flag.StringVar(&database.ConnString, "connstring", database.ConnString, "PostgreSQL connection string")
	flag.StringVar(&csrfStateFile, "csrf-file", csrfStateFile, "file to store CSRF secret in")
	flag.StringVar(&listenAddr, "listen", listenAddr, "listen on [host]:port")
	flag.Parse()
	log.Printf("Starting server, os.Args=%s", strings.Join(os.Args, " "))

	var err error
	if db, err = database.New(); err != nil {
		log.Fatalf("Connecting to database: %v", err)
	}

	err = parseHTMLTemplates([][]string{
		{"error.html", "layout.html"},
		{"index.html", "layout.html"},
		{"404.html", "layout.html"},
		{"preview.html", "layout.html"},
	})
	if err != nil {
		log.Fatalf("Parsing HTML templates: %v", err)
	}

	log.Print("Listening on http://", listenAddr)

	r := mux.NewRouter()
	r.Handle("/favicon.ico", handler(serveFavico))
	r.Handle("/", handler(serveHome))
	r.Handle("/{key}", handler(serveRedirect))
	r.Handle("/add/", handler(serveAdd))
	r.Handle("/p/{key}", http.StripPrefix("/p/", handler(servePreview)))
	r.Handle("/static/style.css", handler(serveCSS))
	r.Handle("/always-preview/enable", handler(serveAlwaysPreview))
	r.Handle("/always-preview/disable", handler(serveAlwaysPreview))

	secret, err := csrfSecret()
	if err != nil {
		log.Fatalf("Loading CSRF: %v", err)
	}
	CSRF := csrf.Protect(secret, csrf.Secure(secure))

	if err := http.ListenAndServe(listenAddr, CSRF(r)); err != nil {
		log.Fatal(err)
	}
}

func csrfSecret() ([]byte, error) {
	if _, err := os.Stat(csrfStateFile); os.IsNotExist(err) {
		randBytes := make([]byte, 32)
		if _, err := rand.Read(randBytes); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(csrfStateFile, randBytes, 0600); err != nil {
			return nil, fmt.Errorf("failed to write csrf to file: %v", err)
		}
	}
	csrfSecret, err := ioutil.ReadFile(csrfStateFile)
	if len(csrfSecret) != 32 {
		return nil, fmt.Errorf("CSRF secret file must be 32 bytes, got %d", len(csrfSecret))
	}
	return csrfSecret, err
}

type handler func(resp http.ResponseWriter, req *http.Request) error

func (h handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	err := h(resp, req)

	var statusCode int
	var errorTitle, errorMessage string
	switch err {
	case nil:
		return
	case database.ErrNotFound:
		statusCode = http.StatusNotFound
		errorTitle = "Short URL not found"
		errorMessage = "Short URL by this id was not found."
	default:
		log.Printf("Unhandled error: %v", err)
		statusCode = http.StatusInternalServerError
		errorTitle = "Internal server error"
		errorMessage = "There was an error and we failed to handle it. Sorry."
	}
	if err := executeTemplate(resp, "error.html", statusCode, nil, map[string]string{
		"ErrorTitle":   errorTitle,
		"ErrorMessage": errorMessage,
	}); err != nil {
		log.Printf("Error sending error response: %v", err)
	}
}

// setAlwaysPreview sets the preview cookie which forces plain
// shorturls to show preview page instead.
func serveAlwaysPreview(resp http.ResponseWriter, req *http.Request) error {
	switch req.URL.Path {
	case "/always-preview/enable":
		setAlwaysPreview(resp, req)
	case "/always-preview/disable":
		unsetAlwaysPreview(resp, req)
	}
	http.Redirect(resp, req, "/", http.StatusFound)
	return nil
}

func setAlwaysPreview(resp http.ResponseWriter, req *http.Request) {
	cookie := http.Cookie{
		Name:   "preview",
		Value:  "true",
		Path:   "/",
		MaxAge: 86400 * 365 * 10, // 10 years
	}
	http.SetCookie(resp, &cookie)
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
}

func serveHome(resp http.ResponseWriter, req *http.Request) error {
	if req.URL.Path != "/" {
		return serveRedirect(resp, req)
	}

	return executeTemplate(resp, "index.html", http.StatusOK, nil,
		map[string]interface{}{csrf.TemplateTag: csrf.TemplateField(req)})
}

func serveRedirect(resp http.ResponseWriter, req *http.Request) error {
	if isAlwaysPreview(req) {
		return servePreview(resp, req)
	}
	uid := req.URL.Path[1:]
	s, err := db.GetByUID(uid)
	if err != nil {
		return err
	}
	http.Redirect(resp, req, s.URL, http.StatusMovedPermanently)
	return nil
}

func executeTemplate(resp http.ResponseWriter, name string, status int, header http.Header, data interface{}) error {
	template, ok := templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	resp.WriteHeader(status)
	err := template.Execute(resp, data)
	if err != nil {
		log.Printf("Executing template %s: %v", name, err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
	return err
}

func isAlwaysPreview(req *http.Request) bool {
	cookies := req.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "preview" && cookie.Value == "true" {
			return true
		}
	}
	return false
}

func serveCSS(resp http.ResponseWriter, req *http.Request) error {
	r := bytes.NewReader(assets.MustAsset("css/style.css"))
	http.ServeContent(resp, req, "style.css", assets.LastModified, r)
	return nil
}

// Preview shows short url details after adding
func servePreview(resp http.ResponseWriter, req *http.Request) error {
	s, err := db.GetByUID(req.URL.Path)
	if err != nil {
		return err
	}
	return executeTemplate(resp, "preview.html", http.StatusOK, nil, s)
}

func checkURLScheme(urlString string) error {
	u, err := url.Parse(urlString)
	if err != nil {
		return err
	}
	for _, scheme := range allowedURLSchemes {
		if u.Scheme == scheme {
			return nil
		}
	}
	return fmt.Errorf("URL scheme %q not allowed", u.Scheme)
}

func checkAllowed(req *http.Request, url string, host string) error {
	switch {
	case len(url) < 20:
		return fmt.Errorf("URL too short for shortening")
	case len(url) > 2048:
		return fmt.Errorf("URL too long for shortening")
	}
	return checkURLScheme(url)
}

// Add adds a new shorturl
func serveAdd(resp http.ResponseWriter, req *http.Request) error {
	url := req.FormValue("url")
	if url == "" {
		http.Redirect(resp, req, "/", http.StatusFound)
		return nil
	}
	host := getIP(req)
	if err := checkAllowed(req, url, host); err != nil {
		return executeTemplate(resp, "index.html", http.StatusOK, nil,
			map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(req),
				"Error":          err,
			})
	}
	clientid := getClientID(req)
	if clientid == "" {
		err := fmt.Errorf("Failed to add shorturl")
		return executeTemplate(resp, "index.html", http.StatusForbidden, nil,
			map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(req),
				"Error":          err,
			})

	}
	s, err := db.Add(url, host, clientid)
	if err != nil {
		return err
	}
	http.Redirect(resp, req, s.PreviewURL(), http.StatusFound)
	return nil
}

func getIP(req *http.Request) string {
	// FIXME: make configurable
	// return req.RemoteAddr
	return req.Header.Get("x-forwarded-for")
}

// List lists short URLs
func serveList(resp http.ResponseWriter, r *http.Request) error {
	shorturls, err := db.List()
	if err != nil {
		return err
	}
	for s := range shorturls {
		if s.ID == 0 {
			return nil
		}
		encoder := json.NewEncoder(resp)
		err := encoder.Encode(s)
		if err != nil {
			return err
		}
		fmt.Fprint(resp, "\n")
	}
	return nil
}

func serveFavico(resp http.ResponseWriter, req *http.Request) error {
	clientid := getClientID(req)

	// Ensure clientid is set
	if clientid == "" {
		cookie := http.Cookie{
			Name:   "clientid",
			Value:  generateClientID(),
			Path:   "/",
			MaxAge: 86400 * 365 * 10, // 10 years
		}
		http.SetCookie(resp, &cookie)
	}
	http.NotFound(resp, req)
	return nil
}

func getClientID(req *http.Request) string {
	cookies := req.Cookies()
	clientid := ""
	for _, cookie := range cookies {
		if cookie.Name == "clientid" {
			matched, err := regexp.MatchString("^[a-f0-9]{32}$", cookie.Value)
			if err == nil && matched {
				clientid = cookie.Value
			}
		}
	}
	return clientid
}

func generateClientID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Failed to generate client id: %v", err)
		return ""
	}
	return hex.EncodeToString(b)
}
