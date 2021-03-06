package shorturl

import (
	"errors"
	"net/url"
	"strconv"
	"time"
)

// Service configuration
var (
	idBase = 36
)

// Errors
var (
	ErrNotFound = errors.New("Shorturl not found")
)

// Shorturl database structure
type Shorturl struct {
	ID    int64
	URL   string
	Host  string
	Added time.Time
}

// UID is the base-36 string representation of ID
func (s *Shorturl) UID() string {
	return strconv.FormatInt(s.ID, idBase)
}

// TargetDomain is the shorturl target domain name
func (s *Shorturl) TargetDomain() string {
	url, err := url.Parse(s.URL)
	if err != nil {
		return ""
	}
	return url.Host
}

// PreviewURL is the view that shows where URL directs
func (s *Shorturl) PreviewURL() string {
	return "/p/" + s.UID()
}
