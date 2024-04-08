package ginutils

import (
	"context"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	redisgo "github.com/gomodule/redigo/redis"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/dialer"
)

type RedisStoreOptions struct {
	user     string
	password string
	db       int
	keyPairs [][]byte
}

type RedisStoreOptionFunc func(o *RedisStoreOptions)

func WithUser(user string) RedisStoreOptionFunc {
	return func(o *RedisStoreOptions) {
		o.user = user
	}
}

func WithPassword(password string) RedisStoreOptionFunc {
	return func(o *RedisStoreOptions) {
		o.password = password
	}
}

func WithDb(db int) RedisStoreOptionFunc {
	return func(o *RedisStoreOptions) {
		o.db = db
	}
}

func WithKeyPairs(keyPairs ...string) RedisStoreOptionFunc {
	return func(o *RedisStoreOptions) {
		var bs [][]byte
		for _, kp := range keyPairs {
			bs = append(bs, []byte(kp))
		}
		o.keyPairs = append(o.keyPairs, bs...)
	}
}

func NewRedisSessionStore(service string, opts ...RedisStoreOptionFunc) (redis.Store, error) {
	opt := &RedisStoreOptions{}

	for _, o := range opts {
		o(opt)
	}

	redisPool := dialer.DialRedisPool(service, opt.db, 100, opt.password)
	return NewRedisSessionStoreWithRedisPool(redisPool, opt.keyPairs)
}

func NewRedisSessionStoreWithRedisPool(pool *redisgo.Pool, keyPairs [][]byte) (redis.Store, error) {
	return redis.NewStoreWithPool(pool, keyPairs...)
}

var (
	ErrorTokenNotFound = errors.New("Token not found!")
)

type Token interface {
	GetKey() string
	Marshal() (string, error)
	UnMarshal(val string) error
}

func SetSessionToken(c *gin.Context, t Token) error {
	session := sessions.Default(c)

	s, err := t.Marshal()
	if err != nil {
		return errors.Wrap(err, "tokenMarshal")
	}
	session.Set(t.GetKey(), s)
	if err := session.Save(); err != nil {
		return errors.Wrap(err, "saveSession")
	}

	return nil
}

func GetSessionToken(c *gin.Context, t Token) error {
	session := sessions.Default(c)

	val := session.Get(t.GetKey())
	if val == nil {
		return ErrorTokenNotFound
	}

	tokenStr, ok := val.(string)
	if !ok {
		return ErrorTokenNotFound
	}

	if err := t.UnMarshal(tokenStr); err != nil {
		return errors.Wrap(err, "decode token")
	}
	return nil
}

type StringToken string

func (st StringToken) GetKey() string {
	return string(st)
}

func (st StringToken) Marshal() (string, error) {
	return string(st), nil
}

func (st *StringToken) UnMarshal(val string) error {
	*st = StringToken(val)
	return nil
}

func NewSessionMiddleware(key string, store sessions.Store) gin.HandlerFunc {
	return sessions.Sessions(key, store)
}

type SessionMiddlewareHandler struct {
	sessionKey string
	store      sessions.Store
}

func NewSessionMiddlewareHandler(key string, store sessions.Store) *SessionMiddlewareHandler {
	return &SessionMiddlewareHandler{
		sessionKey: key,
		store:      store,
	}
}

func (sh *SessionMiddlewareHandler) HandleFunc(ctx context.Context, c *gin.Context) HandleResponse {
	sessions.Sessions(sh.sessionKey, sh.store)(c)
	return nil
}
