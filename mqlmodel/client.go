package mqlmodel

import (
	"strings"

	"github.com/superwhys/goutils/dialer"
	"github.com/superwhys/goutils/lg"
	"gorm.io/gorm"
)

type config struct {
	AuthConf
}

type client struct {
	db     *gorm.DB
	config *config
}

func (c *config) TrimSpace() {
	c.Username = strings.TrimSpace(c.Username)
	c.Password = strings.TrimSpace(c.Password)
	c.Instance = strings.TrimSpace(c.Instance)
	c.Database = strings.TrimSpace(c.Database)
}

func NewClient(conf *config) *client {
	conf.TrimSpace()
	if conf.Instance == "" {
		panic("mqlClient: instance can not be empty")
	}

	c := &client{config: conf}
	c.dial()
	return c
}

func (c *client) dialGorm() (*gorm.DB, error) {
	return dialer.DialGorm(
		c.config.Instance,
		dialer.WithAuth(c.config.Username, c.config.Password),
		dialer.WithDBName(c.config.Database),
	)
}

func (c *client) dial() {
	db, err := c.dialGorm()
	lg.PanicError(err, "mqlClient: new client error")

	c.db = db
}

func (c *client) DB() *gorm.DB {
	return c.db
}
