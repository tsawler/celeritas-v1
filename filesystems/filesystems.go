package filesystems

import "time"

type FS interface {
	CreateDir(p string) error
	Delete(itemsToDelete []string) bool
	Get(destination string, items ...string) error
	List(prefix string) ([]Listing, error)
	Put(fileName, folder string) error
}

type Listing struct {
	Etag         string
	LastModified time.Time
	Key          string
	Size         float64
	IsDir        bool
}
