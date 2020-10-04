package shorturl

import (
	"testing"
)

var testurls = []struct {
	url string
}{
	{"https://www.example.com/"},
	{"https://www.example.com/abcd"},
	{"https://www.example.com/abcde"},
}

var idtests = []struct {
	in   Shorturl
	want string
}{
	{Shorturl{ID: 5}, "5"},
	{Shorturl{ID: 10}, "a"},
	{Shorturl{ID: 1270}, "za"},
}

func TestURLStringFormat(t *testing.T) {
	for _, c := range idtests {
		inString := c.in.UID()
		if inString != c.want {
			t.Errorf("String representation %s != %s", inString, c.want)
		}
	}
}

// func TestCanAddAndGetShorturl (t *testing.T) {
//     for _, c := range testurls {
//         added := Shorturl{URL: c.url}
//         err := added.Save()
//         if err != nil {
//             t.Errorf("Saving shorturl failed: ", err)
//         }
//         shortened_url, err := nurl.Parse(added.String())
//         if err != nil {
//             t.Errorf("parsing shortened URL failed: %s", err)
//             continue
//         }
//         retrieved, err := GetShorturl(shortened_url.Path[1:])
//         if err != nil {
//             t.Error("Failed to retrieve shorturl (uid=%s): %s", added, err)
//             continue
//         }
//         if retrieved.Id != added.Id {
//             t.Errorf("Received unexpected id %s, wanted %s)",
//                      retrieved.Id,
//                      added.Id)
//         }
//     }
// }

// func TestDifferentUriDifferentShort (t *testing.T) {
//     m := make(map[string]bool)
//     for _, c := range testurls {
//         added := Shorturl{URL: c.url}
//         err := added.Save()
//         if err != nil {
//             t.Errorf("Saving shorturl failed: ", err)
//         }
//         _, ok := m[added.String()]
//         if ok {
//             t.Errorf("short url %s(%s) returned multiple times for different URLs",
//                      added, c.url)
//             continue
//         }
//         m[added.String()] = true
//     }
// }
