package celeritas

import (
	"errors"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/tsawler/celeritas/filesystems"
	"github.com/tsawler/celeritas/filesystems/miniofilesystem"
	"github.com/tsawler/celeritas/filesystems/s3filesystem"
	"github.com/tsawler/celeritas/filesystems/sftpfilesystem"
	"github.com/tsawler/celeritas/filesystems/webdavfilesystem"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

// UploadFile uploads fileName (full path) to folder for the specified file system fs; if
// fs is nil, then we upload to the local file system
func (c *Celeritas) UploadFile(r *http.Request, folder string, fs filesystems.FS) error {
	fileName, err := getFileUpload(r)
	if err != nil {
		return err
	}

	var fsInterface interface{} = fs

	log.Printf("type is %T", fsInterface)

	switch fsInterface.(type) {
	case *miniofilesystem.Minio:
		fileSystem := c.FileSystems["MINIO"].(miniofilesystem.Minio)
		err = fileSystem.Put(fileName, folder)
		if err != nil {
			return err
		}
	case *sftpfilesystem.SFTP:
		fileSystem := c.FileSystems["SFTP"].(sftpfilesystem.SFTP)
		err = fileSystem.Put(fileName, folder)
		if err != nil {
			return err
		}
	case *webdavfilesystem.WebDAV:
		fileSystem := c.FileSystems["WEBDAV"].(webdavfilesystem.WebDAV)
		err = fileSystem.Put(fileName, folder)
		if err != nil {
			return err
		}
	case *s3filesystem.S3:
		fileSystem := c.FileSystems["S3"].(s3filesystem.S3)
		err = fileSystem.Put(fileName, folder)
		if err != nil {
			return err
		}
	default:
		// local file system; just move the file to path
		err = os.Rename(fileName, fmt.Sprintf("./%s/%s", folder, path.Base(fileName)))
		if err != nil {
			return err
		}
	}

	return nil
}

// getFileUpload gets an uploaded file from the request and stores it in the tmp folder
func getFileUpload(r *http.Request) (string, error) {
	var maxUploadSize int64

	if max, err := strconv.Atoi(os.Getenv("MAX_UPLOAD_SIZE")); err != nil {
		maxUploadSize = 10 << 20
	} else {
		maxUploadSize = int64(max)
	}

	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		return "", err
	}

	file, header, err := r.FormFile("formFile")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// check mime type against permitted types
	mimeType, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}

	// have to move back to start of file
	_, err = file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	log.Println("Mime type is", mimeType.String())

	validMimeTypes := []string{
		"image/gif",
		"image/jpeg",
		"image/png",
		"application/pdf",
	}

	if !inSlice(validMimeTypes, mimeType.String()) {
		return "", errors.New("invalid mimetype")
	}

	dst, err := os.Create(fmt.Sprintf("./tmp/%s", header.Filename))
	if err != nil {
		return "", err
	}

	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("./tmp/%s", header.Filename), nil
}

func inSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
