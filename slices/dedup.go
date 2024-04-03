package slices

import (
	"errors"
	"reflect"
)

func DeDup(slice interface{}, keyF func(idx int) string) (interface{}, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil, errors.New("not slice")
	}

	m := map[string]struct{}{}
	idx := 0

	for i := 0; i < v.Len(); i++ {
		key := keyF(i)
		if _, hit := m[key]; hit {
			continue
		} else {
			m[key] = struct{}{}
			v.Index(idx).Set(v.Index(i))
			idx++
		}
	}
	return v.Slice(0, idx).Interface(), nil
}

// DupStrings remove duplicated strings in slice.
// Note the given slice will be modified.
func DupStrings(slice []string) []string {
	if len(slice) <= 50 {
		return DupSliceSmall(slice)
	}
	return DupSliceLarge(slice)
}

func DupInt(slice []int) []int {
	if len(slice) <= 50 {
		return DupSliceIntSmall(slice)
	}
	return DupSliceIntLarge(slice)
}

// DupSliceSmall is the faster version of DupStrings with O(n^2) algorithm.
// For n < 50, it has better performance and zero allocation.
func DupSliceSmall(slice []string) []string {
	idx := 0
	for _, s := range slice {
		var j int
		for j = 0; j < idx; j++ {
			if slice[j] == s {
				break
			}
		}
		if j >= idx {
			slice[idx] = s
			idx++
		}
	}
	return slice[:idx]
}

// DupSliceLarge is the hashmap version of DupStrings with O(n) algorithm.
func DupSliceLarge(slice []string) []string {
	m := map[string]struct{}{}
	idx := 0
	for i, s := range slice {
		if _, hit := m[s]; hit {
			continue
		} else {
			m[s] = struct{}{}
			slice[idx] = slice[i]
			idx++
		}
	}
	return slice[:idx]
}

// DupSliceIntSmall is the faster version of DupInt32 with O(n^2) algorithm.
// For n < 50, it has better performance and zero allocation.
func DupSliceIntSmall(slice []int) []int {
	idx := 0
	for _, s := range slice {
		var j int
		for j = 0; j < idx; j++ {
			if slice[j] == s {
				break
			}
		}
		if j >= idx {
			slice[idx] = s
			idx++
		}
	}
	return slice[:idx]
}

// DupSliceIntLarge is the hashmap version of DupInt32 with O(n) algorithm.
func DupSliceIntLarge(slice []int) []int {
	m := map[int]struct{}{}
	idx := 0
	for i, s := range slice {
		if _, hit := m[s]; hit {
			continue
		} else {
			m[s] = struct{}{}
			slice[idx] = slice[i]
			idx++
		}
	}
	return slice[:idx]
}
