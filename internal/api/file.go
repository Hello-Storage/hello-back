package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/gin-gonic/gin"
)

// GetFile returns file details as JSON.
//
// GET /api/file/info/:uid
// Params:
// - uid
func GetFile(router *gin.RouterGroup) {

	router.GET("/info/:uid", func(c *gin.Context) {
		// To Do check access grant
		uid := c.Param("uid")

		p, err := query.FindFileByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}

// GetShareState returns file share state based on file id.
//
// GET /share/state
// Params:
// - file_uid
func GetShareState(router *gin.RouterGroup) {
	router.GET("/share/state", func(c *gin.Context) {
		//get file id from params
		file_uid := c.Query("file_uid")
		fileMutex := sync.Mutex{}

		fileMutex.Lock()
		defer fileMutex.Unlock()

		//check if file exists
		f, err := query.FindFileByUID(file_uid)
		if err != nil {
			log.Errorf("cannot get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//get share state, if doesn't exist, create it
		share_state, err := query.FindShareStateByFileUID(file_uid)
		if err != nil {
			log.Errorf("Error finding share state: %s", err)
			share_state, err = query.CreateShareState(f)
			if err != nil {
				log.Errorf("cannot create share state: %s", err)
				AbortEntityNotFound(c)
				return
			}
		}

		c.JSON(http.StatusOK, share_state)
	})

	router.GET("/share/states", func(c *gin.Context) {
		// Get file UIDs from query params
		fileUIDs := c.QueryArray("file_uids")

		// Print for debugging
		fmt.Println("Received file UIDs:", fileUIDs)

		if len(fileUIDs) == 0 {
			fmt.Println("No file UIDs received")
			AbortEntityNotFound(c)
		}
		var shareStates []entity.FileShareState

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		// Iterate through file UIDs
		for _, fileUID := range fileUIDs {
			// check if file exists
			f, err := query.FindFileByUID(fileUID)
			if err != nil {
				log.Errorf("cannot get file: %s", err)
				// skip this file if it doesn't exist
				continue
			}

			// get share state, if doesn't exist, create it
			shareState, err := query.FindShareStateByFileUID(fileUID)
			if err != nil {
				log.Errorf("Error finding share state: %s", err)
				shareState, err = query.CreateShareState(f)
				if err != nil {
					log.Errorf("cannot create share state: %s", err)
					// skip this file if share state creation fails
					continue
				}
			}

			shareStates = append(shareStates, shareState)
		}

		c.JSON(http.StatusOK, shareStates)
	})
}

// PublishFile publishes a file.
//
// POST /share/publish
// Params:
// - selectedShareFile: form.CustomFileMeta
func PublishFile(router *gin.RouterGroup) {
	router.POST("/share/publish", func(c *gin.Context) {
		//get file id from params
		var selectedShareFile form.CustomFileMeta
		err := c.BindJSON(&selectedShareFile)
		if err != nil {
			log.Errorf("cannot bind json: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//check if file exists
		f, err := query.FindFileByUID(selectedShareFile.UID)
		if err != nil {
			log.Errorf("cannot get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//get share state, if doesn't exist, create it
		share_state, err := query.FindShareStateByFileUID(selectedShareFile.UID)
		if err != nil {
			share_state, err = query.CreateShareState(f)
			if err != nil {
				log.Errorf("cannot create share state: %s", err)
				AbortEntityNotFound(c)
				return
			}
		}

		//update share state

		public_file, err := query.PublishFile(share_state, selectedShareFile)
		if err != nil {
			log.Errorf("cannot update share state: %s", err)
			AbortEntityNotFound(c)
			return
		}

		share_state.PublicFile = *public_file

		share_state.UpdatedAt = time.Now()

		err = share_state.Save()

		if err != nil {
			log.Errorf("cannot update share state: %s", err)
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, share_state)
	})

	router.POST("/share/group", func(c *gin.Context) {
		//get files
		var selectedShareFile form.CustomFileMeta
		err := c.BindJSON(&selectedShareFile)
		if err != nil {
			log.Errorf("cannot bind json: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//check if file exists
		f, err := query.FindFileByUID(selectedShareFile.UID)
		if err != nil {
			log.Errorf("cannot get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//get share state, if doesn't exist, create it
		share_state, err := query.FindShareStateByFileUID(selectedShareFile.UID)
		if err != nil {
			share_state, err = query.CreateShareState(f)
			if err != nil {
				log.Errorf("cannot create share state: %s", err)
				AbortEntityNotFound(c)
				return
			}
		}

		//update share state

		public_file, err := query.PublishFile(share_state, selectedShareFile)
		if err != nil {
			log.Errorf("cannot update share state: %s", err)
			AbortEntityNotFound(c)
			return
		}

		share_state.PublicFile = *public_file

		share_state.UpdatedAt = time.Now()

		err = share_state.Save()

		if err != nil {
			log.Errorf("cannot update share state: %s", err)
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, share_state)
	})
}

// UnpublishFile unpublishes a file.
//
// POST /share/unpublish
// Params:
// - selectedShareFile: form.CustomFileMeta
func UnpublishFile(router *gin.RouterGroup) {
	router.POST("/share/unpublish", func(c *gin.Context) {
		//get file id from params
		var selectedShareFile form.CustomFileMeta
		err := c.BindJSON(&selectedShareFile)
		if err != nil {
			log.Errorf("cannot bind json: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//check if file exists
		f, err := query.FindFileByUID(selectedShareFile.UID)
		if err != nil {
			log.Errorf("cannot get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//get share state, if doesn't exist, create it
		share_state, err := query.FindShareStateByFileUID(selectedShareFile.UID)
		if err != nil {
			share_state, err = query.CreateShareState(f)
			if err != nil {
				log.Errorf("cannot create share state: %s", err)
				AbortEntityNotFound(c)
				return
			}
		}

		//update share state
		//delete public file from db
		//update share state
		if share_state.PublicFile.ID != 0 { // Check if the PublicFile has a valid ID
			err = share_state.PublicFile.Delete()
			if err != nil {
				log.Errorf("cannot delete public file: %s", err)
				AbortEntityNotFound(c)
				return
			}
		} else {
			log.Info("No existing public file to delete.")
		}

		share_state.PublicFile = entity.PublicFile{}

		share_state.UpdatedAt = time.Now()

		err = share_state.Save()

		if err != nil {
			log.Errorf("cannot update share state: %s", err)
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, share_state)
	})
}

// GetPublishedFile publishes a file.
//
// GET /share/published/:hash
// Params:
// - hash: form.PublicFile.ShareHash
func GetPublishedFile(router *gin.RouterGroup) {
	router.GET("/share/published/:hash", func(c *gin.Context) {
		log.Print("GetPublishedFile")
		//get file id from params
		hash := c.Param("hash")

		log.Print(hash)
		//get public file
		public_file, err := query.FindPublicFileByHash(hash)
		if err != nil {
			log.Errorf("cannot get public file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, public_file)
	})
}

// GetPublishedFileName returns a published file name.
//
// GET /share/published/name/:hash
// Params:
// - hash: form.PublicFile.ShareHash
func GetPublishedFileName(router *gin.RouterGroup) {
	router.GET("/share/published/name/:hash", func(c *gin.Context) {
		//get file id from params
		hash := c.Param("hash")

		//get public file
		public_file, err := query.FindPublicFileByHash(hash)
		if err != nil {
			log.Errorf("cannot get public file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, public_file.Name)
	})
}
