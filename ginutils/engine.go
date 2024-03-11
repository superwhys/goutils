package ginutils

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

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

func (e *Engine) RegisterRouter(ctx context.Context, method, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	handlers := make([]gin.HandlerFunc, 0, len(middlewares)+1)
	handlers = append(handlers, baseHandleFunc(ctx, handler))
	handlers = append(handlers, middlewares...)
	e.Handle(method, path, middlewares...)
}

func (e *Engine) GET(ctx context.Context, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	e.RegisterRouter(ctx, http.MethodGet, path, handler, middlewares...)
}

func (e *Engine) POST(ctx context.Context, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	e.RegisterRouter(ctx, http.MethodPost, path, handler, middlewares...)
}

func (e *Engine) PUT(ctx context.Context, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	e.RegisterRouter(ctx, http.MethodPut, path, handler, middlewares...)
}

func (e *Engine) DELETE(ctx context.Context, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	e.RegisterRouter(ctx, http.MethodDelete, path, handler, middlewares...)
}

func (rg *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{
		rg.RouterGroup.Group(relativePath, handlers...),
	}
}

func (rg *RouterGroup) RegisterRouter(ctx context.Context, method, path string, handler RouteHandler, middlewares ...gin.HandlerFunc) {
	handlers := make([]gin.HandlerFunc, 0, len(middlewares)+1)
	handlers = append(handlers, baseHandleFunc(ctx, handler))
	handlers = append(handlers, middlewares...)
	rg.Handle(method, path, handlers...)
}

func baseHandleFunc(ctx context.Context, handler RouteHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		baseHandler(c, ctx, handler)
	}
}

func baseHandler(c *gin.Context, ctx context.Context, handler RouteHandler) {
	data := handler.GetRequestParams()
	if data != nil {
		if dataT := reflect.TypeOf(data); dataT.Kind() != reflect.Pointer {
			lg.Errorc(ctx, "handler request params if not a pointer")
			AbortWithError(c, http.StatusInternalServerError, "parse request params error")
			return
		}

		var bindFunc func(any) error
		switch handler.GetParamsBindType() {
		case UriType:
			bindFunc = c.ShouldBindUri
		case BodyType:
			bindFunc = c.ShouldBind
		default:
			lg.Errorc(ctx, "parse request params error")
		}

		if err := bindFunc(data); err != nil {
			lg.Errorc(ctx, "parse request params error: %v", err)
			AbortWithError(c, http.StatusBadRequest, fmt.Sprintf("parse request params error: %v", err))
			return
		}

		handler.SetRequestParams(data)
	}

	returnData, statusCode, err := handler.HandleFunc(c)
	if err != nil {
		lg.Errorc(ctx, "handle error: %v", err)
		AbortWithError(c, statusCode, err.Error())
		return
	}

	c.Next()
	if c.IsAborted() {
		lg.Debugc(ctx, "%v handle err: %v", lg.StructName(handler), c.Errors.JSON())
		return
	}

	ReturnWithStatus(c, statusCode, Ret{
		Code: 0,
		Data: returnData,
	})
	lg.Debugc(ctx, "%v handle done, status code: %v, data: %v", lg.StructName(handler), statusCode, returnData)
}
