package redisutils

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

const (
	defaultTTL = time.Second * 1000
)

type RedisClient struct {
	pool *redis.Pool
	ttl  time.Duration
}

func NewRedisClient(pool *redis.Pool) *RedisClient {
	return &RedisClient{
		pool: pool,
	}
}

func (rc *RedisClient) GetConn() redis.Conn {
	conn, _ := rc.GetConnWithContext(context.TODO())
	return conn
}

func (rc *RedisClient) GetConnWithContext(ctx context.Context) (redis.Conn, error) {
	return rc.pool.GetContext(ctx)
}

func (rc *RedisClient) Do(command string, args ...any) (reply any, err error) {
	conn := rc.GetConn()
	defer conn.Close()

	return conn.Do(command, args...)
}

func (rc *RedisClient) SetWithTTL(key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "encode")
	}

	if ttl > 0 {
		_, err = rc.Do("SET", key, data, "EX", int(ttl.Seconds()))
	} else {
		_, err = rc.Do("SET", key, data)
	}

	return errors.Wrap(err, "redis.Set")
}

func (rc *RedisClient) Set(key string, value any) error {
	return rc.SetWithTTL(key, value, 0)
}

func (rc *RedisClient) Get(key string, out any) error {
	data, err := redis.Bytes(rc.Do("GET", key))
	if err != nil {
		return errors.Wrap(err, "redis.GET")
	}

	if err := json.Unmarshal(data, &out); err != nil {
		return errors.Wrap(err, "decode")
	}

	return nil
}

func (rc *RedisClient) Delete(key string) error {
	_, err := rc.Do("DEL", key)
	return err
}

func (rc *RedisClient) Lock(key string, expires ...time.Duration) (err error) {
	expire := defaultTTL
	if len(expires) != 0 {
		expire = expires[0]
	}

	_, err = redis.String(rc.Do("SET", key, time.Now().Unix(), "EX", int(expire.Seconds()), "NX"))
	if err == redis.ErrNil {
		return
	}

	if err != nil {
		return err
	}

	return nil
}

func (rc *RedisClient) UnLock(key string) (err error) {
	return rc.Delete(key)
}

func (rc *RedisClient) Close() error {
	return rc.pool.Close()
}
