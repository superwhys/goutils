package httputils

import "time"

type Config struct {
	RequestTimeOut time.Duration
	Proxy          string
}
