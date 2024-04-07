package ginutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/superwhys/goutils/httputils"
)

type TestSessionHandler struct{}

func (th *TestSessionHandler) HandleFunc(ctx context.Context, c *gin.Context) HandleResponse {
	session := sessions.Default(c)

	session.Set("share_code", "aaabbbccc")
	session.Save()

	return &Ret{
		Code: 200,
		Data: "success",
	}
}

func TestMain(m *testing.M) {
	store, err := redis.NewStoreWithDB(11, "tcp", "localhost:6379", "", "0", []byte("ia7KzUr2fjrM"))
	if err != nil {
		panic(err)
	}
	store.Options(sessions.Options{MaxAge: 3600})

	engine := New(sessions.Sessions("__session", store))

	engine.POST("/test_session", &TestSessionHandler{})

	go func() {
		engine.Run(":8081")
	}()

	m.Run()
}

func TestSession(t *testing.T) {
	client := httputils.Default()
	resp := client.Post(context.TODO(), "http://localhost:8081/test_session", nil, nil)

	respStr, err := resp.BodyString()
	if err != nil {
		t.Error(err)
		return
	}
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		fmt.Println(cookie.String())
	}

	fmt.Println(respStr)
	time.Sleep(time.Second * 10)
}
