package slices

import (
	"reflect"
	"testing"
)

func TestReverse(t *testing.T) {
	type args struct {
		slice []interface{}
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{name: "reverse-1", args: args{slice: []interface{}{1, 2, 3, 4, 5}}, want: []interface{}{5, 4, 3, 2, 1}},
		{name: "reverse-2", args: args{slice: []interface{}{1, 2, 3, 4, 5, 6}}, want: []interface{}{6, 5, 4, 3, 2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reverse(tt.args.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}
