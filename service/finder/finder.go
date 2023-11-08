package finder

import (
	"sync"

	"github.com/superwhys/goutils/service/finder/consul"
)

type Service struct {
	ServiceName string
	Address     string
	Tags        string
}

type ServiceFinder interface {
	GetAddress(service string) string
	GetAllAddress(service string) []string
	GetAddressWithTag(service, tag string) string
	GetAllAddressWithTag(service, tag string) []string

	RegisterService(service, address string) error
	RegisterServiceWithTag(service, address, tag string) error
	Close()
}

var (
	defaultServiceFinder ServiceFinder
	finderMutex          sync.RWMutex
)

func init() {
	defaultServiceFinder = GetServiceFinder()
}

func GetServiceFinder() ServiceFinder {
	finderMutex.RLock()
	defer finderMutex.RUnlock()
	return defaultServiceFinder
}

func SetConsulFinderToDefault() {
	finderMutex.Lock()
	defer finderMutex.Unlock()
	defaultServiceFinder = GetConsulServiceFinder()
}

func GetConsulServiceFinder() ServiceFinder {
	return consul.GetConsulClient()
}
