package ginutils

import (
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
