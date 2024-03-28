package httputils

import (
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
)

type Response struct {
	*http.Response
	respByte []byte
	err      error
}

func (r *Response) Error() error {
	return r.err
}

func (r *Response) BodyString() (string, error) {
	bytes, err := r.BodyBytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *Response) BodyBytes() ([]byte, error) {
	if r.err != nil {
		return []byte{}, r.err
	}

	if r.respByte != nil {
		return r.respByte, nil
	}

	defer r.Response.Body.Close()
	bytes, err := ioutil.ReadAll(r.Response.Body)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil

}

func (r *Response) BodyJson(v interface{}) error {
	if r.err != nil {
		return r.err
	}

	if v == nil {
		return errors.New("value is nil")
	}

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("value is not ptr")
	}

	bytes, err := r.BodyBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}
