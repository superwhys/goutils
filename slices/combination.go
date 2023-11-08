package slices

import (
	"fmt"
	"math/rand"
	"reflect"
)

func CombinationsSlice(list interface{}, n int) [][]interface{} {
	listValue := reflect.ValueOf(list)
	if listValue.Kind() != reflect.Slice {
		panic("Input must be a slice")
	}

	if n == 1 {
		result := make([][]interface{}, 0)
		for i := 0; i < listValue.Len(); i++ {
			result = append(result, []interface{}{listValue.Index(i).Interface()})
		}
		return result
	}

	result := make([][]interface{}, 0)
	for i := 0; i < listValue.Len()-n+1; i++ {
		for _, c := range CombinationsSlice(listValue.Slice(i+1, listValue.Len()).Interface(), n-1) {
			result = append(result, append([]interface{}{listValue.Index(i).Interface()}, c...))
		}
	}
	return result
}

func RandomSelect(source interface{}, count int) (interface{}, error) {
	sliceValue := reflect.ValueOf(source)
	if sliceValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("Input must be a slice")
	}

	n := sliceValue.Len()

	shuffled := reflect.MakeSlice(sliceValue.Type(), n, n)
	indexes := rand.Perm(n)

	for i, j := range indexes {
		shuffled.Index(i).Set(sliceValue.Index(j))
	}

	if count > n {
		// the count is greater than the length of the slice
		// so we need to append some random elements
		for i := 0; i < count-n; i++ {
			shuffled = reflect.Append(shuffled, sliceValue.Index(rand.Intn(n)))
		}
	}

	return shuffled.Slice(0, count).Interface(), nil
}
