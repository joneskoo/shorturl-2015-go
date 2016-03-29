package shorturl

import (
	"net/http"
	"path"
    "strconv"
	"time"
    "google.golang.org/appengine"
    "google.golang.org/appengine/datastore"

	"html/template"
)

// View is the base for all views
type View struct {
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
func NewView(contentRoot string) *View {
	templates := template.New("main")
	templateGlob := path.Join(contentRoot, "templates", "*.html")
	funcMap := template.FuncMap{
		"truncate":   truncate,
		"formattime": formatTime,
	}
	templates = template.Must(templates.Funcs(funcMap).ParseGlob(templateGlob))

	v := View{templates}
	return &v
}

// Index serves the main page
func (view View) Index(w http.ResponseWriter, req *http.Request) {
	view.renderTemplate(w, "index", nil)
}

func getShorturl(req *http.Request) (*Shorturl, error) {
	uid := req.URL.Path[1:]

    ctx := appengine.NewContext(req)
    var s Shorturl

    id, err := strconv.ParseInt(uid, idBase, 64)
    if err != nil {
            return nil, err
    }

    key := datastore.NewKey(ctx, "Shorturl", "", id, nil)
	if err != nil {
		return nil, ErrNotFound
	}
    if err = datastore.Get(ctx, key, &s); err != nil {
        return nil, err
    }
    s.ID = id
    s.ServiceDomain = appengine.DefaultVersionHostname(ctx)
    return &s, nil
}

// Redirect redirects to short URL target
func (view View) Redirect(w http.ResponseWriter, req *http.Request) {
    s, err := getShorturl(req)
    if err == ErrNotFound {
        http.NotFound(w, req)
        return
    } else if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	http.Redirect(w, req, s.URL, http.StatusMovedPermanently)
}

// Preview shows short url details after adding
func (view View) Preview(w http.ResponseWriter, req *http.Request) {
    s, err := getShorturl(req)
    if err == ErrNotFound {
        http.NotFound(w, req)
        return
    } else if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	view.renderTemplate(w, "preview", &s)
}

// Add adds a new shorturl
func (view View) Add(w http.ResponseWriter, req *http.Request) {
	url := req.FormValue("url")
	host := getIP(req)

	//s, err := Add(url, host)
    s := Shorturl{
    	URL: url,
	    Host: host,
    	Added: time.Now()}

    ctx := appengine.NewContext(req)
    key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "Shorturl", nil), &s)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    s.ID = key.IntID()
	http.Redirect(w, req, s.PreviewURL(), http.StatusFound)
}

func getIP(req *http.Request) string {
	return req.RemoteAddr
}

func formatTime(t *time.Time, format string) string {
	return t.Format(format)
}
