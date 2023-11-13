package server

import (
	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/middlewares"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

var APIv1 *gin.RouterGroup
var AuthAPIv1 *gin.RouterGroup

func registerRoutes(router *gin.Engine) {
	// Enables automatic redirection if the current route cannot be matched but a
	// handler for the path with (without) the trailing slash exists.
	// router.RedirectTrailingSlash = true

	// Create API router group.
	APIv1 = router.Group("/api")
	// Create AuthAPI router group.
	tokenMaker, err := token.NewPasetoMaker(config.Env().TokenSymmetricKey)
	if err != nil {
		log.Errorf("cannot create token maker: %s", err)
		panic(err)
	}



	AuthAPIv1 := router.Group("/api")
	AuthAPIv1.Use(middlewares.AuthMiddleware(tokenMaker))
	// routes
	api.Ping(APIv1)
	api.FetchReferredUsers(APIv1)
	api.GetUserCount(APIv1)

	//statistics routes
	api.GetStatistics(APIv1)
	api.GetWeeklyPublicStats(APIv1)
	api.GetWeeklyUserStats(APIv1)

	// auth routes
	api.LoginUser(APIv1, tokenMaker)
	api.RenewAccessToken(APIv1, tokenMaker)
	api.OAuthGoogle(APIv1, tokenMaker)
	api.OAuthGithub(APIv1, tokenMaker)
	api.RequestNonce(APIv1)
	api.StartOTP(APIv1)
	api.VerifyOTP(APIv1, tokenMaker)

	// user routes
	api.LoadUser(AuthAPIv1)
	api.GetUserDetail(AuthAPIv1)

	// file routes
	FileRoutes := AuthAPIv1.Group("/file")
	api.GetFile(FileRoutes)
	api.PutUploadFiles(FileRoutes)
	api.DeleteFile(FileRoutes)
	api.DownloadFile(FileRoutes)
	api.UpdateFileRoot(FileRoutes)
	api.CheckFilesExistInPool(FileRoutes)
	api.GetShareState(FileRoutes)
	api.PublishFile(FileRoutes)
	api.UnpublishFile(FileRoutes)
	api.GetPublishedFile(FileRoutes)


	api.GetPublishedFileName(router.Group("/api/file"))

	// folder routes
	api.SearchFolderByRoot(AuthAPIv1)
	api.CreateFolder(AuthAPIv1)
	api.DownloadFolder(AuthAPIv1)
	api.DeleteFolder(AuthAPIv1)
	api.UpdateFolderRoot(AuthAPIv1)
}
