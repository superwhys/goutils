package slices

import (
	"testing"
)

func TestStringSet_Contains(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		ss   StringSet
		args args
		want bool
	}{
		{name: "contains-true", ss: NewStringSet([]string{"a", "b", "c"}), args: args{"a"}, want: true},
		{name: "contains-false", ss: NewStringSet([]string{"a", "b", "c"}), args: args{"d"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ss.Contains(tt.args.s); got != tt.want {
				t.Errorf("StringSet.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSet_Add(t *testing.T) {
	type args struct {
		strs []string
	}
	tests := []struct {
		name    string
		ss      StringSet
		args    args
		wantErr bool
	}{
		{name: "add-no-error", ss: NewStringSet([]string{"a", "b", "c"}), args: args{[]string{"d"}}, wantErr: false},
		{name: "add-has-error", ss: nil, args: args{[]string{"d"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ss.Add(tt.args.strs...)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("StringSet.Add() error = %v, wantErr %v", err, tt.wantErr)
				} else {
					return
				}
			}

			for _, arg := range tt.args.strs {
				if tt.ss.Contains(arg) != true {
					t.Errorf("StringSet.Add(%v), err: not add success", arg)
				}
			}
		})
	}
}

func TestStringSet_Merge(t *testing.T) {
	type args struct {
		obj StringSet
	}
	tests := []struct {
		name        string
		ss          StringSet
		args        args
		finalSlices []string
		wantErr     bool
	}{
		// TODO: Add test cases.
		{name: "merge-no-error-1", ss: NewStringSet([]string{"a", "b", "c"}), args: args{NewStringSet([]string{"a", "b"})}, finalSlices: []string{"a", "b", "c"}, wantErr: false},
		{name: "merge-no-error-2", ss: NewStringSet([]string{"a", "b", "c"}), args: args{NewStringSet([]string{"d", "e"})}, finalSlices: []string{"a", "b", "c", "d", "e"}, wantErr: false},
		{name: "merge-has-error", ss: nil, args: args{NewStringSet([]string{"d", "e"})}, finalSlices: []string{"a", "b", "c", "d", "e"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ss.Merge(tt.args.obj)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("StringSet.Merge() error = %v, wantErr %v", err, tt.wantErr)
				} else {
					return
				}
			}

			for _, arg := range tt.finalSlices {
				if tt.ss.Contains(arg) != true {
					t.Errorf("StringSet.Merge(%v), err: not merge success", arg)
				}
			}
		})
	}
}

func TestStringSet_Exclude(t *testing.T) {
	type args struct {
		obj StringSet
	}
	tests := []struct {
		name        string
		ss          StringSet
		args        args
		finalSlices []string
		wantErr     bool
	}{
		{name: "exclude-no-error-1", ss: NewStringSet([]string{"a", "b", "c"}), args: args{NewStringSet([]string{"a", "b"})}, finalSlices: []string{"c"}, wantErr: false},
		{name: "exclude-no-error-2", ss: NewStringSet([]string{"a", "b", "c"}), args: args{NewStringSet([]string{"a", "b", "c"})}, finalSlices: []string{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ss.Exclude(tt.args.obj)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("StringSet.Exclude() error = %v, wantErr %v", err, tt.wantErr)
				} else {
					return
				}
			}

			for _, arg := range tt.finalSlices {
				if tt.ss.Contains(arg) != true {
					t.Errorf("StringSet.Exclude(%v), err: not exclude success", arg)
				}
			}
		})
	}
}

func TestStringSet_Delete(t *testing.T) {
	type args struct {
		strs []string
	}
	tests := []struct {
		name        string
		ss          StringSet
		args        args
		finalSlices []string
		wantErr     bool
	}{
		{name: "delete-no-error-1", ss: NewStringSet([]string{"a", "b", "c"}), args: args{[]string{"a", "b"}}, finalSlices: []string{"c"}, wantErr: false},
		{name: "delete-no-error-2", ss: NewStringSet([]string{"a", "b", "c"}), args: args{[]string{"a", "b", "c"}}, finalSlices: []string{}, wantErr: false},
		{name: "delete-has-error", ss: nil, args: args{[]string{"a", "b", "c"}}, finalSlices: []string{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ss.Delete(tt.args.strs...)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("StringSet.Delete() error = %v, wantErr %v", err, tt.wantErr)
				} else {
					return
				}
			}

			for _, arg := range tt.finalSlices {
				if tt.ss.Contains(arg) != true {
					t.Errorf("StringSet.Delete(%v), err: not delete success", arg)
				}
			}
		})
	}
}

func TestStringSet_Length(t *testing.T) {
	tests := []struct {
		name string
		ss   StringSet
		want int
	}{
		{name: "length-no-error-1", ss: NewStringSet([]string{"a", "b", "c"}), want: 3},
		{name: "length-no-error-2", ss: NewStringSet([]string{"a", "b", "c", "a"}), want: 3},
		{name: "length-no-error-3", ss: NewStringSet([]string{}), want: 0},
		{name: "length-has-error", ss: nil, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ss.Length(); got != tt.want {
				t.Errorf("StringSet.Length() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSet_Slice(t *testing.T) {
	tests := []struct {
		name string
		ss   StringSet
		want []string
	}{
		{name: "slice-no-error-1", ss: NewStringSet([]string{"a", "b", "c"}), want: []string{"a", "b", "c"}},
		{name: "slice-no-error-2", ss: NewStringSet([]string{"a", "b", "c", "a"}), want: []string{"a", "b", "c"}},
		{name: "slice-no-error-3", ss: NewStringSet([]string{}), want: []string{}},
		{name: "slice-has-error", ss: nil, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ss.Slice()
			for _, v := range got {
				if SliceContainString(tt.want, v) != true {
					t.Errorf("StringSet.Slice() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSliceContainString(t *testing.T) {
	type args struct {
		s  []string
		s1 string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "slice-contain-string-no-error-1", args: args{s: []string{"a", "b", "c"}, s1: "a"}, want: true},
		{name: "slice-contain-string-no-error-2", args: args{s: []string{"a", "b", "c"}, s1: "d"}, want: false},
		{name: "slice-contain-string-no-error-3", args: args{s: []string{}, s1: "a"}, want: false},
		{name: "slice-contain-string-no-error-4", args: args{s: nil, s1: "a"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SliceContainString(tt.args.s, tt.args.s1); got != tt.want {
				t.Errorf("SliceContainString() = %v, want %v", got, tt.want)
			}
		})
	}
}
