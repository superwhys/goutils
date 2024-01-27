package ginutils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Ret struct {
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func AbortWithError(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, Ret{
		Code:    code,
		Message: message,
	})
}

func StatusOk(c *gin.Context, data any) {
	ReturnWithStatus(c, http.StatusOK, Ret{
		Code: 0,
		Data: data,
	})
}

func ReturnWithStatus(c *gin.Context, code int, data any) {
	c.JSON(code, data)
}

type Handler func() (data any, statusCode int, err error)

func HandlerFunc(handler Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, statusCode, err := handler()
		if err != nil || statusCode != http.StatusOK {
			AbortWithError(c, statusCode, err.Error())
			return
		}

		StatusOk(c, data)
	}
}
