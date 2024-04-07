package ginutils

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superwhys/goutils/lg"
)

type RouterGroup struct {
	*gin.RouterGroup
	ctx context.Context
}

type Engine struct {
	*RouterGroup
	engine *gin.Engine
}

func NewGinEngine(middlewares ...gin.HandlerFunc) *gin.Engine {
	if !lg.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	engine.MaxMultipartMemory = 100 << 20
	engine.Use(lg.LoggerMiddleware(), gin.Recovery())
	engine.Use(middlewares...)

	return engine
}

func New(middlewares ...gin.HandlerFunc) *Engine {
	engine := NewGinEngine(middlewares...)
	return &Engine{
		RouterGroup: &RouterGroup{
			RouterGroup: &engine.RouterGroup,
			ctx:         context.Background(),
		},
		engine: engine,
	}
}

func (e *Engine) Run(addr ...string) error {
	return e.engine.Run(addr...)
}

func (e *Engine) GetGinEngine() *gin.Engine {
	return e.engine
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.engine.ServeHTTP(w, req)
}

func (g *RouterGroup) Static(relativePath, root string) gin.IRoutes {
	return g.RouterGroup.Static(relativePath, root)
}

func (g *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) gin.IRoutes {
	return g.RouterGroup.StaticFS(relativePath, fs)
}

func (g *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{
		RouterGroup: g.RouterGroup.Group(relativePath, handlers...),
		ctx:         g.ctx,
	}
}

func (g *RouterGroup) RegisterRouter(method, path string, handlers ...Handler) {
	g.Handle(method, path, wrapHandler(g.ctx, handlers...)...)
}

func (g *RouterGroup) GET(path string, handler ...Handler) {
	g.RegisterRouter(http.MethodGet, path, handler...)
}

func (g *RouterGroup) POST(path string, handler ...Handler) {
	g.RegisterRouter(http.MethodPost, path, handler...)
}

func (g *RouterGroup) PUT(path string, handler ...Handler) {
	g.RegisterRouter(http.MethodPut, path, handler...)
}

func (g *RouterGroup) DELETE(path string, handler ...Handler) {
	g.RegisterRouter(http.MethodDelete, path, handler...)
}
