package redisutils

import (
	"github.com/superwhys/goutils/dialer"
	"github.com/superwhys/goutils/flags"
	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/slowinit"
)

type RedisConf struct {
	Server   string `desc:"redis server name (default localhost:6379)"`
	Password string `desc:"redis server password"`
	Db       int    `desc:"redis db (default 0)"`
	MaxIdle  int    `desc:"redis maxidle (default 100)"`
}

func (rc *RedisConf) SetDefault() {
	rc.Server = "localhost:6379"
	rc.Db = 0
	rc.MaxIdle = 100
}

var (
	redisConfFlag = flags.Struct("redisConf", &RedisConf{}, "Redis config")
)

var Client *RedisClient

func init() {
	slowinit.RegisterObject("redisClient", func() error {
		conf := &RedisConf{}
		lg.PanicError(redisConfFlag(conf))

		var pwd []string
		if conf.Password != "" {
			pwd = append(pwd, conf.Password)
		}
		Client = NewRedisClient(dialer.DialRedisPool(
			conf.Server,
			conf.Db,
			conf.MaxIdle,
			pwd...,
		))
		return nil
	})
}
