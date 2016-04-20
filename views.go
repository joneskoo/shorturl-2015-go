package shorturl

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"errors"
	"io"
	"net/http"
	"path"
	"github.com/gorilla/csrf"
	nurl "net/url"

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

var allowedURLSchemes = []string{"http", "https", "ftp", "ftps", "feed", "gopher", "magnet", "spotify"}

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
	view.renderTemplate(w, "index", map[string]interface{}{
        csrf.TemplateTag: csrf.TemplateField(req),
    })
}

func isAlwaysPreview(req *http.Request) bool {
	cookies := req.Cookies()
	for _, cookie := range(cookies) {
		if cookie.Name == "preview" && cookie.Value == "true" {
			return true
		}
	}
	return false
}

// Redirect redirects to short URL target
func (view View) Redirect(w http.ResponseWriter, req *http.Request) {
	if isAlwaysPreview(req) {
		view.Preview(w, req)
		return
	}
	uid := req.URL.Path[1:]
	s, err := GetByUID(view.DB, uid)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	http.Redirect(w, req, s.URL, http.StatusMovedPermanently)
}

// Preview shows short url details after adding
func (view View) Preview(w http.ResponseWriter, req *http.Request) {
	s, err := GetByUID(view.DB, req.URL.Path)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	view.renderTemplate(w, "preview", &s)
}

func checkURLScheme(url string) error {
	u, err := nurl.Parse(url)
	if err != nil {
		return err
	}
	for _, scheme := range allowedURLSchemes {
		if u.Scheme == scheme {
			return nil
		}
	}
	return errors.New("URL scheme not allowed")
}

func checkAllowed(req *http.Request, url string, host string) error {
	switch {
		case len(url) < 20:
			return errors.New("URL too short for shortening")
		case len(url) > 2048:
			return errors.New("URL too long for shortening")
	}
	if err := checkURLScheme(url); err != nil {
		return err
	}
	return nil
}

// Add adds a new shorturl
func (view View) Add(w http.ResponseWriter, req *http.Request) {
	url := req.FormValue("url")
	if url == "" {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}
	host := getIP(req)
	if err := checkAllowed(req, url, host); err != nil {
		data := map[string]interface{}{
        csrf.TemplateTag: csrf.TemplateField(req),
		"Error": err,
    	}
		view.renderTemplate(w, "index", data)
		return
	}
	s, err := Add(view.DB, url, host)
	if err != nil {
		http.Error(w, "Failed to add", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, s.PreviewURL(), http.StatusFound)
}

// List lists short URLs
func (view View) List(w http.ResponseWriter, r *http.Request) {
	shorturls, err := List(view.DB)
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
