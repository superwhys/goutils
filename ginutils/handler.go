package ginutils

import "github.com/gin-gonic/gin"

type RouteHandler interface {
	HandleFunc(c *gin.Context) (data any, statusCode int, err error)
}
