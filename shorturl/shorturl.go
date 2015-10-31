package main

import (
    "fmt"
    "errors"
    "time"
    "strconv"
    "log"
    )

const Domain = "yx.fi"
const IdBase = 36
var memory = make(map[int64] Shorturl)

type Shorturl struct {
    Id int64
    URL string
    Added time.Time
}

func (s *Shorturl) Save() (err error) {
    // FIXME: store shorturl to database
    s.Id = int64(len(memory)) + 1
    log.Print("Added shorturl ", s)
    memory[s.Id] = *s
    return nil
}

func GetShorturl(uid string) (s Shorturl, err error) {
    s = Shorturl{}
    id, err := strconv.ParseInt(uid, IdBase, 64)
    if err != nil {
        return s, err
    }
    // Fetch shorturl from in-memory database
    // FIXME: Fetch shorturl from actual database
    s, ok := memory[id]
    if !ok {
        err = errors.New("Shorturl not found")
    }
    return s, err
}

func (s Shorturl) Uid() string {
    return strconv.FormatInt(s.Id, IdBase)
}

func (s Shorturl) String() string {
    return fmt.Sprintf("http://%s/%s", Domain, s.Uid())
}
