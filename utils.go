package shorturl

import (
	"strings"
	"time"
	"unicode/utf8"
)

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
