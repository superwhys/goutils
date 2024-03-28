package httputils

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type Client struct {
	HandlerGroup
	conf       *Config
	httpClient *http.Client
	isDefault  bool
	isDebug    bool
}

func newClient(conf *Config) (*Client, error) {
	var (
		transportProxy  func(*http.Request) (*url.URL, error)
		tlsClientConfig *tls.Config
	)

	if conf.Proxy != "" {
		transportProxy = func(_ *http.Request) (*url.URL, error) {
			return url.Parse(conf.Proxy)
		}
		tlsClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		transportProxy = http.ProxyFromEnvironment
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	cli := &http.Client{
		Transport: &http.Transport{
			Proxy:                 transportProxy,
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       tlsClientConfig,
		},
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       conf.RequestTimeOut,
	}
	return &Client{
		conf:       conf,
		httpClient: cli,
	}, nil
}

func New(conf *Config) *Client {
	if conf == nil {
		panic("need httpclient config")
	}
	if conf.RequestTimeOut == 0 {
		panic("httpclient request timeout is not legal")
	}
	client, err := newClient(conf)
	if err != nil {
		panic(errors.Wrap(err, "new httpclient failed"))
	}
	client.isDefault = false
	return client
}

func Default() *Client {
	conf := &Config{
		RequestTimeOut: 10 * time.Second,
	}

	cli := New(conf)
	cli.Use(
		HandlerDuration(),
		HandlerDebugDuration(),
		RequestDefaultHeaderHandler(),
		RequestParamsHandler(),
		RequestBodyReaderHandler(),
		DefaultHTTPHandler(),
		DefaultResponseBodyHandler(),
	)
	cli.isDefault = true
	return cli
}

func (cli *Client) Use(handler ...HandleFunc) {
	cli.HandlerGroup.Use(handler...)
}

func (cli *Client) SetTimeout(timeout time.Duration) {
	cli.httpClient.Timeout = timeout
}

func (cli *Client) Start() *Context {
	ctx := NewContext(cli)
	return ctx
}

func (cli *Client) DoRequest(ctx context.Context, url, method string, queryParams Params, header *Header, body []byte, callBack ...HandleFunc) *Response {
	cli.Use(callBack...)
	return cli.
		Start().
		SetMethod(method).
		SetURL(url).
		SetQueryParams(queryParams).
		SetHeaders(header).
		SetBody(body).
		Fetch(ctx)
}

func (cli *Client) Get(ctx context.Context, url string, queryParams Params, header *Header, callBack ...HandleFunc) *Response {
	return cli.DoRequest(ctx, url, http.MethodGet, queryParams, header, nil, callBack...)
}

func (cli *Client) Post(ctx context.Context, url string, body []byte, header *Header, callBack ...HandleFunc) *Response {
	return cli.DoRequest(ctx, url, http.MethodPost, nil, header, body, callBack...)
}

func (cli *Client) Delete(ctx context.Context, url string, body []byte, header *Header, callBack ...HandleFunc) *Response {
	return cli.DoRequest(ctx, url, http.MethodDelete, nil, header, body, callBack...)
}

func (cli *Client) Put(ctx context.Context, url string, body []byte, header *Header, callBack ...HandleFunc) *Response {
	return cli.DoRequest(ctx, url, http.MethodPut, nil, header, body, callBack...)
}
