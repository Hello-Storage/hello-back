package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// GeneratePresignedURL function will generate a presigned URL for client to upload directly to S3

func GeneratePresignedURL(s3Config aws.Config, bucket, key string) (string, error) {

	goSession, err := session.NewSessionWithOptions(session.Options{
		Config:  s3Config,
		Profile: "wasabi",
	})

	// check if the session was created correctly.
	if err != nil {
		return "", err
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	req, _ := s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(15 * 60) // 15 minutes
	if err != nil {
		return "", err
	}


	return urlStr, nil
}
