/*
Package api provides REST API authentication and request handlers.
*/

package api

import (
	"net/http"
	"net/http/httptest"

	"github.com/Hello-Storage/hello-back/internal/event"
	"github.com/gin-gonic/gin"
)

var log = event.Log


func Ping(router *gin.RouterGroup) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "hello backend api endpoints")
	})
}

// NewApiTest returns new API test helper.
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup) {
	gin.SetMode(gin.TestMode)

	app = gin.New()
	router = app.Group("/api")

	return app, router
}

// Executes an API request with an empty request body.
// See https://medium.com/@craigchilds94/testing-gin-json-responses-1f258ce3b0b1
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}
