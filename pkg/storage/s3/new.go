package s3storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func New(bucketname, prefix string) (*S3Storage, error) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("config load error: %s", err)
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	return &S3Storage{
		bucketname: bucketname,
		prefix:     prefix,
		s3Client:   s3Client,
	}, nil
}
