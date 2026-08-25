package response

import (
	"net/http"

	"github.com/awydd/iam/pkg/response/code"
	"github.com/gin-gonic/gin"
)

type Response[T any] struct {
	Code    int    `json:"code"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

func JSON[T any](c *gin.Context, httpStatus int, bizCode code.Code, data T) {
	c.JSON(httpStatus, Response[T]{
		Code:    bizCode.Int(),
		Status:  bizCode == code.Success,
		Message: bizCode.Message(),
		Data:    data,
	})
}

func Success[T any](c *gin.Context, data T) {
	JSON(c, http.StatusOK, code.Success, data)
}

func Err(c *gin.Context, bizCode code.Code) {
	JSON[any](c, bizCode.HTTPStatus(), bizCode, nil)
}

func ErrWithData[T any](c *gin.Context, bizCode code.Code, data T) {
	JSON(c, bizCode.HTTPStatus(), bizCode, data)
}

func ErrMessage(c *gin.Context, bizCode code.Code, message string) {
	c.JSON(bizCode.HTTPStatus(), Response[any]{
		Code:    bizCode.Int(),
		Status:  false,
		Message: message,
	})
}

func OK(c *gin.Context) {
	Success[any](c, nil)
}

func BadRequest(c *gin.Context)      { Err(c, code.BadRequest) }
func Unauthorized(c *gin.Context)    { Err(c, code.Unauthorized) }
func Forbidden(c *gin.Context)       { Err(c, code.Forbidden) }
func NotFound(c *gin.Context)        { Err(c, code.NotFound) }
func Conflict(c *gin.Context)        { Err(c, code.Conflict) }
func TooManyRequests(c *gin.Context) { Err(c, code.TooManyRequests) }
func InternalError(c *gin.Context)   { Err(c, code.InternalError) }
