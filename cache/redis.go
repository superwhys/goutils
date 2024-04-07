package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/redisutils"
)

var _ Cache = (*RedisCache)(nil)

var (
	defaultRedisTTL = time.Minute * 10
)

type RedisCache struct {
	*redisutils.RedisClient
	prefix string
}

type RedisCacheOption func(c *RedisCache)

func WithPrefix(prefix string) RedisCacheOption {
	return func(c *RedisCache) {
		c.prefix = prefix
	}
}

func NewRedisCache(pool *redis.Pool, opts ...RedisCacheOption) *RedisCache {
	rc := &RedisCache{
		RedisClient: redisutils.NewRedisClient(pool),
	}

	for _, opt := range opts {
		opt(rc)
	}

	return rc
}

func (c *RedisCache) Get(key string, out any) error {
	var p payload

	if err := c.RedisClient.Get(key, &p); err != nil {
		return err
	}

	return p.Get(out)
}

func (c *RedisCache) Set(key string, value any) error {
	return c.SetWithTTL(key, value, 0)
}

func (c *RedisCache) GetOrCreate(key string, creater Creater, out any) error {
	return c.GetOrCreateWithTTL(key, 0, creater, out)
}

func (c *RedisCache) Delete(key string) error {
	return c.RedisClient.Delete(key)
}

func (c *RedisCache) setWithTTL(conn redis.Conn, key string, value any, ttl time.Duration) (err error) {
	if ttl > 0 {
		_, err = conn.Do("SET", key, value, "EX", int(ttl.Seconds()))
	} else {
		_, err = conn.Do("SET", key, value)
	}

	return
}

func (c *RedisCache) packKey(key string) string {
	if c.prefix != "" {
		key = fmt.Sprintf("%v::%v", c.prefix, key)
	}

	return key
}

func (c *RedisCache) GetOrCreateWithTTL(key string, ttl time.Duration, creator Creater, out any) error {
	conn := c.GetConn()
	defer conn.Close()

	key = c.packKey(key)

	var p payload
	data, err := redis.Bytes(conn.Do("GET", key))
	// no data in redis
	if err != nil && errors.Is(err, redis.ErrNil) {
		p = newPayload(creator())
		data, err = json.Marshal(p)
		if err != nil {
			return errors.Wrap(err, "json.Marshal.redisData")
		}

		if err = c.setWithTTL(conn, key, data, ttl); err != nil {
			return errors.Wrap(err, "redis.setTTL")
		}
		return nil
	}

	// other error
	if err != nil {
		return errors.Wrap(err, "do.redis.get")
	}

	// get data from redis
	if err := json.Unmarshal(data, &p); err != nil {
		return errors.Wrap(err, "json.Unmarshal.redisData")
	}

	return p.Get(out)
}

func (c *RedisCache) SetWithTTL(key string, value any, ttl time.Duration) error {
	p := payload{Content: value}
	data, err := json.Marshal(p)
	if err != nil {
		return errors.Wrap(err, "json.Marshal.redisData")
	}

	key = c.packKey(key)

	if ttl > 0 {
		_, err = c.RedisClient.Do("SET", key, data, "EX", int(ttl.Seconds()))
	} else {
		_, err = c.RedisClient.Do("SET", key, data)
	}

	return errors.Wrap(err, "do.redis.set")
}

func (c *RedisCache) Close() error {
	return c.RedisClient.Close()
}
