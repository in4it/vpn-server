package s3storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Storage) ReadDir(pathname string) ([]string, error) {
	objectList, err := s.s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketname),
		Prefix: aws.String(s.prefix + "/" + strings.TrimLeft(pathname, "/")),
	})
	if err != nil {
		return []string{}, fmt.Errorf("list object error: %s", err)
	}
	res := make([]string, len(objectList.Contents))
	for k, object := range objectList.Contents {
		res[k] = *object.Key
	}
	return res, nil
}
