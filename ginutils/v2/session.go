package ginutils

import (
	"context"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/dialer"
	"github.com/superwhys/goutils/lg"
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
	return redis.NewStoreWithPool(redisPool, opt.keyPairs...)
}

var (
	ErrorTokenNotFound = errors.New("Token not found!")
)

type Token interface {
	GetKey() string
	Marshal() string
	UnMarshal(val string) error
}

func SetToken(c *gin.Context, t Token) {
	session := sessions.Default(c)

	session.Set(t.GetKey(), t.Marshal())
	if err := session.Save(); err != nil {
		lg.Errorf("set token error: %v", err)
	}
}

func GetToken(c *gin.Context, t Token) error {
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

type StringToken struct {
	key string
	val string
}

func (st StringToken) GetKey() string {
	return st.key
}

func (st StringToken) Marshal() string {
	return st.val
}

func (st StringToken) UnMarshal(val *string) error {
	*val = st.val
	return nil
}

type SessionMiddlewareHandler struct {
	sessionKey string
	store      sessions.Store
}

func NewSessionMiddleware(key string, store sessions.Store) *SessionMiddlewareHandler {
	return &SessionMiddlewareHandler{
		sessionKey: key,
		store:      store,
	}
}

func (sh *SessionMiddlewareHandler) HandleFunc(ctx context.Context, c *gin.Context) HandleResponse {
	sessions.Sessions(sh.sessionKey, sh.store)(c)
	return nil
}
