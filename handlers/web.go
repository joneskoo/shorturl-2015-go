package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/joneskoo/shorturl-go/assets"
	"github.com/joneskoo/shorturl-go/database"
)

var allowedURLSchemes = []string{"http", "https", "ftp", "ftps", "feed", "gopher", "magnet", "spotify"}

type handler struct {
	db     *database.Database
	secure bool
	*http.ServeMux
}

func New(db *database.Database, secure bool) http.Handler {
	mux := http.NewServeMux()
	h := handler{db, secure, mux}
	mux.HandleFunc("/", h.serveHome)
	mux.Handle("/p/", http.StripPrefix("/p", http.HandlerFunc(h.servePreview)))
	mux.HandleFunc("/add/", h.serveAdd)
	mux.HandleFunc("/always-preview/enable", h.serveAlwaysPreview)
	mux.HandleFunc("/always-preview/disable", h.serveAlwaysPreview)
	mux.HandleFunc("/favicon.ico", h.serveFavico)
	mux.HandleFunc("/static/style.css", h.serveCSS)
	return &h
}

type errorResponse struct {
	ErrorTitle   string
	ErrorMessage string
	StatusCode   int
	Template     string
}

var errorNotFound = errorResponse{
	StatusCode:   http.StatusNotFound,
	ErrorTitle:   "Short URL not found",
	ErrorMessage: "Short URL by this id was not found.",
}

var internalError = errorResponse{
	StatusCode:   500,
	ErrorTitle:   "Internal server error",
	ErrorMessage: "There was an error and we failed to handle it. Sorry.",
}

func (h handler) serverError(w http.ResponseWriter, e errorResponse) {
	h.executeTemplate(w, "error.html", e.StatusCode, nil, e)
}

func (h handler) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.serveRedirect(w, r)
		return
	}

	data := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	}
	h.executeTemplate(w, "index.html", http.StatusOK, nil, data)
}

func (h handler) serveRedirect(w http.ResponseWriter, r *http.Request) {
	if isAlwaysPreview(r) && !isLocalReferer(r) {
		h.servePreview(w, r)
		return
	}
	shortCode := r.URL.Path[1:]
	s, err := h.db.Get(shortCode)
	switch err {
	case nil:
		h.executeTemplate(w, "preview.html", http.StatusOK, nil, s)
	case database.ErrNotFound:
		h.serverError(w, errorNotFound)
	default:
		h.serverError(w, internalError)
	}
	http.Redirect(w, r, s.URL, http.StatusFound)
}

func isLocalReferer(req *http.Request) bool {
	url, err := url.Parse(req.Referer())
	if err != nil {
		return false
	}
	return strings.EqualFold(url.Host, database.Domain)
}

// Preview shows short url details after adding
func (h handler) servePreview(w http.ResponseWriter, r *http.Request) {
	s, err := h.db.Get(r.URL.Path[1:])
	switch err {
	case nil:
		h.executeTemplate(w, "preview.html", http.StatusOK, nil, s)
	case database.ErrNotFound:
		h.serverError(w, errorNotFound)
	default:
		h.serverError(w, internalError)
	}
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

// Add stores a new shorturl to database or returns the existing
// if the same URL was already in database.
//
// In either case, the view redirects to the preview page showing
// the details and when the URL was first added.
func (h handler) serveAdd(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	host := getIP(r)
	if err := checkAllowed(r, url, host); err != nil {
		h.executeTemplate(w, "index.html", http.StatusOK, nil,
			map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(r),
				"Error":          err,
			})
		return
	}
	clientid := getClientID(r)
	if clientid == "" {
		err := fmt.Errorf("Failed to add shorturl")
		h.executeTemplate(w, "index.html", http.StatusForbidden, nil,
			map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(r),
				"Error":          err,
			})
		return
	}
	s, err := h.db.Add(url, host, clientid)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Printf("unhandled error: %v", err)
		return
	}
	http.Redirect(w, r, s.PreviewURL(), http.StatusFound)
}

func getIP(req *http.Request) string {
	// FIXME: make configurable
	// return req.RemoteAddr
	return req.Header.Get("x-forwarded-for")
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

func (h handler) serveList(w http.ResponseWriter, r *http.Request) {
	shorturls, err := h.db.List()
	if err != nil {
		log.Printf("unhandled error: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	for s := range shorturls {
		if s.ID == 0 {
			return // FIXME: what does this do?
		}
		err := json.NewEncoder(w).Encode(s)
		if err != nil {
			log.Printf("failed to encode response to output: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		_, _ = fmt.Fprint(w, "\n")
	}
}

// setAlwaysPreview sets the preview cookie which forces plain
// shorturls to show preview page instead.
func (h handler) serveAlwaysPreview(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/always-preview/enable":
		setAlwaysPreview(w, r)
		return
	case "/always-preview/disable":
		unsetAlwaysPreview(w, r)
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return
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

func (h handler) serveFavico(w http.ResponseWriter, r *http.Request) {
	clientid := getClientID(r)

	// Ensure clientid is set
	if clientid == "" {
		cookie := http.Cookie{
			Name:   "clientid",
			Value:  generateClientID(),
			Path:   "/",
			MaxAge: 86400 * 365 * 10, // 10 years
		}
		http.SetCookie(w, &cookie)
	}
	http.NotFound(w, r)
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

func (h handler) serveCSS(w http.ResponseWriter, r *http.Request) {
	style := assets.MustAsset("css/style.css")
	http.ServeContent(w, r, "style.css", assets.LastModified, bytes.NewReader(style))
}

func (h handler) executeTemplate(w http.ResponseWriter, name string, status int, header http.Header, data interface{}) {
	template, ok := templates[name]
	if !ok {
		log.Printf("template %s not found", name)
		http.Error(w, "", http.StatusInternalServerError)
	}
	w.WriteHeader(status)
	protocol := "http://"
	if h.secure {
		protocol = "https://"
	}

	err := template.Execute(w, struct {
		Protocol string
		Domain   string
		Data     interface{}
	}{protocol, database.Domain, data})
	if err != nil {
		log.Printf("error executing template %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
