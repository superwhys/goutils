package cache

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	c *RedisCache
)

type testData struct {
	Message string `json:"message"`
}

func TestMain(m *testing.M) {
	pool := redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", "localhost:6379", redis.DialDatabase(1))
	}, 3)
	c = NewRedisCache(pool)
	m.Run()
}

func TestRedisCache_Set(t *testing.T) {
	type args struct {
		key   string
		value any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"cache_set", args{"test_key", "test_value"}, false},
		{"cache_set_struct", args{"test_struct_key", &testData{Message: "this is message"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("RedisCache.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	type args struct {
		key  string
		out  any
		want any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "cache_get", args: args{"test_key", "", "test_value"}, wantErr: false},
		{name: "cache_get_struct", args: args{"test_struct_key", &testData{}, &testData{Message: "this is message"}}, wantErr: false},
		{name: "cache_get_nil", args: args{"test_nil_key", nil, nil}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Get(tt.args.key, &tt.args.out)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisCache.Get() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.args.out, tt.args.want) {
				t.Errorf("RedisCache.Get() want = %v, get = %v", tt.args.want, tt.args.out)
			}
		})
	}
}

func TestRedisCache_GetOrCreate(t *testing.T) {
	type args struct {
		key     string
		creater Creater
		out     any
		want    any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"test_create_str", args{"test_create_key", func() (any, error) {
				return "create_value", nil
			}, "", "create_value"}, false,
		},
		{
			"test_create_struct", args{"test_create_struct_key", func() (any, error) {
				return &testData{Message: "this is create message"}, nil
			}, &testData{}, &testData{Message: "this is create message"}}, false,
		},
		{
			"test_key_exists", args{"test_key", func() (any, error) {
				return "create_value", nil
			}, "", "test_value"}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.GetOrCreate(tt.args.key, tt.args.creater, &tt.args.out)
			t.Logf("getOrCreate value: %v", tt.args.out)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisCache.GetOrCreate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.args.out, tt.args.want) {
				t.Errorf("RedisCache.GetOrCreate() want = %v, get = %v", tt.args.want, tt.args.out)
			}
		})
	}
}

func TestRedisCache_Delete(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
	}{
		{"test-delete", args{"test_key"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out string
			c.Delete(tt.args.key)
			err := c.Get(tt.args.key, &out)
			if err != nil && !errors.Is(err, redis.ErrNil) {
				t.Errorf("redisCache get key error = %v", err)
			}

			if err == nil {
				t.Error("redisCache key exists")
			}
		})
	}
}

func TestRedisCache_GetOrCreateWithTTL(t *testing.T) {
	type args struct {
		key     string
		ttl     time.Duration
		creator Creater
		out     any
		want    any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"test_not_exists_key", args{"test_create_key_ttl", time.Second * 5, func() (any, error) {
				return "create_value", nil
			}, "", ""}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.GetOrCreateWithTTL(tt.args.key, tt.args.ttl, tt.args.creator, tt.args.out)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisCache.GetOrCreateWithTTL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.args.out, tt.args.want) {
				t.Errorf("RedisCache.GetOrCreate() want = %v, get = %v", tt.args.want, tt.args.out)
			}
		})
	}
}

func TestRedisCache_SetWithTTL(t *testing.T) {
	type args struct {
		key   string
		ttl   time.Duration
		value any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.SetWithTTL(tt.args.key, tt.args.ttl, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("RedisCache.SetWithTTL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
