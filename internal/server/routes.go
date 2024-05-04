package server

import (
	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/api/arweave"
	v1 "github.com/Hello-Storage/hello-back/internal/api/v1"
	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/middlewares"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

func registerRoutes(router *gin.Engine) {
	var APIv1 *gin.RouterGroup
	var AuthAPIv1 *gin.RouterGroup
	tokenMaker, err := token.NewPasetoMaker(config.Env().TokenSymmetricKey)
	if err != nil {
		log.Errorf("cannot create token maker: %s", err)
		panic(err)
	}

	// Create router groups.
	APIv1 = router.Group("/api")
	AuthAPIv1 = router.Group("/api")
	AuthAPIv1.Use(middlewares.AuthMiddleware(tokenMaker))

	// routes
	api.Ping(APIv1)
	api.FetchReferredUsers(APIv1)
	api.GetUserCount(APIv1)

	//api keys routes
	api.ApiKey(AuthAPIv1, tokenMaker)

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
	api.CreateFile(FileRoutes)
	api.DeleteFile(FileRoutes)
	api.DownloadFile(FileRoutes)
	api.DownloadMultipartFile(FileRoutes)
	api.UpdateFileRoot(FileRoutes)
	api.CheckFilesExistInPool(FileRoutes)
	api.GetShareState(FileRoutes)
	api.PublishFile(FileRoutes)
	api.UnpublishFile(FileRoutes)
	api.GetPublishedFile(FileRoutes)
	api.EncryptFile(FileRoutes)
	api.UploadFileMultipart(FileRoutes)

	api.GetPublishedFileName(router.Group("/api/file"))

	// folder routes
	api.SearchFolderByRoot(AuthAPIv1)
	api.CreateFolder(AuthAPIv1)
	api.GetFolderFiles(AuthAPIv1)
	api.DownloadFolder(AuthAPIv1)
	api.DownloadMultipartFolder(AuthAPIv1)
	api.DeleteFolder(AuthAPIv1)
	api.UpdateFolderRoot(AuthAPIv1)

	arweave.GetArweaveTransactions(APIv1.Group("/arweave"))

}

func RegisterApiRoutes(router *gin.Engine) {
	var ApiKeyAPIv1 *gin.RouterGroup

	tokenMaker, err := token.NewPasetoMaker(config.Env().TokenSymmetricKey)
	if err != nil {
		log.Errorf("cannot create token maker: %s", err)
		panic(err)
	}

	// Create router groups.

	//Public route without apiKey
	ApiPublic := router.Group("/public-api/v1")
	//Public api route
	v1.InvestPostData(ApiPublic)
	v1.InvestGetDataByCode(ApiPublic)

	//Public route with apiKey
	ApiKeyAPIv1 = router.Group("/public-api/v1")
	ApiKeyAPIv1.Use(middlewares.APIKeyAuthMiddleware(tokenMaker))

	//api routes
	v1.Ping(ApiKeyAPIv1)
	v1.FileCreate(ApiKeyAPIv1)
	v1.GetFile(ApiKeyAPIv1)
	v1.FileUpdate(ApiKeyAPIv1)
	v1.DeleteFile(ApiKeyAPIv1)
	v1.DownloadFile(ApiKeyAPIv1)

}
