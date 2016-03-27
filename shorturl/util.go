package shorturl

import (
    "strings"
    	"unicode/utf8"

)

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
