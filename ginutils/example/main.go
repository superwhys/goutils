package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/superwhys/goutils/flags"
	"github.com/superwhys/goutils/ginutils"
	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/service"
)

type Params struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

type TestHandler struct {
	ginutils.DefaultHandler
}

func (h *TestHandler) GetRequestParams() any {
	return &Params{}
}

func (h *TestHandler) HandleFunc(c *gin.Context) (data any, statusCode int, err error) {
	params := h.GetParams()
	lg.Info(lg.Jsonify(params))

	return fmt.Sprintf("success: %v", lg.Jsonify(params)), 200, nil
}

func main() {
	flags.Parse()

	router := ginutils.New()
	grp := router.Group("/v1")
	grp.GET(context.Background(), "/api/test", &TestHandler{})

	srv := service.NewSuperService(
		service.WithHttpHandler("", router),
	)

	srv.ListenAndServer(8080)
}
