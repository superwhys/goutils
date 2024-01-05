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
