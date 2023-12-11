package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"

	"github.com/Hello-Storage/hello-back/internal/config"
	s3Local "github.com/Hello-Storage/hello-back/pkg/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

// A global variable to store the upload state
var (
	uploadStates = make(map[string]*multipartUploadState)
	statesMutex  sync.RWMutex
)

type multipartUploadState struct {
	UploadID       string
	CompletedParts []*s3.CompletedPart
	PartNumber     int
}

func UploadFileMultipart(router *gin.RouterGroup) {
	router.POST("/upload/multipart", func(c *gin.Context) {
		//parse multipart form
		err := c.Request.ParseMultipartForm(10 << 20) // Max 10MB memory used
		if err != nil {
			log.Errorf("cannot parse multipart form: %s", err)
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}

		file, header, err := c.Request.FormFile("chunk")
		if err != nil {
			log.Errorf("cannot get chunk: %s", err)
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}
		defer file.Close()

		cid := c.PostForm("cid")
		offsetStr := c.PostForm("offset")
		offset, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			log.Errorf("cannot parse offset: %s", err)
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}

		statesMutex.Lock()
		defer statesMutex.Unlock()
		//process the chunks and handle multipart upload
		isFirstChunk := offset == 0
		var state *multipartUploadState
		var ok bool

		if isFirstChunk {
			// Initiate multipart upload
			state, err = initiateMultipartUpload(s3Local.GetS3Client(), cid)
			if err != nil {
				log.Errorf("cannot initiate multipart upload: %s", err)
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
				return
			}
			uploadStates[cid] = state
		} else {
			// Retrivee the existing upload state
			state, ok = uploadStates[cid]
			if !ok {
				log.Errorf("upload state not found")
				abortMultipartUploadAsync(s3Local.GetS3Client(), cid) // Abort if already started
				c.JSON(500, gin.H{
					"message": "Upload state not found",
				})
				return
			}
		}

		// Upload the chunk
		completedPart, err := uploadPart(s3Local.GetS3Client(), state, file, header, cid)
		if err != nil {
			log.Errorf("cannot upload part: %s", err)
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}
		state.CompletedParts = append(state.CompletedParts, completedPart)
		state.PartNumber++

		// Check if this is the last chunk
		isLastChunk := c.PostForm("isLastChunk") == "true"
		if isLastChunk {
			err = completeMultipartUpload(s3Local.GetS3Client(), state, cid)
			if err != nil {
				log.Errorf("cannot complete multipart upload: %s", err)
				abortMultipartUploadAsync(s3Local.GetS3Client(), cid) // Abort if already started
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
				return
			}
			log.Printf("Multipart upload completed for %s at %s", cid, state.UploadID)
			delete(uploadStates, cid)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Chunk received",
		})

	})

}

func initiateMultipartUpload(svc *s3.S3, cid string) (*multipartUploadState, error) {
	var awsBucketName = config.Env().WasabiBucket
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(awsBucketName),
		Key:    aws.String(cid),
	}

	resp, err := svc.CreateMultipartUpload(input)
	if err != nil {
		return nil, err
	}

	return &multipartUploadState{
		UploadID:       *resp.UploadId,
		CompletedParts: []*s3.CompletedPart{},
		PartNumber:     1,
	}, nil
}

func uploadPart(svc *s3.S3, state *multipartUploadState, file multipart.File, fileHeader *multipart.FileHeader, cid string) (*s3.CompletedPart, error) {
	var awsBucketName = config.Env().WasabiBucket
	log.Println("starting upload part")
	buffer := make([]byte, fileHeader.Size)
	_, err := file.Read(buffer)
	if err != nil {
		return nil, err
	}
	log.Println("readed file")

	partInput := &s3.UploadPartInput{
		Body:       bytes.NewReader(buffer),
		Bucket:     aws.String(awsBucketName),
		Key:        aws.String(cid), // cid should be passed or derived
		PartNumber: aws.Int64(int64(state.PartNumber)),
		UploadId:   aws.String(state.UploadID),
	}
	log.Println("uploading part")

	// Upload part in a separate goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	var uploadResult *s3.UploadPartOutput
	go func() {
		defer wg.Done()
		uploadResult, err = svc.UploadPart(partInput)
		if err != nil {
			log.Printf("Error uploading part: %d %s", state.PartNumber, err)
		}
	}()
	wg.Wait()
	if err != nil {
		return nil, err
	}
	log.Println("uploaded part")

	return &s3.CompletedPart{
		ETag:       uploadResult.ETag,
		PartNumber: aws.Int64(int64(state.PartNumber)),
	}, nil
}

func completeMultipartUpload(svc *s3.S3, state *multipartUploadState, cid string) error {
	var awsBucketName = config.Env().WasabiBucket
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(awsBucketName),
		Key:      aws.String(cid), // cid should be passed or derived
		UploadId: aws.String(state.UploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: state.CompletedParts,
		},
	}

	_, err := svc.CompleteMultipartUpload(completeInput)
	return err
}

func abortMultipartUploadAsync(svc *s3.S3, cid string) {
	go func() {
		log.Println("Aborting multipart upload for cid:", cid)

		// Lock the state mutex to safely read the upload state
		statesMutex.Lock()
		state, exists := uploadStates[cid]
		statesMutex.Unlock()

		if !exists {
			log.Printf("No upload state found for cid: %s", cid)
			return
		}

		// Prepare the input for the multipart upload abort
		abortInput := &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(config.Env().WasabiBucket),
			Key:      aws.String(cid),
			UploadId: aws.String(state.UploadID),
		}

		// Call AbortMultipartUpload operation
		_, err := svc.AbortMultipartUpload(abortInput)
		if err != nil {
			log.Printf("Error aborting multipart upload for cid: %s, %s", cid, err.Error())
			return
		}

		log.Printf("Successfully aborted multipart upload for cid: %s", cid)

		// Clean up the state after aborting
		statesMutex.Lock()
		delete(uploadStates, cid)
		statesMutex.Unlock()
	}()
}
