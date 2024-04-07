package httputils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	Cli    *Client
	srvApi = "http://127.0.0.1:29940"
)

type params struct {
	Name string `json:"name" form:"name"`
}

type FormParams struct {
	Name     string `json:"name" form:"name"`
	Password string `json:"password" form:"password"`
}

type BodyParams struct {
	Text string `json:"text" form:"text"`
}

func TestMain(m *testing.M) {
	fmt.Println("done")
	Cli = Default()
	Cli.isDebug = true

	go func() {
		r := gin.Default()
		r.GET("/test_get", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "do success",
			})
		})
		r.GET("/test_get_params", func(c *gin.Context) {
			p := &params{}
			err := c.ShouldBind(p)
			if err != nil {
				c.JSON(400, gin.H{"err": err.Error()})
				return
			}
			fmt.Printf("params:%v\n", p)
			c.JSON(200, gin.H{"message": fmt.Sprintf("hello %v", p.Name)})
		})
		r.POST("/test_post", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "do success",
			})
		})
		r.POST("/test_post_form", func(c *gin.Context) {
			postParams := &FormParams{}
			err := c.ShouldBind(postParams)
			if err != nil {
				c.JSON(400, gin.H{"err": err.Error()})
				return
			}
			c.JSON(200, gin.H{
				"message": fmt.Sprintf("%v-%v success", postParams.Name, postParams.Password),
			})
		})
		r.POST("/test_post_body", func(c *gin.Context) {
			bodyParams := &BodyParams{}
			err := c.ShouldBind(bodyParams)
			if err != nil {
				c.JSON(400, gin.H{"err": err.Error()})
				return
			}
			fmt.Println("get post body: ", bodyParams)

			c.JSON(200, gin.H{
				"message": fmt.Sprintf("%v success", bodyParams.Text),
			})

		})
		r.GET("/test_get_error", func(c *gin.Context) {
			c.JSON(400, gin.H{
				"message": "do error",
			})
		})
		panic(r.Run(":29940"))
	}()
	time.Sleep(2 * time.Second)

	os.Exit(m.Run())
}

func addParamsHandler(c *Context) {
	fmt.Printf("url:%v\n", c.Url)
	urlParse, err := url.ParseRequestURI(c.Url)
	if err != nil {
		c.err = errors.Wrap(err, "parse request url")
		return
	}
	q := urlParse.Query()
	p := Params{"name": "superwhys"}
	for key, value := range p {
		if !q.Has(key) {
			q.Add(key, value)
		}
	}
	urlParse.RawQuery = q.Encode()
	c.Url = urlParse.String()
}

func TestNewClientGet(t *testing.T) {
	t.Run("testNewClientGet", func(t *testing.T) {
		newCli := New(&Config{
			RequestTimeOut: 5 * time.Second,
		})
		newCli.Use(addParamsHandler)
		resp, err := newCli.Get(
			context.Background(),
			fmt.Sprintf("%v/%v", srvApi, "test_get_params"),
			nil,
			DefaultJsonHeader(),
		).BodyString()
		assert.Nil(t, err)
		assert.Equal(t, `{"message":"hello superwhys"}`, resp)
	})
}

func TestClientGet(t *testing.T) {
	t.Run("testContextFetch", func(t *testing.T) {
		resp, err := Cli.Get(
			context.Background(),
			fmt.Sprintf("%v/%v", srvApi, "test_get"),
			nil,
			DefaultJsonHeader(),
		).BodyString()
		assert.Nil(t, err)
		assert.Equal(t, `{"message":"do success"}`, resp)
	})
}

type message struct {
	Message string `json:"message"`
}

func TestClientGetWithCallBack(t *testing.T) {
	msg := &message{}

	t.Run("testContextFetch", func(t *testing.T) {
		cli := New(&Config{RequestTimeOut: time.Second * 10})
		cli.Use(DefaultHTTPHandler())
		resp := cli.Get(
			context.Background(),
			fmt.Sprintf("%v/%v", srvApi, "test_get"),
			nil,
			DefaultJsonHeader(),
			func(c *Context) {
				defer c.Response.Body.Close()
				b, err := ioutil.ReadAll(c.Response.Body)
				if err != nil {
					fmt.Printf("read body error:%v\n", err)
					return
				}
				fmt.Printf("body:%v\n", string(b))
				err = json.Unmarshal(b, &msg)
				if err != nil {
					fmt.Printf("unmarshal body error:%v\n", err)
					return
				}
				fmt.Printf("message:%v\n", msg.Message)
			},
		)
		assert.Nil(t, resp.Error())
		assert.Equal(t, &message{Message: "do success"}, msg)
	})
}

func TestClientGetWithCallBack2(t *testing.T) {
	msg := &message{}

	t.Run("testContextFetch", func(t *testing.T) {
		cli := New(&Config{RequestTimeOut: time.Second * 10})
		cli.Use(DefaultHTTPHandler())
		resp := cli.Get(
			context.Background(),
			fmt.Sprintf("%v/%v", srvApi, "test_get"),
			nil,
			DefaultJsonHeader(),
			func(c *Context) {
				if c.Response.StatusCode == 200 {
					c.Abort()
				}
			},
			func(c *Context) {
				defer c.Response.Body.Close()
				b, err := ioutil.ReadAll(c.Response.Body)
				if err != nil {
					fmt.Printf("read body error:%v\n", err)
					return
				}
				fmt.Printf("body:%v\n", string(b))
				err = json.Unmarshal(b, &msg)
				if err != nil {
					fmt.Printf("unmarshal body error:%v\n", err)
					return
				}
				fmt.Printf("message:%v\n", msg.Message)
			},
		)
		assert.Nil(t, resp.Error())
		assert.Equal(t, &message{}, msg)
	})
}

func TestClientPost(t *testing.T) {
	t.Run("testClientPost", func(t *testing.T) {
		resp, err := Cli.Post(
			context.Background(),
			fmt.Sprintf("%v/%v", srvApi, "test_post"),
			nil,
			DefaultJsonHeader(),
		).BodyString()
		assert.Nil(t, err)
		fmt.Println(resp)
		assert.Equal(t, `{"message":"do success"}`, resp)
	})
}

func TestClientPostForm(t *testing.T) {
	t.Run("testClientPostForm", func(t *testing.T) {
		resp, err := Cli.Post(
			context.Background(),
			fmt.Sprintf("%v/%v", srvApi, "test_post_form"),
			NewForm().Add("name", "superwhys").Add("password", "123456").Encode(),
			DefaultFormUrlEncodedHeader(),
		).BodyString()
		assert.Nil(t, err)
		fmt.Println(resp)
		assert.Equal(t, `{"message":"superwhys-123456 success"}`, resp)
	})
}

func TestClientGetError(t *testing.T) {
	t.Run("testClientGetError", func(t *testing.T) {
		resp := Cli.Get(
			context.Background(),
			"abcd",
			nil,
			DefaultJsonHeader(),
		)
		assert.NotNil(t, resp.Error())
	})
}

func TestClientPostBody(t *testing.T) {
	t.Run("testClientPostBody", func(t *testing.T) {
		resp, err := Cli.Post(
			context.Background(),
			fmt.Sprintf("%v/%v", srvApi, "test_post_body"),
			NewJsonBody().Add("text", "helloworld").Encode(),
			DefaultJsonHeader(),
		).BodyString()
		assert.Nil(t, err)
		fmt.Println(resp)
		assert.Equal(t, `{"message":"helloworld success"}`, resp)
	})
}
