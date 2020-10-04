package shorturl

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/joneskoo/shorturl-go/assets"
	"golang.org/x/net/context"
)

func Handler(db *DB, secure bool) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", serveHome(db))
	mux.Handle("/p/", http.StripPrefix("/p", servePreview(db)))
	mux.Handle("/static/style.css", serveStatic("css/style.css"))
	// apply middlewares
	var h http.Handler = mux
	h = setContextSecure(h, secure)
	return h
}

type response struct {
	// Template is the file name of the template to render.
	Template string
	// StatusCode is the response status code.
	StatusCode int
	// Context is the template context used to render the HTML template.
	Context interface{}
}

func (r response) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	template, ok := templates[r.Template]
	if !ok {
		log.Printf("template %s not found", r.Template)
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(r.StatusCode)
	protocol := "http://"
	secure := req.Context().Value(contextSecure)
	if secure, ok := secure.(bool); ok && secure {
		protocol = "https://"
	}

	err := template.Execute(rw, map[string]interface{}{
		"Protocol": protocol,
		"Domain":   Domain,
		"Data":     r.Context,
	})
	if err != nil {
		log.Printf("error executing template %s: %v", r.Template, err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

var errorEOL = response{
	Template:   "error.html",
	StatusCode: http.StatusGone,
	Context: map[string]string{
		"ErrorTitle":   "Service is end-of-life",
		"ErrorMessage": "This short url service is end of life. Existing redirects continue to work for now.",
	},
}

var errorNotFound = response{
	Template:   "error.html",
	StatusCode: http.StatusNotFound,
	Context: map[string]string{
		"ErrorTitle":   "Short URL not found",
		"ErrorMessage": "Short URL by this id was not found.",
	},
}

var internalError = response{
	Template:   "error.html",
	StatusCode: 500,
	Context: map[string]string{
		"ErrorTitle":   "Internal server error",
		"ErrorMessage": "There was an error and we failed to handle it. Sorry.",
	},
}

func serveHome(db *DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/":
			errorEOL.ServeHTTP(w, req)
		default:
			serveRedirect(db).ServeHTTP(w, req)
		}
	})
}

func serveRedirect(db *DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if isAlwaysPreview(req) && !isLocalReferer(req) {
			servePreview(db).ServeHTTP(w, req)
			return
		}
		shortCode := req.URL.Path[1:]
		s, err := db.Get(shortCode)
		switch err {
		case ErrNotFound:
			errorNotFound.ServeHTTP(w, req)
		case nil:
			http.Redirect(w, req, s.URL, http.StatusFound)
		default:
			internalError.ServeHTTP(w, req)
		}
	})
}

func isLocalReferer(req *http.Request) bool {
	url, err := url.Parse(req.Referer())
	if err != nil {
		return false
	}
	return strings.EqualFold(url.Host, Domain)
}

// Preview shows short url details after adding
func servePreview(db *DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s, err := db.Get(req.URL.Path[1:])
		switch err {
		case ErrNotFound:
			errorNotFound.ServeHTTP(w, req)
		case nil:
			response{
				Template:   "preview.html",
				Context:    s,
				StatusCode: http.StatusOK,
			}.ServeHTTP(w, req)
		default:
			internalError.ServeHTTP(w, req)
		}
	})
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

func serveStatic(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		style := assets.MustAsset(name)
		http.ServeContent(w, req, path.Base(name), assets.LastModified, bytes.NewReader(style))
	})
}

type contextKey int

const (
	contextSecure contextKey = iota
)

func setContextSecure(h http.Handler, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), contextSecure, secure)
		h.ServeHTTP(w, req.WithContext(ctx))
	})
}
