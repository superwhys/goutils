package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var _ CacheWithTTL = (*MemoryCache)(nil)

type payloadWithExpire struct {
	payload  payload
	expireAt time.Time
}

func (p *payloadWithExpire) IsExpire() bool {
	if p == nil {
		return false
	}

	if p.expireAt.IsZero() {
		return false
	}

	return !time.Now().Before(p.expireAt)
}

func (p *payloadWithExpire) Get(out any) error {
	return p.payload.Get(out)
}

type MemoryCache struct {
	lock             sync.RWMutex
	payload          map[string]payloadWithExpire
	cancel           func()
	rotationInterval time.Duration
}

func NewMemoryCache(rotationInterval time.Duration) *MemoryCache {
	mc := &MemoryCache{
		payload:          make(map[string]payloadWithExpire),
		rotationInterval: rotationInterval,
	}
	ctx, cancel := context.WithCancel(context.TODO())
	mc.cancel = cancel
	go mc.runRotation(ctx, rotationInterval)
	return mc
}

func (mc *MemoryCache) runRotation(ctx context.Context, rotationInterval time.Duration) {
	ticker := time.NewTicker(rotationInterval)
	defer ticker.Stop()

	for range ticker.C {
		if ctx.Err() != nil {
			return
		}
		mc.lock.Lock()
		for key, value := range mc.payload {
			if value.IsExpire() {
				delete(mc.payload, key)
			}
		}
		mc.lock.Unlock()
	}
}

func (mc *MemoryCache) Get(key string, out any) error {
	mc.lock.RLock()
	p, ok := mc.payload[key]
	mc.lock.RUnlock()

	if !ok || !p.IsExpire() {
		return errors.New("not found")
	}

	return p.Get(out)
}

func (mc *MemoryCache) Set(key string, value any) error {
	return mc.SetWithTTL(key, 0, value)
}

func (mc *MemoryCache) GetOrCreate(key string, creater Creater, out any) error {
	return mc.GetOrCreateWithTTL(key, 0, creater, out)
}

func (mc *MemoryCache) Delete(key string) error {
	mc.lock.Lock()
	_, ok := mc.payload[key]
	if ok {
		delete(mc.payload, key)
	}
	mc.lock.Unlock()
	return nil
}

func (mc *MemoryCache) Close() error {
	mc.cancel()
	mc.lock.Lock()
	mc.payload = nil
	mc.lock.Unlock()
	return nil
}

func (mc *MemoryCache) GetOrCreateWithTTL(key string, ttl time.Duration, creator Creater, out any) error {
	mc.lock.RLock()
	p, ok := mc.payload[key]
	mc.lock.RUnlock()

	if ok && !p.IsExpire() {
		return p.Get(out)
	}

	np := newPayload(creator())
	mc.lock.Lock()

	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}

	mc.payload[key] = payloadWithExpire{
		payload:  np,
		expireAt: expireAt,
	}
	mc.lock.Unlock()
	return np.Get(out)
}

func (mc *MemoryCache) SetWithTTL(key string, ttl time.Duration, value any) error {
	p := payload{Content: value}

	mc.lock.Lock()

	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}

	mc.payload[key] = payloadWithExpire{
		payload:  p,
		expireAt: expireAt,
	}

	mc.lock.Unlock()

	return nil
}
