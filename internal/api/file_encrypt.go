package api

import (
	"net/http"

	//"github.com/Hello-Storage/hello-back/internal/constant"
	//"github.com/Hello-Storage/hello-back/internal/db"
	//"github.com/Hello-Storage/hello-back/internal/form"
	//"github.com/Hello-Storage/hello-back/pkg/token"
	encryptionutils "github.com/Hello-Storage/hello-back/pkg/encryption"
	"github.com/gin-gonic/gin"
)

// EncryptFile encrypts a file with a given key and returns the encrypted file attributes
//
// POST /api/file/encrypt
// - file
// - personalSignature (key)
// - isFolder (bool)
// - encryptedPathMapping is a map of encrypted file path to original file path
func EncryptFile(router *gin.RouterGroup) {

	router.POST("/encrypt", func(ctx *gin.Context) {
		log.Print("entered encrypt")
		//tx := db.Db().Begin()
		//authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		//Get root from post
		personalSignature := ctx.PostForm("personalSignature")
		isFolder := ctx.PostForm("isFolder") == "true"
		root := ctx.PostForm("root")
		//if isFolder, get webkitRelativePath
		var webkitRelativePath string
		if isFolder {
			webkitRelativePath = ctx.PostForm("webkitRelativePath")
			log.Printf("webkitRelativePath: %s", webkitRelativePath)
		}

		// Parse JSON field
		//var encryptedPathsMapping map[string]string

		log.Printf("Root: %s", root)
		log.Printf("isFolder: %v", isFolder)

		// Handle file
		form, err := ctx.MultipartForm()
		if err != nil {
			log.Errorf("cannot get multipart form: %s", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid multipart form"})
			return
		}

		fileHeaders := form.File["file"]
		if len(fileHeaders) == 0 {
			log.Error("no file uploaded")
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "No file uploaded"})
			return
		}
		file := fileHeaders[0]
		//log file path

		//read file size

		//encrypt file metadata
		encryptedFilename, encryptedFiletype, err := encryptionutils.EncryptMetadata(file, personalSignature)
		if err != nil {
			log.Errorf("cannot encrypt file metadata: %s", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Cannot encrypt file metadata"})
			return
		}

		// encrypt the fileBuffer
		cidOriginalStr,
		cidOfEncryptedBufferStr,
		encryptionTime,
		encryptedFileBuffer,
		err := encryptionutils.EncryptFileBuffer(file)

		if err != nil {
			log.Errorf("cannot encrypt file buffer: %s", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Cannot encrypt file buffer"})
			return
		}

		//transform cidOfEncryptedBufferStr to []uint8 or []byte
		cidOfEncryptedBufferBytes := []uint8(cidOfEncryptedBufferStr)

		cidOriginalEncryptedBuffer, err := encryptionutils.EncryptBuffer(personalSignature, cidOfEncryptedBufferBytes)
		if err != nil {
			log.Errorf("cannot encrypt cidOfEncryptedBufferStr: %s", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Cannot encrypt cidOfEncryptedBufferStr"})
			return
		}

		//transform cidOriginalEncryptedBuffer to base64 url
		cidOriginalEncryptedBase64Url := encryptionutils.BufferToBase64Url(cidOriginalEncryptedBuffer)

		//print first 20 characters of encryptedFilename
		log.Printf("Encrypted filename: %s", encryptedFilename)
		log.Printf("Encrypted filetype hex: %s", encryptedFiletype)

		log.Printf("cidOriginalStr: %s", cidOriginalStr)
		log.Printf("cidOfEncryptedBufferStr: %s", cidOfEncryptedBufferStr)
		log.Printf("cidOriginalEncryptedBase64Url: %s", cidOriginalEncryptedBase64Url)
		log.Printf("Encrypted file buffer: %s", string(encryptedFileBuffer[:10]))
		log.Printf("Encryption time: %s", encryptionTime)

		//add all of the above to the response
		ctx.JSON(http.StatusOK, gin.H{
			"encryptedFilename":             encryptedFilename,
			"encryptedFiletype":             encryptedFiletype,
			"cidOriginalStr":                cidOriginalStr,
			"cidOfEncryptedBufferStr":       cidOfEncryptedBufferStr,
			"cidOriginalEncryptedBase64Url": cidOriginalEncryptedBase64Url,
			"encryptionTime":                encryptionTime,
			"encryptedWebkitRelativePath":   webkitRelativePath,
		})
	})
}
