package qmgoutil

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/qiniu/qmgo"
	qoptions "github.com/qiniu/qmgo/options"
	"github.com/superwhys/goutils/lg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	clientCache = make(map[string]*Client)
)

type Config struct {
	Address     string
	Username    string
	Password    string
	AuthSource  string
	MaxPoolSize *uint64
}

type Client struct {
	cli  *qmgo.Client
	conf *Config
	adds []string
	uri  string
}

func (c *Config) TrimSpace() {
	c.Address = strings.TrimSpace(c.Address)
	c.Username = strings.TrimSpace(c.Username)
	c.Password = strings.TrimSpace(c.Password)
	c.AuthSource = strings.TrimSpace(c.AuthSource)
}

func (c *Config) setDefaultVal() {
	if c.AuthSource == "" {
		c.AuthSource = "admin"
	}
	if c.MaxPoolSize == nil || *c.MaxPoolSize == 0 {
		c.MaxPoolSize = new(uint64)
		*c.MaxPoolSize = 100
	}
}

func (c *Client) dial() {
	c.adds = []string{c.conf.Address}
	c.uri = c.assembleUri()
	if qmgoClient, err := c.newQmgoClient(); err != nil {
		panic(fmt.Errorf("QMongoCli: newQmgoClient err, uri:%s cause:%v", c.uri, err))
	} else {
		if c.cli != nil {
			c.cli.Close(context.Background())
		}
		c.cli = qmgoClient
	}
}

func (c *Client) newQmgoClient() (*qmgo.Client, error) {
	registryBuilder := bson.NewRegistryBuilder()
	if structCodec, err := bsoncodec.NewStructCodec(bsoncodec.JSONFallbackStructTagParser); err == nil {
		registryBuilder.RegisterDefaultEncoder(reflect.Struct, structCodec)
		registryBuilder.RegisterDefaultDecoder(reflect.Struct, structCodec)
	} else {
		lg.Warn("init JSONFallbackStructTagParser err, using default bson tag!!! cause:%v", err)
	}

	clientOptions := options.Client().
		SetMaxPoolSize(100).
		SetMaxConnIdleTime(time.Second * 10).
		SetReadPreference(readpref.Nearest()).
		SetRetryWrites(false).
		SetRegistry(registryBuilder.Build())

	conf := &qmgo.Config{
		Uri:         c.uri,
		MaxPoolSize: c.conf.MaxPoolSize,
	}
	return qmgo.NewClient(context.Background(), conf, qoptions.ClientOptions{ClientOptions: clientOptions})
}

func (c *Client) assembleUri() string {
	uri := url.URL{}
	uri.Scheme = "mongodb"
	if c.conf.Username != "" && c.conf.Password != "" {
		uri.User = url.UserPassword(c.conf.Username, c.conf.Password)
	} else if c.conf.Username != "" {
		uri.User = url.User(c.conf.Username)
	}
	uri.Host = strings.Join(c.adds, ",")
	c.conf.setDefaultVal()
	query := uri.Query()
	query.Add("authSource", c.conf.AuthSource)
	uri.RawQuery = query.Encode()
	uriStr := uri.String()
	uriStr = strings.Replace(uriStr, "?", "/?", 1)
	lg.Debugf("QMongoCli: mongodb uri -> %s", uriStr)
	return uriStr
}

type NewClientOption func(*Config)

func WithAuth(user, password string) NewClientOption {
	return func(c *Config) {
		c.Username = user
		c.Password = password
	}
}

func WithAuthSource(authSource string) NewClientOption {
	return func(c *Config) {
		c.AuthSource = authSource
	}
}

func WithMaxPoolSize(maxPoolSize int) NewClientOption {
	return func(c *Config) {
		c.MaxPoolSize = new(uint64)
		*c.MaxPoolSize = uint64(maxPoolSize)
	}
}

func NewClient(address string, opts ...NewClientOption) *Client {
	conf := &Config{}

	conf.Address = address

	for _, opt := range opts {
		opt(conf)
	}
	return NewClientWithConfig(conf)
}

func NewClientWithModel(m QmgoModel) *Client {
	conf := m.DBConfig()
	return NewClientWithConfig(&conf)
}

func NewClientWithConfig(conf *Config) *Client {
	conf.TrimSpace()
	if conf.Address == "" {
		panic("QMongoCli: address can't be empty")
	}
	c := &Client{
		conf: conf,
	}
	c.dial()

	return c
}

func (c *Client) Client() *qmgo.Client {
	return c.cli
}

func GetDBInstance(m QmgoModel) *Client {
	conf := m.DBConfig()

	if _, ok := clientCache[conf.Address]; !ok {
		clientCache[conf.Address] = NewClientWithConfig(&conf)
	}

	return clientCache[conf.Address]
}
