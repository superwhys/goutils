package lg

import (
	"reflect"
	"runtime"
)

func FuncName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func StructName(i interface{}) string {
	tPtr := reflect.TypeOf(i)

	if tPtr.Kind() == reflect.Ptr {
		tPtr = tPtr.Elem()
	}

	return tPtr.Name()
}
