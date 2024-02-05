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

type RouterGroup struct {
	*gin.RouterGroup
}

func New(middlewares ...gin.HandlerFunc) *Engine {
	if !lg.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	engine.MaxMultipartMemory = 100 << 20
	engine.Use(lg.LoggerMiddleware(), gin.Recovery())
	engine.Use(middlewares...)

	return &Engine{engine}
}

func (e *Engine) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{
		e.Engine.Group(relativePath, handlers...),
	}
}

func (e *Engine) RegisterRouter(ctx context.Context, method, path string, handler RouteHandler) {
	e.Handle(method, path, func(c *gin.Context) {
		baseHandler(c, ctx, handler)
	})
}

func (rg *RouterGroup) RegisterRouter(ctx context.Context, method, path string, handler RouteHandler) {
	rg.Handle(method, path, func(c *gin.Context) {
		baseHandler(c, ctx, handler)
	})
}

func baseHandler(c *gin.Context, ctx context.Context, handler RouteHandler) {
	if err := c.ShouldBind(handler); err != nil {
		AbortWithError(c, http.StatusBadRequest, "parse request params error")
		return
	}

	returnData, statusCode, err := handler.HandleFunc(c)
	if err != nil {
		AbortWithError(c, statusCode, fmt.Sprintf("%v handler error", lg.StructName(handler)))
		return
	}

	lg.Debugc(ctx, "%v handle done, status code: %v, data: %v", lg.StructName(handler), statusCode, returnData)

	ReturnWithStatus(c, statusCode, Ret{
		Code: 0,
		Data: returnData,
	})
}
