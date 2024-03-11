package ginutils

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superwhys/goutils/lg"
)

type RouterGroup struct {
	*gin.RouterGroup
}

type Engine struct {
	*RouterGroup
	engine *gin.Engine
}

func New(middlewares ...gin.HandlerFunc) *Engine {
	if !lg.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	engine.MaxMultipartMemory = 100 << 20
	engine.Use(lg.LoggerMiddleware(), gin.Recovery())
	engine.Use(middlewares...)

	return &Engine{
		RouterGroup: &RouterGroup{
			&engine.RouterGroup,
		},
		engine: engine,
	}
}

func (e *Engine) GetGinEngine() *gin.Engine {
	return e.engine
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.engine.ServeHTTP(w, req)
}

func (g *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{
		g.RouterGroup.Group(relativePath, handlers...),
	}
}

func (g *RouterGroup) Static(relativePath, root string) gin.IRoutes {
	return g.RouterGroup.Static(relativePath, root)
}

func (g *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) gin.IRoutes {
	return g.RouterGroup.StaticFS(relativePath, fs)
}

func (g *RouterGroup) GET(ctx context.Context, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	g.RegisterRouter(ctx, http.MethodGet, path, handler, middlewares...)
}

func (g *RouterGroup) POST(ctx context.Context, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	g.RegisterRouter(ctx, http.MethodPost, path, handler, middlewares...)
}

func (g *RouterGroup) PUT(ctx context.Context, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	g.RegisterRouter(ctx, http.MethodPut, path, handler, middlewares...)
}

func (g *RouterGroup) DELETE(ctx context.Context, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	g.RegisterRouter(ctx, http.MethodDelete, path, handler, middlewares...)
}

func (g *RouterGroup) RegisterRouter(ctx context.Context, method, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	handlers := make([]gin.HandlerFunc, 0, len(middlewares)+1)
	handlers = append(handlers, BaseHandleFuncWithContext(ctx, handler))
	handlers = append(handlers, middlewares...)
	g.Handle(method, path, handlers...)
}
