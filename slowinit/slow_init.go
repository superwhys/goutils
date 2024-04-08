package slowinit

import (
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/lg"
)

type slowerObject struct {
	name string
	fn   func() error
}

var (
	objs = make([]*slowerObject, 0)
)

func RegisterObject(name string, fn func() error) {
	objs = append(objs, &slowerObject{
		name: name,
		fn:   fn,
	})
}

func Init() {
	for _, obj := range objs {
		if err := obj.fn(); err != nil {
			lg.PanicError(errors.Wrapf(err, "slower init obj: %v", obj.name))
		}
		lg.Debugf("slower init obj: %v success!", obj.name)
	}
}
