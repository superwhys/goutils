package httputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type HandleFunc func(c *Context)
type HandlersChain []HandleFunc

type HandlerGroup struct {
	Handlers HandlersChain
	client   *Client
}

func (group *HandlerGroup) Use(handler ...HandleFunc) {
	group.Handlers = append(group.Handlers, handler...)
}

func HandlerDuration() HandleFunc {
	return func(c *Context) {
		startTime := time.Now()
		c.Next()
		c.duration = time.Now().Sub(startTime)
	}
}

func HandlerDebugDuration() HandleFunc {
	return func(c *Context) {
		c.Next()
		if !c.cli.isDebug {
			return
		}
		now := time.Now().Format("2006-01-02 15:04:05")
		params, err := json.Marshal(c.Params)
		if err != nil {
			c.err = multierror.Append(c.err, errors.Wrap(err, "json marshal params"))
			return
		}
		fmt.Printf("[http] %v | %v | url: %v, params: %v, duration: %v, err: %v\n", now, c.Method, c.Url, string(params), c.duration, c.err)
	}
}

func RequestBodyReaderHandler() HandleFunc {
	return func(c *Context) {
		var bodyReader io.Reader
		if c.Body != nil {
			bodyReader = bytes.NewReader(c.Body)
		}
		c.bodyReader = bodyReader
	}
}

func RequestParamsHandler() HandleFunc {
	return func(c *Context) {
		if c.Params != nil {
			urlParse, err := url.ParseRequestURI(c.Url)
			if err != nil {
				c.err = multierror.Append(c.err, errors.Wrap(err, "parse request url"))
				c.Abort()
				return
			}
			q := urlParse.Query()
			for key, value := range c.Params {
				if !q.Has(key) {
					q.Add(key, value)
				}
			}
			urlParse.RawQuery = q.Encode()
			c.Url = urlParse.String()
		}
	}
}

func RequestDefaultHeaderHandler() HandleFunc {
	return func(c *Context) {
		if c.Header == nil {
			c.Header = DefaultJsonHeader()
		}
	}
}

func DefaultHTTPHandler() HandleFunc {
	return func(c *Context) {
		req, err := http.NewRequest(c.Method, c.Url, c.bodyReader)
		if err != nil {
			c.err = multierror.Append(c.err, errors.Wrap(err, "generate request"))
			c.Abort()
			return
		}
		req.Header = c.Header.Header

		resp, err := c.client.Do(req)
		c.Request = req
		c.Response = resp
		if err != nil {
			c.err = multierror.Append(c.err, errors.Wrap(err, "do request"))
			c.Abort()
			return
		}
	}
}

func DefaultResponseBodyHandler() HandleFunc {
	return func(c *Context) {
		if c.err != nil {
			c.Abort()
			return
		}
		defer c.Response.Body.Close()
		respByte, err := ioutil.ReadAll(c.Response.Body)
		if err != nil {
			c.err = multierror.Append(c.err, errors.Wrap(err, "read response body"))
			c.Abort()
			return
		}
		c.ResponseBody = respByte
	}
}
