package middlewares

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/superwhys/goutils/ginutils/v2"
	"github.com/superwhys/goutils/httputils"
	"github.com/superwhys/goutils/lg"
)

type UserInfoClaims struct {
	User string `json:"user"`
	jwt.RegisteredClaims
}

type TestAuthGetter struct{}

func (g *TestAuthGetter) GetAuth(name string) (string, error) {
	if name != "yong" {
		return "", fmt.Errorf("name err")
	}
	return "testpwd", nil
}
func (g *TestAuthGetter) SetAuth(name string) error {
	return nil
}

type TestHandler struct {
}

func (h *TestHandler) HandleFunc(ctx context.Context, c *gin.Context) ginutils.HandleResponse {
	lg.Info("success")
	return &ginutils.Ret{
		Code: 200,
		Data: "success",
	}
}

func TestMain(m *testing.M) {
	engine := ginutils.New()
	engine.POST("/test_basic", BasicAuthMiddlewareHandler(&TestAuthGetter{}), &TestHandler{})
	engine.POST("/test_jwt", JWTMiddlewareHandler("test-key", &UserInfoClaims{}), &TestHandler{})
	go func() {
		engine.Run(":8081")
	}()
	m.Run()
}

func TestBasicAuth(t *testing.T) {
	client := httputils.Default()
	resp := client.Post(context.TODO(), "http://localhost:8081/test_basic", nil, httputils.NewHeader().BasicAuth("yong", "testpwd"))
	respStr, err := resp.BodyString()
	if err != nil {
		t.Error(err)
		return
	}
	if strings.Contains(respStr, "Basic auth failure") {
		t.Error("basic auth expect success but failed")
	}
	fmt.Println("request resp: ", respStr)

	resp = client.Post(context.TODO(), "http://localhost:8081/test_basic", nil, httputils.NewHeader().BasicAuth("yong", "testpwd1"))
	respStr, err = resp.BodyString()
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(respStr, "Basic auth failure") {
		t.Error("basic auth expect failed but success")
	}
	fmt.Println("request resp: ", respStr)
}
