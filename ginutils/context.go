package ginutils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AbortWithError(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

func StatusOk(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"success": "OK",
		"data":    data,
	})
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
