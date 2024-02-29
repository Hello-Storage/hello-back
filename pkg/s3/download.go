package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func DownloadObject(s3Config aws.Config, bucket, key string) (*s3.GetObjectOutput, error) {
	// create a new session using the config above and profile
	goSession, err := session.NewSessionWithOptions(session.Options{
		Config:  s3Config,
		Profile: "wasabi",
	})

	// check if the session was created correctly.
	if err != nil {
		return nil, err
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

		

	// create put object input
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key), //object-key
	}

	// get file
	file, err := s3Client.GetObject(getObjectInput)

	// print if there is an error
	if err != nil {
		return nil, err
	}

	return file, nil
}
