package ginutils

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	httpUtils "github.com/superwhys/goutils/http"
	"github.com/superwhys/goutils/lg"
)

type RouteTest struct {
	Text string `form:"text"`
}

func (rt *RouteTest) HandleFunc() (data any, statusCode int, err error) {
	return fmt.Sprintf("handle text: %v", rt.Text), 200, nil
}

func TestMain(m *testing.M) {
	lg.EnableDebug()

	engine := New()
	engine.RegisterRouter(context.Background(), http.MethodGet, "/test", &RouteTest{})

	go engine.Run(":8000")

	m.Run()
}

func TestTestRequest(t *testing.T) {
	t.Run("TestRequest", func(t *testing.T) {
		cli := httpUtils.Default()
		resp, err := cli.Get(context.Background(), "http://localhost:8000/test", httpUtils.NewParams().Add("text", "hello world"), nil).BodyString()
		if err != nil {
			t.Error(err)
			return
		}
		t.Logf("request resp: %v", resp)
	})
}
