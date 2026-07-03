package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/pkg/apperror"
)

// Response 通用响应结构
// swagger:model Response
type Response struct {
	Code      int         `json:"code" example:"200"`
	ErrorCode string      `json:"error_code,omitempty" example:"ROUTE_NOT_FOUND"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

const (
	CodeSuccess = 200
	CodeFail    = 400
	CodeError   = 500
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      CodeSuccess,
		Message:   "success",
		Data:      data,
		RequestID: platformlogger.RequestIDFromGin(c),
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      CodeSuccess,
		Message:   message,
		Data:      data,
		RequestID: platformlogger.RequestIDFromGin(c),
	})
}

func Fail(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:      CodeFail,
		Message:   message,
		RequestID: platformlogger.RequestIDFromGin(c),
	})
}

func FailWithCode(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:      code,
		Message:   message,
		RequestID: platformlogger.RequestIDFromGin(c),
	})
}

func Error(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:      CodeError,
		Message:   message,
		RequestID: platformlogger.RequestIDFromGin(c),
	})
}

func FromError(c *gin.Context, err error) {
	appErr := apperror.From(err)
	if appErr == nil {
		Success(c, nil)
		return
	}
	c.JSON(appErr.HTTPStatus, Response{
		Code:      appErr.HTTPStatus,
		ErrorCode: appErr.Code,
		Message:   appErr.Message,
		RequestID: platformlogger.RequestIDFromGin(c),
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:      401,
		Message:   message,
		RequestID: platformlogger.RequestIDFromGin(c),
	})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:      400,
		Message:   message,
		RequestID: platformlogger.RequestIDFromGin(c),
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:      404,
		Message:   message,
		RequestID: platformlogger.RequestIDFromGin(c),
	})
}
