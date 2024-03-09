package s3

import (
	"bytes"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

//max part size of 5 KB
var maxPartSize = int64(5 * 1024 * 1024) // 5MB 
var maxRetries = 3
var awsAccessKeyID = config.Env().WasabiAccessKey
var awsSecretAccessKey = config.Env().WasabiSecretKey
var awsBucketRegion = aws.String(config.Env().WasabiRegion)
var awsBucketName = config.Env().WasabiBucket

func UploadMultipart(
	file *multipart.FileHeader,
	cid string,
) {
	creds := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, "")
	_, err := creds.Get()
	if err != nil {
		log.Printf("bad credentials: %s", err)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      awsBucketRegion,
	}))

	if err != nil {
		log.Printf("failed to create session: %s", err)
		return
	}

	svc := s3.New(sess)

	src, err := file.Open()
	if err != nil {
		log.Printf("err opening file: %s", err)
		return
	}
	defer src.Close()
	size := file.Size
	buffer := make([]byte, size)
	fileType := http.DetectContentType(buffer)
	src.Read(buffer)

	path := cid
	input := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(awsBucketName),
		Key:         aws.String(path),
		ContentType: aws.String(fileType),
	}

	resp, err := svc.CreateMultipartUpload(input)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("Created multipart upload request")

	var curr, partLength int64
	var remaining = size
	var completedParts []*s3.CompletedPart
	partNumber := 1
	for curr = 0; remaining != 0; curr += partLength {
		if remaining < maxPartSize {
			partLength = remaining
		} else {
			partLength = maxPartSize
		}
		completedPart, err := uploadPart(svc, resp, buffer[curr:curr+partLength], partNumber)
		if err != nil {
			log.Println(err.Error())
			err := abortMultipartUpload(svc, resp)
			if err != nil {
				log.Println(err.Error())
			}
			return
		}
		remaining -= partLength
		partNumber++
		completedParts = append(completedParts, completedPart)
	}

	completeResponse, err := completeMultipartUpload(svc, resp, completedParts)
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Printf("Successfully uploaded file: %s\n", completeResponse.String())
}

func completeMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, completedParts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}
	return svc.CompleteMultipartUpload(completeInput)
}

func uploadPart(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, fileBytes []byte, partNumber int) (*s3.CompletedPart, error) {
	tryNum := 1
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(fileBytes),
		Bucket:        resp.Bucket,
		Key:           resp.Key,
		PartNumber:    aws.Int64(int64(partNumber)),
		UploadId:      resp.UploadId,
		ContentLength: aws.Int64(int64(len(fileBytes))),
	}

	for tryNum <= maxRetries {
		uploadResult, err := svc.UploadPart(partInput)
		if err != nil {
			if tryNum == maxRetries {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				}
				return nil, err
			}
			log.Printf("Retrying to upload part #%v\n", partNumber)
			tryNum++
		} else {
			return &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}, nil
		}
	}
	return nil, nil
}

func abortMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput) error {
	log.Println("Aborting multipart upload for UploadId#" + *resp.UploadId)
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
	}
	_, err := svc.AbortMultipartUpload(abortInput)
	return err
}
