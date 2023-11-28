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
		var awsAccessKeyID = config.Env().WasabiAccessKey
		var awsSecretAccessKey = config.Env().WasabiSecretKey
		var awsBucketRegion = config.Env().WasabiRegion
		var awsEndpoint = config.Env().WasabiEndpoint

		log.Printf("awsAccessKeyID: %s", awsAccessKeyID)
		log.Printf("awsSecretAccessKey: %s", awsSecretAccessKey)
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
