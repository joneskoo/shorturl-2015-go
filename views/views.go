package views

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"shorturl/shorturl"
	"strings"
	"time"
	"unicode/utf8"

	"html/template"
)

// View is the base for all views
type View struct {
	DB        *sql.DB
	templates *template.Template
}

func (view View) renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := view.templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// NewView initializes a base view. This can be then cast to
// other views.
func NewView(contentRoot string, db *sql.DB) *View {
	templates := template.New("main")
	templateGlob := path.Join(contentRoot, "templates", "*.html")
	funcMap := template.FuncMap{
		"truncate":   truncate,
		"formattime": formatTime,
	}
	templates = template.Must(templates.Funcs(funcMap).ParseGlob(templateGlob))

	v := View{
		DB:        db,
		templates: templates}
	return &v
}

// Index serves the main page
func (view View) Index(w http.ResponseWriter, req *http.Request) {
	view.renderTemplate(w, "index", nil)
}

// Redirect redirects to short URL target
func (view View) Redirect(w http.ResponseWriter, req *http.Request) {
	uid := req.URL.Path[1:]
	s, err := shorturl.GetByUID(view.DB, uid)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	http.Redirect(w, req, s.URL, http.StatusMovedPermanently)
}

// Preview shows short url details after adding
func (view View) Preview(w http.ResponseWriter, req *http.Request) {
	uid := req.URL.Path[1:]
	s, err := shorturl.GetByUID(view.DB, uid)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	view.renderTemplate(w, "preview", &s)
}

// Add adds a new shorturl
func (view View) Add(w http.ResponseWriter, req *http.Request) {
	url := req.FormValue("url")
	host := getIP(req)
	s, err := shorturl.Add(view.DB, url, host)
	if err != nil {
		http.Error(w, "Failed to add", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, s.PreviewURL(), http.StatusFound)
}

// List lists short URLs
func (view View) List(w http.ResponseWriter, r *http.Request) {
	shorturls, err := shorturl.List(view.DB)
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

func getIP(req *http.Request) string {
	return req.RemoteAddr
}

// truncate limits the string to 25 unicode characters
func truncate(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	n := 0
	reader := strings.NewReader(s)
	// calculate number of bytes for limit-1 runes
	for i := 0; i < limit-1; i++ {
		_, size, err := reader.ReadRune()
		if err != nil {
			break // unexpected end of string
		}
		n += size
	}
	return s[:n] + "…"
}

func formatTime(t *time.Time, format string) string {
	return t.Format(format)
}
