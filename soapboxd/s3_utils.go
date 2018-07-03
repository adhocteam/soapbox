package soapboxd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

type s3storage struct {
	svc *s3.S3
}

func newS3Storage(sess *session.Session) *s3storage {
	return &s3storage{svc: s3.New(sess)}
}

func (s *s3storage) uploadFile(bucket string, key string, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening file %s: %v", filename, err)
	}
	defer f.Close()
	input := &s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	_, err = s.svc.PutObject(input)
	return err
}

func (s *s3storage) downloadFile(bucket string, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	object, err := s.svc.GetObject(input)
	if err != nil {
		return nil, errors.Wrap(err, "downloading file")
	}

	defer object.Body.Close()

	body, err := ioutil.ReadAll(object.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading file")
	}

	return body, nil
}

func (s *s3storage) deleteFile(bucket string, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s.svc.DeleteObject(input)
	if err != nil {
		return errors.Wrap(err, "deleting file")
	}
	return nil
}
