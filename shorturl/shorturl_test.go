package shorturl

import (
    "testing"
    nurl "net/url"
    )

var testurls = []struct {
        url string
}{
    {"http://www.example.com/"},
    {"http://www.example.com/abcd"},
    {"http://www.example.com/abcde"},
}

var idtests = []struct {
    in Shorturl
    want string
}{
    {Shorturl{Id:5}, "http://yx.fi/5"},
    {Shorturl{Id:10}, "http://yx.fi/a"},
    {Shorturl{Id:1270}, "http://yx.fi/za"},
}

func TestStringRepresentationFormat (t *testing.T) {
    for _, c := range idtests {
        in_string := c.in.String()
        if in_string != c.want {
            t.Errorf("String representation %s != %s", in_string, c.want)
        }
    }
}

func TestCanAddAndGetShorturl (t *testing.T) {
    for _, c := range testurls {
        added := Shorturl{URL: c.url}
        err := added.Save()
        if err != nil {
            t.Errorf("Saving shorturl failed: ", err)
        }
        shortened_url, err := nurl.Parse(added.String())
        if err != nil {
            t.Errorf("parsing shortened URL failed: %s", err)
            continue
        }
        retrieved, err := GetShorturl(shortened_url.Path[1:])
        if err != nil {
            t.Error("Failed to retrieve shorturl (uid=%s): %s", added, err)
            continue
        }
        if retrieved.Id != added.Id {
            t.Errorf("Received unexpected id %s, wanted %s)",
                     retrieved.Id,
                     added.Id)
        }
    }
}

func TestDifferentUriDifferentShort (t *testing.T) {
    m := make(map[string]bool)
    for _, c := range testurls {
        added := Shorturl{URL: c.url}
        err := added.Save()
        if err != nil {
            t.Errorf("Saving shorturl failed: ", err)
        }
        _, ok := m[added.String()]
        if ok {
            t.Errorf("short url %s(%s) returned multiple times for different URLs",
                     added, c.url)
            continue
        }
        m[added.String()] = true
    }
}
