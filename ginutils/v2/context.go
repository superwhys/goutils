package ginutils

import (
	"encoding/json"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-viper/mapstructure/v2"
	"github.com/pkg/errors"
)

type HandleResponse interface {
	GetCode() int
	GetError() error
	GetData() any
	GetMessage() string
}

type Ret struct {
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Err     error  `json:"-"`
}

func (r *Ret) GetCode() int {
	return r.Code
}

func (r *Ret) GetError() error {
	return r.Err
}

func (r *Ret) GetData() any {
	return r
}

func (r *Ret) GetMessage() string {
	return r.Message
}

func (r *Ret) PackContent(code int, data any, err error, message string) *Ret {
	r.Err = err
	r.Message = message
	r.Code = code
	r.Data = data
	return r
}

func AbortWithError(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, Ret{
		Code:    code,
		Message: message,
	})
	c.Error(errors.New(message))
}

func StatusOk(c *gin.Context, data any) {
	ReturnWithStatus(c, 200, data)
}

func ReturnWithStatus(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

const (
	paramsKey = "ginutils:paramsKey"
)

func ParseMapParams(c *gin.Context) (map[string]any, error) {
	params := make(map[string]any)

	bodyBytes, err := c.GetRawData()
	if err != nil {
		return nil, err
	}

	var bodyMap map[string]interface{}
	if json.Unmarshal(bodyBytes, &bodyMap) == nil {
		for k, v := range bodyMap {
			params[k] = v
		}
	}

	for _, p := range c.Params {
		params[p.Key] = p.Value
	}

	queryMap := c.Request.URL.Query()
	for k, v := range queryMap {
		params[k] = v[0]
	}

	return params, nil
}

func GetParams(c *gin.Context) (map[string]any, bool) {
	val, exists := c.Get(paramsKey)
	if !exists {
		return nil, false
	}

	params := val.(map[string]any)
	return params, true
}

func BindParms(c *gin.Context, data any) error {
	if dt := reflect.TypeOf(data); dt.Kind() != reflect.Pointer {
		return errors.New("data need a struct pointer")
	}

	params, exists := GetParams(c)
	if !exists {
		return errors.New("no params to bind")
	}

	config := &mapstructure.DecoderConfig{
		Result:  data,
		TagName: "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return errors.Wrap(err, "new decoder")
	}

	if err := decoder.Decode(params); err != nil {
		return errors.Wrap(err, "deocde parasm")
	}

	return nil
}
