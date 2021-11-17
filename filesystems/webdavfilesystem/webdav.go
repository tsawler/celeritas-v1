package webdavfilesystem

import (
	"fmt"
	"github.com/studio-b12/gowebdav"
	"github.com/tsawler/celeritas/filesystems"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type WebDAV struct {
	User string
	Pass string
	Host string
}

func (w *WebDAV) getCredentials() *gowebdav.Client {
	c := gowebdav.NewClient(w.Host, w.User, w.Pass)
	return c
}

func (w *WebDAV) Put(fileName, folder string) error {
	client := w.getCredentials()

	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	err = client.WriteStream(path.Base(fileName), file, 0644)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (w *WebDAV) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing

	client := w.getCredentials()
	files, _ := client.ReadDir(prefix)
	for _, key := range files {
		if !strings.HasPrefix(key.Name(), ".") {
			b := float64(key.Size())
			kb := b / 1024
			mb := kb / 1024
			current := filesystems.Listing{
				LastModified: key.ModTime(),
				Key:          key.Name(),
				Size:         mb,
				IsDir:        key.IsDir(),
			}
			listing = append(listing, current)
		}
	}

	return listing, nil
}

func (w *WebDAV) Delete(itemsToDelete []string) bool {
	client := w.getCredentials()
	for _, item := range itemsToDelete {
		err := client.Remove(item)
		if err != nil {
			return false
		}
	}
	return true
}

func (w *WebDAV) Get(destination string, items ...string) error {
	client := w.getCredentials()

	for _, item := range items {
		err := func() error {
			webdavFilePath := item
			localFilePath := fmt.Sprintf("%s/%s", destination, path.Base(item))

			reader, err := client.ReadStream(webdavFilePath)
			if err != nil {
				return err
			}

			file, _ := os.Create(localFilePath)
			defer file.Close()

			_, err = io.Copy(file, reader)
			if err != nil {
				return err
			}
			return nil
		}()

		if err != nil {
			return err
		}
	}
	return nil
}
