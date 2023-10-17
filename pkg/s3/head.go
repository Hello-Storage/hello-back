package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// gets head object
func HeadObject(
	s3Config aws.Config,
	bucket, key string,
) (*s3.HeadObjectOutput, error) {

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

	// set parameter for bucket name
	b := aws.String(bucket)

	// set parameter for key name
	k := aws.String(key)

	// get object
	headObject, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: b,
		Key:    k,
	})

	if err != nil {
		return nil, err
	}

	return headObject, nil
}
