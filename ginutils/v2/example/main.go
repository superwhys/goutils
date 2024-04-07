package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superwhys/goutils/ginutils/v2"
	"github.com/superwhys/goutils/lg"
)

type BeforeRequestMiddleware struct {
}

func (m *BeforeRequestMiddleware) HandleFunc(ctx context.Context, c *gin.Context) ginutils.HandleResponse {
	lg.Infoc(ctx, "into before request middleware")
	c.Next()
	lg.Infoc(ctx, "before request middleware done")
	return nil
}

type AfterRequestMiddleware struct {
}

func (m *AfterRequestMiddleware) HandleFunc(ctx context.Context, c *gin.Context) ginutils.HandleResponse {
	lg.Infoc(ctx, "into after request middleware")
	return nil
}

type HelloHandler struct {
	Id   string `uri:"id" json:"id,omitempty"`
	Name string `form:"name" json:"name,omitempty"`
	Age  int    `form:"age" json:"age,omitempty"`
}

func (h *HelloHandler) HandleFunc(ctx context.Context, c *gin.Context) ginutils.HandleResponse {
	if err := ginutils.BindParms(c, h); err != nil {
		lg.Errorf("bind params error: %v", err)
		ginutils.AbortWithError(c, http.StatusBadRequest, err.Error())
		return nil
	}
	lg.Info(lg.Jsonify(h))

	ret := &ginutils.Ret{
		Code: 200,
		Data: h,
	}

	return ret
}

func main() {
	lg.EnableDebug()
	engine := ginutils.New()

	engine.POST("/hello/:id", &BeforeRequestMiddleware{}, &HelloHandler{}, &AfterRequestMiddleware{})

	engine.Run(":8080")
}
