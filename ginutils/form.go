package ginutils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func SaveFormFile(c *gin.Context, key string, savePath func(fileName string) string) (string, error) {
	header, err := c.FormFile(key)
	if err != nil {
		return "", errors.Wrap(err, "getFileFromForm")
	}

	filePath := savePath(header.Filename)
	err = c.SaveUploadedFile(header, filePath)
	if err != nil {
		return "", errors.Wrap(err, "save upload file")
	}
	return filePath, nil
}

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
