/*
Package s3 provides AWS s3 functions
*/

package s3

import (
	"log"
	"sync"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	svc  *s3.S3
	once sync.Once
)

func GetS3Client() *s3.S3 {
	once.Do(func() {
		var awsAccessKeyID = config.Env().StorageAccessKey
		var awsSecretAccessKey = config.Env().StorageSecretKey
		var awsBucketRegion = config.Env().StorageRegion
		var awsEndpoint = config.Env().StorageEndpoint

		creds := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, "")
		_, err := creds.Get()
		if err != nil {
			log.Fatalf("bad credentials: %s", err)
		}

		sess, err := session.NewSession(&aws.Config{
			Credentials: creds,
			Region:      aws.String(awsBucketRegion),
			Endpoint:    aws.String(awsEndpoint),
		})
		if err != nil {
			log.Fatalf("failed to create session: %s", err)
		}

		svc = s3.New(sess)
	})
	return svc
}
