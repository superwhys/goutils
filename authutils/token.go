package authutils

import (
	"fmt"
	"time"

	"github.com/superwhys/goutils/lg"
)

const (
	defaultTTL = 3600 * time.Second
)

type TokenStorager interface {
	SetWithTTL(key string, value any, ttl time.Duration) error
	Get(key string, out any) error
	Delete(key string) error
}

type Token interface {
	GetKey() string
	SetKey(string)
}

type TokenManager struct {
	storager    TokenStorager
	cachePrefix string
	cacheTTL    time.Duration
}

type TokenManagerOption func(*TokenManager)

func WithCachePrefix(prefix string) TokenManagerOption {
	return func(tm *TokenManager) {
		tm.cachePrefix = prefix
	}
}

func WithCacheTTL(ttl time.Duration) TokenManagerOption {
	return func(tm *TokenManager) {
		tm.cacheTTL = ttl
	}
}

func NewTokenManager(storager TokenStorager, opts ...TokenManagerOption) *TokenManager {
	tm := &TokenManager{
		storager: storager,
		cacheTTL: defaultTTL,
	}

	for _, opt := range opts {
		opt(tm)
	}

	return tm
}

func (tm *TokenManager) getKey(t Token) string {
	key := tm.cachePrefix
	if key == "" {
		key = lg.StructName(t)
	}

	return fmt.Sprintf("%v:%v", key, t.GetKey())
}

func (tm *TokenManager) Save(t Token) error {
	return tm.storager.SetWithTTL(tm.getKey(t), t, tm.cacheTTL)
}

func (tm *TokenManager) Read(t Token) error {
	return tm.storager.Get(tm.getKey(t), t)
}

func (tm *TokenManager) Remove(t Token) error {
	return tm.storager.Delete(tm.getKey(t))
}
