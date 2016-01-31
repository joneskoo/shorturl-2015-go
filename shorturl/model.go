package shorturl

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// Shorturl database structure
type Shorturl struct {
	ID        int64
	URL, Host string
	Added     time.Time
}

// UID is the base-36 string representation of ID
func (s *Shorturl) UID() string {
	return strconv.FormatInt(s.ID, idBase)
}

// URLString is the shortened URL as string
func (s *Shorturl) URLString() string {
	return "http://" + Domain + "/" + s.UID()
}

// Represent Short URL in pretty format
func (s Shorturl) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s\n", s.URLString())
	fmt.Fprintf(&buf, "  Target: %s\n", truncate(s.URL, 64))
	fmt.Fprintf(&buf, "   Added: %s\n", s.Added.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(&buf, "      IP: %s", s.Host)
	return buf.String()
}

func truncate(str string, n int) string {
	if len(str) > n {
		return str[:n-1] + "…"
	}
	return str
}
