package httputils

import "net/http"

type Header struct {
	http.Header
}

func NewHeader() *Header {
	header := http.Header{}
	return &Header{Header: header}
}

func (h *Header) Add(key, value string) *Header {
	h.Set(key, value)
	return h
}

func DefaultJsonHeader() *Header {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &Header{Header: header}
}

func DefaultFormUrlEncodedHeader() *Header {
	header := http.Header{}
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	return &Header{Header: header}
}

func DefaultFormHeader() *Header {
	header := http.Header{}
	header.Add("Content-Type", "multipart/form-data")
	return &Header{Header: header}
}
