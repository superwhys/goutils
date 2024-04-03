package redisutils

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
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
	conn := rc.GetConn()
	defer conn.Close()

	_, err = rc.Do("DEL", key)
	return
}
