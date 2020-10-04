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
	mux.Handle("/", serveHome(db, secure))
	mux.Handle("/p/", http.StripPrefix("/p", servePreview(db, secure)))
	mux.Handle("/static/style.css", serveStatic("css/style.css"))
	return mux
}

type errorResponse struct {
	ErrorTitle   string
	ErrorMessage string
	StatusCode   int
	Template     string
}

var errorEOL = errorResponse{
	StatusCode:   http.StatusGone,
	ErrorTitle:   "Service is end-of-life",
	ErrorMessage: "This short url service is end of life. Existing redirects continue to work for now.",
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

func serverError(e errorResponse, secure bool) http.Handler {
	return executeTemplate("error.html", e.StatusCode, e, secure)
}

func serveHome(db *DB, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/":
			serverError(errorEOL, secure).ServeHTTP(w, req)
		default:
			serveRedirect(db, secure).ServeHTTP(w, req)
		}
	})
}

func serveRedirect(db *DB, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if isAlwaysPreview(req) && !isLocalReferer(req) {
			servePreview(db, secure).ServeHTTP(w, req)
			return
		}
		shortCode := req.URL.Path[1:]
		s, err := db.Get(shortCode)
		switch err {
		case ErrNotFound:
			serverError(errorNotFound, secure).ServeHTTP(w, req)
		case nil:
			http.Redirect(w, req, s.URL, http.StatusFound)
		default:
			serverError(internalError, secure).ServeHTTP(w, req)
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
func servePreview(db *DB, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s, err := db.Get(req.URL.Path[1:])
		switch err {
		case ErrNotFound:
			serverError(errorNotFound, secure).ServeHTTP(w, req)
		case nil:
			executeTemplate("preview.html", http.StatusOK, s, secure).ServeHTTP(w, req)
		default:
			serverError(internalError, secure).ServeHTTP(w, req)
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

func executeTemplate(name string, status int, data interface{}, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		template, ok := templates[name]
		if !ok {
			log.Printf("template %s not found", name)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(status)
		protocol := "http://"
		if secure {
			protocol = "https://"
		}

		err := template.Execute(w, struct {
			Protocol string
			Domain   string
			Data     interface{}
		}{protocol, Domain, data})
		if err != nil {
			log.Printf("error executing template %s: %v", name, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
