package cache

import (
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type Creater func() (any, error)

type Cache interface {
	Get(key string, out any) error
	Set(key string, value any) error
	GetOrCreate(key string, creater Creater, out any) error
	Delete(key string)
	Close() error
}

type CacheWithTTL interface {
	Cache
	GetOrCreateWithTTL(key string, ttl time.Duration, creator Creater, out any) error
	SetWithTTL(key string, ttl time.Duration, value any) error
}

type payload struct {
	Content any   `json:"content"`
	Error   error `json:"error"`
}

func (p payload) Get(out any) error {
	if p.Error != nil {
		return p.Error
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  out,
		TagName: "json",
	})
	if err != nil {
		return errors.Wrap(err, "mapstructure.NewDecoder")
	}

	return decoder.Decode(p.Content)
}

func newPayload(content any, err error) payload {
	if err != nil {
		return payload{Error: err}
	}

	return payload{Content: content}
}
