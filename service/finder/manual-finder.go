package finder

import "sync"

type ManualFinder struct {
	// serviceMap key is service name, value is a list of service struct
	serviceMap map[string][]*Service
	lock       sync.RWMutex
}

func NewManualFinder() *ManualFinder {
	return &ManualFinder{
		serviceMap: make(map[string][]*Service),
	}
}

func (mf *ManualFinder) GetAddress(service string) string {
	addresses := mf.GetAllAddress(service)
	if len(addresses) > 0 {
		return addresses[0]
	}

	return ""
}

func (mf *ManualFinder) GetAllAddress(service string) []string {
	mf.lock.RLock()
	defer mf.lock.RUnlock()

	var ret []string

	if services, ok := mf.serviceMap[service]; ok {
		for _, s := range services {
			ret = append(ret, s.Address)
		}
		return ret
	}

	return nil
}

func (mf *ManualFinder) GetAddressWithTag(service string, tag string) string {
	addresses := mf.GetAllAddressWithTag(service, tag)
	if len(addresses) > 0 {
		return addresses[0]
	}

	return ""
}

func (mf *ManualFinder) GetAllAddressWithTag(service string, tag string) []string {
	mf.lock.RLock()
	defer mf.lock.RUnlock()

	var ret []string

	if services, ok := mf.serviceMap[service]; ok {
		for _, s := range services {
			if s.Tags == tag {
				ret = append(ret, s.Address)
			}
		}
		return ret
	}

	return nil
}

func (mf *ManualFinder) RegisterService(service string, address string) error {
	return mf.RegisterServiceWithTag(service, address, "")
}

func (mf *ManualFinder) RegisterServiceWithTag(service string, address string, tag string) error {
	mf.lock.Lock()
	defer mf.lock.Unlock()

	// check if service already registered
	if services, ok := mf.serviceMap[service]; ok {
		for _, s := range services {
			if s.Address == address && s.Tags == tag {
				return nil
			}
		}
		mf.serviceMap[service] = append(mf.serviceMap[service], &Service{
			ServiceName: service,
			Address:     address,
			Tags:        tag,
		})
		return nil
	}
	mf.serviceMap[service] = []*Service{
		{
			ServiceName: service,
			Address:     address,
			Tags:        tag,
		},
	}

	return nil
}

func (mf *ManualFinder) Close() {
	// do nothing
}
