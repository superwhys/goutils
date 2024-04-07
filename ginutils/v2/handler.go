package ginutils

import (
	"context"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/lg"
)

type Handler interface {
	HandleFunc(ctx context.Context, c *gin.Context) HandleResponse
}

func wrapHandler(ctx context.Context, handlers ...Handler) []gin.HandlerFunc {
	handlerFuncs := make([]gin.HandlerFunc, 0, len(handlers))

	for _, handler := range handlers {
		handlerFuncs = append(handlerFuncs, wrapDefaultHandler(ctx, handler))
	}

	return handlerFuncs
}

func bindData(c *gin.Context, data any) error {
	if dataT := reflect.TypeOf(data); dataT.Kind() != reflect.Pointer {
		return errors.New("data instance need a struct pointer")
	}

	if err := c.ShouldBind(data); err != nil {
		return errors.Wrap(err, "parse body params")
	}

	if err := c.ShouldBindUri(data); err != nil {
		return errors.Wrap(err, "parse uri params")
	}

	return nil
}

func getHandlerName(handler Handler) string {
	ele := reflect.TypeOf(handler).Elem()
	return ele.Name()
}

func wrapDefaultHandler(ctx context.Context, handler Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx = lg.With(ctx, "[%v]", getHandlerName(handler))

		_, exists := GetParams(c)
		if !exists {
			params, err := ParseMapParams(c)
			if err != nil {
				lg.Errorc(ctx, "parse params error: %v", err)
				AbortWithError(c, http.StatusInternalServerError, "请求失败")
				return
			}
			c.Set(paramsKey, params)
		}

		ret := handler.HandleFunc(ctx, c)
		if ret != nil && ret.GetError() != nil {
			lg.Errorc(ctx, "%v handle err: %v", lg.StructName(handler), c.Errors.JSON())
			AbortWithError(c, ret.GetCode(), ret.GetMessage())
			return
		}

		if c.IsAborted() {
			return
		}

		if ret != nil {
			ReturnWithStatus(c, ret.GetCode(), ret.GetData())
		}
	}
}

type ginHandlerFuncHandler struct {
	handlerFunc gin.HandlerFunc
}

func (h *ginHandlerFuncHandler) HandleFunc(_ context.Context, c *gin.Context) HandleResponse {
	h.handlerFunc(c)
	return nil
}

func WrapGinHandlerFunc(handlerFunc gin.HandlerFunc) Handler {
	return &ginHandlerFuncHandler{handlerFunc}
}
