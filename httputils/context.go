package httputils

import (
	"io"
	"math"
	"net/http"
	"time"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
)

const abortIndex = math.MaxInt8 >> 1

type Context struct {
	ctx        context.Context
	conf       *Config
	Header     *Header
	Method     string
	Url        string
	Params     Params
	Body       []byte
	bodyReader io.Reader

	handlers HandlersChain
	index    int8

	cli          *Client
	client       *http.Client
	Request      *http.Request
	Response     *http.Response
	ResponseBody []byte

	duration time.Duration
	err      error
}

func NewContext(client *Client) *Context {
	return &Context{
		cli:      client,
		conf:     client.conf,
		client:   client.httpClient,
		handlers: client.HandlerGroup.Handlers,
		index:    -1,
	}
}

func (c *Context) AddError(err error) {
	c.err = multierror.Append(c.err, err)
}

func (c *Context) GetError() error {
	return c.err
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) SetURL(url string) *Context {
	if url != "" {
		c.Url = url
	}
	return c
}

func (c *Context) SetMethod(method string) *Context {
	c.Method = method
	return c
}

func (c *Context) SetQueryParams(params Params) *Context {
	if params != nil {
		c.Params = params
	}
	return c
}

func (c *Context) SetHeaders(header *Header) *Context {
	if header != nil {
		c.Header = header
	}
	return c
}

func (c *Context) SetBody(body []byte) *Context {
	if body != nil {
		c.Body = body
	}
	return c
}

func (c *Context) FormBody(form *Form) *Context {
	if form != nil {
		c.Body = form.Encode()
	}
	return c
}

func (c *Context) Fetch(ctx context.Context) *Response {
	if c.err != nil {
		return &Response{err: c.err}
	}
	c.ctx = ctx
	c.Next()
	return &Response{Response: c.Response, respByte: c.ResponseBody, err: c.err}
}
