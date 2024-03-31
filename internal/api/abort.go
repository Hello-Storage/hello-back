package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Err     string `json:"error"`
	Msg     string `json:"message"`
	Details string `json:"details"`
}

func Abort(c *gin.Context, code int, msg string, params ...interface{}) {
	resp := Response{
		Code: code,
		Msg:  fmt.Sprintf(msg, params...),
	}

	log.Debugf("api: abort %s with code %d (%s)", c.FullPath(), code, strings.ToLower(resp.Msg))

	c.AbortWithStatusJSON(code, resp)
}

// AbortUnauthorized aborts with status code 401.
func AbortUnauthorized(c *gin.Context) {
	Abort(c, http.StatusUnauthorized, "Please log in to your account")
}

// AbortForbidden aborts with status code 403.
func AbortForbidden(c *gin.Context) {
	Abort(c, http.StatusForbidden, "Permission denied")
}

// AbortInternalServerError aborts with status code 500.
func AbortInternalServerError(c *gin.Context) {
	Abort(c, http.StatusForbidden, "internal server error")
}

// AbortNotFound aborts with status code 404.
func AbortNotFound(c *gin.Context) {
	Abort(c, http.StatusNotFound, "Not found")
}

// AbortEntityNotFound aborts with status code 404.
func AbortEntityNotFound(c *gin.Context) {
	Abort(c, http.StatusNotFound, "Entity not found")
}

func AbortSaveFailed(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, "Changes could not be saved")
}

func AbortDeleteFailed(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, "Could not be deleted")
}

func AbortUnexpected(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, "Something went wrong, try again")
}

func AbortBadRequest(c *gin.Context) {
	Abort(c, http.StatusBadRequest, "Unable to do that")
}

func AbortFeatureDisabled(c *gin.Context) {
	Abort(c, http.StatusForbidden, "Feature disabled")
}

func AbortBusy(c *gin.Context) {
	Abort(c, http.StatusTooManyRequests, "Busy, please try again later")
}

func ErrorResponse(err error, code ...string) gin.H {
    response := gin.H{"error": err.Error(), "status": "fail"}

    if len(code) > 0 {
        response["code"] = code
    }

    return response
}
