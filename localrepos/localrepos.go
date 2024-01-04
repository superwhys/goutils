package localrepos

import (
	"context"
	"sync"
	"time"

	"github.com/superwhys/goutils/lg"
)

type HashData interface {
	GetId() string
}

type DataStore interface {
	ReloadEntries(ctx context.Context, ch chan HashData) error
	Close() error
}

type ReposKVEntry struct {
	Key   string
	Value HashData
}

type LocalRepos struct {
	sync.RWMutex

	parentCtx context.Context

	dataStore DataStore
	data      map[string]HashData

	refreshInterval time.Duration
	ticker          *time.Ticker
}

type ReposOptions func(*LocalRepos)

var defaultRefreshInterval = 5 * time.Minute

func WithRefreshInterval(interval time.Duration) ReposOptions {
	return func(r *LocalRepos) {
		r.refreshInterval = interval
	}
}

func NewLocalRepos(dataStore DataStore, opts ...ReposOptions) *LocalRepos {
	lp := &LocalRepos{
		parentCtx: context.Background(),
		dataStore: dataStore,
	}

	for _, opt := range opts {
		opt(lp)
	}
	return lp
}

func (r *LocalRepos) Start() {
	r.reloadEntries(r.parentCtx)

	r.ticker = time.NewTicker(r.refreshInterval)
	go func() {
		for range r.ticker.C {
			r.reloadEntries(r.parentCtx)
		}
	}()
}

func (r *LocalRepos) reloadEntries(ctx context.Context) {
	ch := make(chan HashData)
	go func() {
		newData := make(map[string]HashData)
		for {
			select {
			case <-ctx.Done():
				lg.Errorf("reloadEntries done: %v", ctx.Err())
				return
			case data, ok := <-ch:
				if !ok {
					lg.Debugc(ctx, "reloadEntries chan closed")
					goto Done
				}
				newData[data.GetId()] = data
			default:
			}
		}

	Done:
		r.Lock()
		r.data = newData
		r.Unlock()
	}()

	if err := r.dataStore.ReloadEntries(ctx, ch); err != nil {
		lg.Errorc(ctx, "reloadEntries error: %v", err)
	}
}

func (r *LocalRepos) AllValues() []HashData {
	r.RLock()
	defer r.RUnlock()
	var ret []HashData
	for _, v := range r.data {
		ret = append(ret, v)
	}
	return ret
}

func (r *LocalRepos) AllKeys() []string {
	r.RLock()
	defer r.RUnlock()
	var ret []string
	for k := range r.data {
		ret = append(ret, k)
	}
	return ret
}

func (r *LocalRepos) AllItems() []*ReposKVEntry {
	r.RLock()
	defer r.RUnlock()
	var ret []*ReposKVEntry
	for k, v := range r.data {
		ret = append(ret, &ReposKVEntry{k, v})
	}
	return ret
}

func (r *LocalRepos) Get(id string) HashData {
	r.RLock()
	defer r.RUnlock()
	return r.data[id]
}

func (r *LocalRepos) Len() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.data)
}

func (r *LocalRepos) Close() error {
	r.ticker.Stop()
	return r.dataStore.Close()
}
