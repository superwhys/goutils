package ginutils

import "github.com/gin-gonic/gin"

type RouteHandler interface {
	HandleFunc() (data any, statusCode int, err error)
}

type TestHandler struct {
	User string `form:"user"`
}

func (th *TestHandler) HandleFunc() (data any, statusCode int, err error) {
	return nil, 200, nil
}

func test() {
	engine := gin.Default()

	engine.GET("/test", func(ctx *gin.Context) {})
}
