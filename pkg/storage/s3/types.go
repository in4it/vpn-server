package s3storage

import "github.com/aws/aws-sdk-go-v2/service/s3"

const CONFIG_PATH = "config"

type S3Storage struct {
	bucketname string
	prefix     string
	s3Client   *s3.Client
}
