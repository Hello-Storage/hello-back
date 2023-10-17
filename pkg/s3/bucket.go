/*
Package s3 provides AWS s3 functions
*/

package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// creates a new Filebase bucket
func CreateBucket(s3Config aws.Config, bucket string) error {

	goSession, err := session.NewSessionWithOptions(session.Options{
		Config:  s3Config,
		Profile: "wasabi",
	})

	// check if the session was created correctly.
	if err != nil {
		return err
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	

	

	// set parameter for bucket name
	b := aws.String(bucket)

	// create a bucket
	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: b,
	})

	if err != nil {
		return err
	}

	return nil
}
