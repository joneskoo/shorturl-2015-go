package shorturl

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/joneskoo/shorturl-go/assets"
)

func Handler(db *DB, secure bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/":
			// Index page: This service is end of life.
			errorEOL.ServeHTTP(w, req)
		default:
			// If shorturl exists, redirect to it.
			shorturlHandler(db).ServeHTTP(w, req)
		}
	})
	mux.Handle("/p/", http.StripPrefix("/p", previewHandler(db)))
	mux.Handle("/static/style.css", staticHandler("css/style.css"))
	return mux
}

func shorturlHandler(db *DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if alwaysPreviewPref(req) && !isLocalReferer(req) {
			previewHandler(db).ServeHTTP(w, req)
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

// Preview shows short url details after adding
func previewHandler(db *DB) http.Handler {
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

func staticHandler(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		style := assets.MustAsset(name)
		http.ServeContent(w, req, path.Base(name), assets.LastModified, bytes.NewReader(style))
	})
}

func isLocalReferer(req *http.Request) bool {
	url, err := url.Parse(req.Referer())
	if err != nil {
		return false
	}
	return strings.EqualFold(url.Host, req.Host)
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
	if isSecure(req) {
		protocol = "https://"
	}

	err := template.Execute(rw, map[string]interface{}{
		"Protocol": protocol,
		"Domain":   req.Host,
		"Data":     r.Context,
	})
	if err != nil {
		log.Printf("error executing template %s: %v", r.Template, err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// isSecure checks if request was done over HTTPS.
func isSecure(req *http.Request) bool {
	return req.Header.Get("X-Forwarded-Proto") == "https"
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
