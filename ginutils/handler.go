package ginutils

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/superwhys/goutils/lg"
)

type BindType uint

const (
	BodyType BindType = iota
	UriType
)

type RouteHandler interface {
	GetParamsBindType() BindType
	GetRequestParams() any
	SetRequestParams(any)
	HandleFunc(c *gin.Context) (data any, statusCode int, err error)
}

type DefaultHandler struct {
	Params any
}

func (h *DefaultHandler) GetParamsBindType() BindType {
	return BodyType
}

func (h *DefaultHandler) GetRequestParams() any {
	return h.Params
}

func (h *DefaultHandler) SetRequestParams(data any) {
	h.Params = data
}

func (h *DefaultHandler) GetParams() any {
	return h.Params
}

func BaseHandleFunc(handler RouteHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		baseHandler(c, context.Background(), handler)
	}
}

func BaseHandleFuncWithContext(ctx context.Context, handler RouteHandler) gin.HandlerFunc {
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
