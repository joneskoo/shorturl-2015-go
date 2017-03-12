package assets

import "time"

//go:generate go-bindata -pkg assets -o assets.go css templates

var LastModified = time.Now()
