package timeutils

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestFirstDayOfMonth(t *testing.T) {
	type args struct {
		date time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"test1", args{time.Date(2024, 1, 20, 22, 30, 0, 0, time.Now().Location())}, time.Date(2024, 1, 1, 0, 0, 0, 0, time.Now().Location())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FirstDayOfMonth(tt.args.date)
			fmt.Printf("FirstDayOfMonth %v: %v\n", tt.args.date, got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FirstDayOfMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastDayOfMonth(t *testing.T) {
	type args struct {
		date time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"test1", args{time.Date(2024, 1, 20, 22, 30, 0, 0, time.Now().Location())}, time.Date(2024, 1, 31, 23, 59, 59, 0, time.Now().Location())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LastDayOfMonth(tt.args.date)
			fmt.Printf("LastDayOfMonth %v: %v\n", tt.args.date, got)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LastDayOfMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirstDayOfWeek(t *testing.T) {
	type args struct {
		date time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"test1", args{time.Date(2024, 1, 20, 22, 30, 0, 0, time.Now().Location())}, time.Date(2024, 1, 15, 0, 0, 0, 0, time.Now().Location())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FirstDayOfWeek(tt.args.date); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FirstDayOfWeek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastDayOfWeek(t *testing.T) {
	type args struct {
		date time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"test1", args{time.Date(2024, 1, 20, 22, 30, 0, 0, time.Now().Location())}, time.Date(2024, 1, 21, 0, 0, 0, 0, time.Now().Location())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastDayOfWeek(tt.args.date); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LastDayOfWeek() = %v, want %v", got, tt.want)
			}
		})
	}
}
