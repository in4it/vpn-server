package s3storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Storage) WriteFile(name string, data []byte) error {
	_, err := s.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketname),
		Key:    aws.String(name),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("put object error: %s", err)
	}
	return nil
}

func (s *S3Storage) AppendFile(name string, data []byte) error {
	return nil
}

func (s *S3Storage) OpenFileForWriting(name string) (io.WriteCloser, error) {
	return nil, nil
}

func (s *S3Storage) OpenFileForAppending(name string) (io.WriteCloser, error) {
	return nil, nil
}
