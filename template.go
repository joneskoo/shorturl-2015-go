package shorturl

import (
	"html/template"
	"io"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/joneskoo/shorturl-go/assets"
)

func init() {
	err := parseHTMLTemplates([][]string{
		{"error.html", "layout.html"},
		{"index.html", "layout.html"},
		{"404.html", "layout.html"},
		{"preview.html", "layout.html"},
	})
	if err != nil {
		log.Fatalf("Parsing HTML templates: %v", err)
	}
}

var templates = map[string]interface {
	Execute(io.Writer, interface{}) error
}{}

var htmlTemplateFuncs = template.FuncMap{
	"truncate":   truncate,
	"formattime": formatTime,
	"upper":      strings.ToUpper,
}

func parseHTMLTemplates(sets [][]string) error {
	for _, set := range sets {
		templateName := set[0]
		t := template.New(templateName).Funcs(htmlTemplateFuncs)
		for _, assetName := range set {
			asset, err := assets.Asset("templates/" + assetName)
			if err != nil {
				return err
			}
			if _, err := t.Parse(string(asset)); err != nil {
				return err
			}
		}
		templates[templateName] = t
	}
	return nil
}

// truncate limits the string to 25 unicode characters
func truncate(str string, limit int) string {
	if utf8.RuneCountInString(str) <= limit {
		return str
	}
	n := 0
	reader := strings.NewReader(str)
	// calculate number of bytes for limit-1 runes
	for i := 0; i < limit-1; i++ {
		_, size, err := reader.ReadRune()
		if err != nil {
			break // unexpected end of string
		}
		n += size
	}
	return str[:n] + "…"
}

func formatTime(t *time.Time, format string) string {
	return t.Format(format)
}
