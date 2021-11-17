package s3filesystem

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/tsawler/celeritas/filesystems"
	"log"
	"net/http"
	"os"
	"path"
)

type S3 struct {
	Key      string
	Secret   string
	Region   string
	Endpoint string
	Bucket   string
}

// getCredentials returns a credential object with key/secret
func (s *S3) getCredentials() *credentials.Credentials {
	c := credentials.NewStaticCredentials(s.Key, s.Secret, "")
	return c
}

// Put uploads an asset to the S3 bucket
func (s *S3) Put(fileName, folder string) error {
	c := s.getCredentials()
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &s.Endpoint,
		Region:      &s.Region,
		Credentials: c,
	}))

	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to open file %q, %v", fileName, err)
	}
	defer f.Close()

	fileInfo, _ := f.Stat()
	var size = fileInfo.Size()

	buffer := make([]byte, size)
	_, _ = f.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(fmt.Sprintf("%s/%s", folder, path.Base(fileName))),
		Body:        fileBytes,
		ACL:         aws.String("public-read"),
		ContentType: aws.String(fileType),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"),
		},
	})
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to upload file, %v", err)
	}

	return nil
}

// List gets a listing of files in the directory prefix from DO/AWS
func (s *S3) List(prefix string) ([]filesystems.Listing, error) {
	c := s.getCredentials()
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &s.Endpoint,
		Region:      &s.Region,
		Credentials: c,
	}))

	svc := s3.New(sess)
	input := &s3.ListObjectsInput{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(""),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, nil
	}

	var listing []filesystems.Listing

	for _, key := range result.Contents {
		b := float64(*key.Size)
		kb := b / 1024
		mb := kb / 1024
		current := filesystems.Listing{
			Etag:         *key.ETag,
			LastModified: *key.LastModified,
			Key:          *key.Key,
			Size:         mb,
		}
		listing = append(listing, current)
	}

	return listing, nil
}

// Delete deletes one more more items from DO/AWS
func (s *S3) Delete(itemsToDelete []string) bool {
	c := s.getCredentials()
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &s.Endpoint,
		Region:      &s.Region,
		Credentials: c,
	}))

	svc := s3.New(sess)

	for _, item := range itemsToDelete {
		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(s.Bucket),
			Delete: &s3.Delete{
				Objects: []*s3.ObjectIdentifier{
					{
						Key: aws.String(item),
					},
				},
				Quiet: aws.Bool(false),
			},
		}

		_, err := svc.DeleteObjects(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					fmt.Println(aerr.Error())
					return false
				}
			} else {
				fmt.Println(err.Error())
				return false
			}
		}
	}
	return true
}

func (s *S3) Get(destination string, items ...string) error {
	c := s.getCredentials()
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &s.Endpoint,
		Region:      &s.Region,
		Credentials: c,
	}))

	for _, item := range items {
		err := func() error {
			file, err := os.Create(fmt.Sprintf("%s/%s", destination, item))
			if err != nil {
				return err
			}

			defer file.Close()

			downloader := s3manager.NewDownloader(sess)
			_, err = downloader.Download(file,
				&s3.GetObjectInput{
					Bucket: aws.String(s.Bucket),
					Key:    aws.String(item),
				})
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
