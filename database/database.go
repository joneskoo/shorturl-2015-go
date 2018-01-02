package database

import (
	"errors"

	"github.com/joneskoo/shorturl-go/models"
)

type Database interface {
	Get(shortCode string) (*models.Shorturl, error)
	List() (<-chan models.Shorturl, error)
	Add(url, host, clientid string) (s models.Shorturl, err error)
}

var (
	NotFound = errors.New("does not exist in database")
)
