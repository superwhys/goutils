package service

import "github.com/superwhys/goutils/service/finder"

func DiscoverServiceWithTag(service, tag string) string {
	return finder.GetServiceFinder().GetAddressWithTag(service, tag)
}

func DiscoverService(service string) string {
	return DiscoverServiceWithTag(service, "")
}
