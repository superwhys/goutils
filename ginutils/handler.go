package ginutils

import (
	"github.com/gin-gonic/gin"
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
