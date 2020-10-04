package shorturl

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/joneskoo/shorturl-go/assets"
)

type handler struct {
	db     *Database
	secure bool
	*http.ServeMux
}

func NewHandlers(db *Database, secure bool) http.Handler {
	mux := http.NewServeMux()
	h := handler{db, secure, mux}
	mux.HandleFunc("/", h.serveHome)
	mux.Handle("/p/", http.StripPrefix("/p", http.HandlerFunc(h.servePreview)))
	mux.HandleFunc("/static/style.css", h.serveCSS)
	return &h
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

func (h handler) serverError(w http.ResponseWriter, e errorResponse) {
	h.executeTemplate(w, "error.html", e.StatusCode, nil, e)
}

func (h handler) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.serveRedirect(w, r)
		return
	}

	h.serverError(w, errorEOL)
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
		http.Redirect(w, r, s.URL, http.StatusFound)
		return
	case ErrNotFound:
		h.serverError(w, errorNotFound)
	default:
		h.serverError(w, internalError)
	}
}

func isLocalReferer(req *http.Request) bool {
	url, err := url.Parse(req.Referer())
	if err != nil {
		return false
	}
	return strings.EqualFold(url.Host, Domain)
}

// Preview shows short url details after adding
func (h handler) servePreview(w http.ResponseWriter, r *http.Request) {
	s, err := h.db.Get(r.URL.Path[1:])
	switch err {
	case nil:
		h.executeTemplate(w, "preview.html", http.StatusOK, nil, s)
	case ErrNotFound:
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

func (h handler) serveCSS(w http.ResponseWriter, r *http.Request) {
	style := assets.MustAsset("css/style.css")
	http.ServeContent(w, r, "style.css", assets.LastModified, bytes.NewReader(style))
}

func (h handler) executeTemplate(w http.ResponseWriter, name string, status int, header http.Header, data interface{}) {
	template, ok := templates[name]
	if !ok {
		log.Printf("template %s not found", name)
		http.Error(w, "", http.StatusInternalServerError)
		return
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
	}{protocol, Domain, data})
	if err != nil {
		log.Printf("error executing template %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
