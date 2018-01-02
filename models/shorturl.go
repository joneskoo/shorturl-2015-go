package models

import (
	"errors"
	"net/url"
	"time"
)

const idBase = 36

// Errors
var (
	ErrNotFound = errors.New("Shorturl not found")
)

// Shorturl database structure
type Shorturl struct {
	UID   string
	URL   string
	Host  string
	Added time.Time
}

// TargetDomain is the shorturl target domain name
func (s *Shorturl) TargetDomain() string {
	url, err := url.Parse(s.URL)
	if err != nil {
		return ""
	}
	return url.Host
}
