package ginutils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superwhys/goutils/lg"
)

type Engine struct {
	*gin.Engine
}

func New() *Engine {
	engine := gin.New()

	if !lg.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine.MaxMultipartMemory = 100 << 20
	engine.Use(lg.LoggerMiddleware(), gin.Recovery())

	return &Engine{engine}
}

func (e *Engine) RegisterRouter(ctx context.Context, method, path string, handler RouteHandler) {
	e.Handle(method, path, func(c *gin.Context) {
		if err := c.ShouldBind(handler); err != nil {
			AbortWithError(c, http.StatusBadRequest, "parse request params error")
			return
		}

		returnData, statusCode, err := handler.HandleFunc()
		if err != nil {
			AbortWithError(c, statusCode, fmt.Sprintf("%v handler error", lg.StructName(handler)))
			return
		}

		lg.Debugc(ctx, "%v handle done, status code: %v, data: %v", lg.StructName(handler), statusCode, returnData)

		ReturnWithStatus(c, statusCode, Ret{
			Code: 0,
			Data: returnData,
		})
	})
}
