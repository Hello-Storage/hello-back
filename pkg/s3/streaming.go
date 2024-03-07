package s3

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// OpenStream opens a stream to an S3 object

type S3Service struct {
	Client *s3.S3
}

// NewS3Service creates a new instance of S3Service
func NewS3Service(accessKey, secretKey, region, endpoint string) *S3Service {
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
	}))

	return &S3Service{
		Client: s3.New(sess),
	}
}

// OpenStream opens a stream to an S3 object
func (svc *S3Service) OpenStream(bucket, key string) (io.ReadCloser, int64, string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := svc.Client.GetObject(input)
	if err != nil {
		return nil, 0, "", err
	}

	return result.Body, *result.ContentLength, *result.ContentType, nil
}
