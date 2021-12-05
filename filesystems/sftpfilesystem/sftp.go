package sftpfilesystem

import (
	"fmt"
	"github.com/pkg/sftp"
	"github.com/tsawler/celeritas/filesystems"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type SFTP struct {
	Host string
	User string
	Pass string
	Port string
}

func (s *SFTP) Put(fileName, folder string) error {
	client, err := s.getCredentials()
	if err != nil {
		log.Println("Error getting credentials")
		return err
	}
	defer client.Close()

	f, err := os.Open(fileName)
	if err != nil {
		log.Println("Error opening file", fileName)
		return err
	}
	defer f.Close()

	f2, err := client.Create(fmt.Sprintf("%s/%s", folder, path.Base(fileName)))
	if err != nil {
		log.Println("Error!", err)
		return err
	}
	defer f2.Close()

	if _, err = io.Copy(f2, f); err != nil {
		log.Println("Error copying file to server")
		return err
	}
	return nil
}

func (s *SFTP) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing
	client, err := s.getCredentials()
	defer client.Close()

	if err != nil {
		return listing, err
	}

	files, err := client.ReadDir(prefix)
	if err != nil {
		return listing, err
	}

	for _, x := range files {
		var item filesystems.Listing

		if !strings.HasPrefix(x.Name(), ".") {
			b := float64(x.Size())
			kb := b / 1024
			mb := kb / 1024
			item.Key = x.Name()
			item.Size = mb
			item.LastModified = x.ModTime()
			item.IsDir = x.IsDir()
			listing = append(listing, item)
		}
	}

	return listing, nil
}

func (s *SFTP) Delete(itemsToDelete []string) bool {
	client, err := s.getCredentials()
	if err != nil {
		return false
	}
	defer client.Close()

	for _, x := range itemsToDelete {
		deleteErr := client.Remove(x)
		if deleteErr != nil {
			return false
		}
	}

	return true
}

func (s *SFTP) Get(destination string, items ...string) error {
	client, err := s.getCredentials()
	if err != nil {
		return err
	}
	defer client.Close()

	for _, item := range items {
		// create destination file
		dstFile, err := os.Create(fmt.Sprintf("%s/%s", destination, path.Base(item)))
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// open source file
		srcFile, err := client.Open(item)
		if err != nil {
			return err
		}

		// copy source file to destination file
		bytes, err := io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
		fmt.Printf("%d bytes copied\n", bytes)

		// flush in-memory copy
		err = dstFile.Sync()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SFTP) getCredentials() (*sftp.Client, error) {
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}
	cwd, err := client.Getwd()
	log.Println("Current working directory:", cwd)

	return client, nil
}

func (s *SFTP) CreateDir(p string) error {
	client, err := s.getCredentials()
	if err != nil {
		return err
	}
	defer client.Close()

	return client.Mkdir(p)
}
