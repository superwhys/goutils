package slices

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCombinationsSlice(t *testing.T) {
	type args struct {
		list interface{}
		n    int
	}
	tests := []struct {
		name string
		args args
		want [][]interface{}
	}{
		{"comb-2", args{
			list: []string{"a", "b", "c"},
			n:    2,
		}, [][]interface{}{
			{"a", "b"},
			{"a", "c"},
			{"b", "c"},
		}},
		{"comb-3", args{
			list: []string{"a", "b", "c", "d"},
			n:    3,
		}, [][]interface{}{
			{"a", "b", "c"},
			{"a", "b", "d"},
			{"a", "c", "d"},
			{"b", "c", "d"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CombinationsSlice(tt.args.list, tt.args.n)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CombinationsSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomSelect(t *testing.T) {
	type args struct {
		source interface{}
		count  int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"random-select-1", args{
			source: []string{"a", "b", "c", "d"},
			count:  1,
		}, false},
		{"random-select-2", args{
			source: []string{"a", "b", "c", "d"},
			count:  2,
		}, false},
		{"random-select-5", args{
			source: []string{"a", "b", "c", "d"},
			count:  5,
		}, false},
		{"random-select-error", args{
			source: "1234",
			count:  5,
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomSelect(tt.args.source, tt.args.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomSelect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)

		})
	}
}
