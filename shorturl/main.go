package shorturl

import (
	"log"
	"net/http"
	"os"
)

var contentRoot = "content"

var view *View

func init() {
	if len(os.Args) >= 2 {
		contentRoot = os.Args[1]
	}

	view = NewView(contentRoot)

	http.HandleFunc("/", handler)
	http.Handle("/p/", http.StripPrefix("/p", http.HandlerFunc(view.Preview)))
	http.Handle("/add/", http.HandlerFunc(view.Add))
	//http.Handle("/static/", http.FileServer(http.Dir(contentRoot)))
}

func main() {
	addr := "[::1]:8000"
	log.Print("Listening on", addr)
	http.ListenAndServe(addr, nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	// Redirect yx.fi/xxx to target
	if req.URL.Path != "/" {
		view.Redirect(w, req)
		return
	}
	// For root URI /, serve index page
	view.Index(w, req)
}
