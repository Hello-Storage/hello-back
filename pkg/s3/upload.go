package s3

import (
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func UploadObject(
	s3Config aws.Config,
	file *multipart.FileHeader,
	bucket, key string,
) error {

	// create a new session using the config above and profile
	goSession, err := session.NewSessionWithOptions(session.Options{
		Config:  s3Config,
		Profile: "wasabi",
	})

	// check if the session was created correctly.
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(goSession, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.LeavePartsOnError = true
	})

	// Create a progress reader that wraps the file reader
	// reader := &progressReader{
	// 	file: file,
	// 	src:  src,
	// 	size: file.Size,
	// 	key:  key,

	// }

	// Set the S3 upload input parameters
	input := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   src,
	}

	// Upload the file to S3.
	_, err = uploader.Upload(input)
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	// fmt.Printf("file uploaded to, %s\n", aws.StringValue(&result.Location))

	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	return nil
}