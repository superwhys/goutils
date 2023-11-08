package slices

import "errors"

type StringSet map[string]struct{}

func NewStringSet(slices ...[]string) StringSet {
	ret := StringSet{}
	for _, slice := range slices {
		for _, s := range slice {
			ret.Add(s)
		}
	}
	return ret
}

// Contains return true if given string is in the set.
func (ss StringSet) Contains(s string) bool {
	if ss == nil {
		return false
	}
	_, hit := ss[s]
	return hit
}

// Add pushs string into set. If already in, it's a no-op.
func (ss StringSet) Add(strs ...string) error {
	if ss == nil {
		return errors.New("String Set not initialized")
	}
	for _, s := range strs {
		ss[s] = struct{}{}
	}
	return nil
}

// Merge adds all content in another string set to the current set.
func (ss StringSet) Merge(obj StringSet) error {
	if ss == nil {
		return errors.New("String Set not initialized")
	}
	for s := range obj {
		ss.Add(s)
	}
	return nil
}

// Exclude removes content in another string set from the current set.
func (ss StringSet) Exclude(obj StringSet) error {
	if ss == nil {
		return nil
	}
	for s := range obj {
		ss.Delete(s)
	}
	return nil
}

// Delete remove string from set. If not in, it's a no-op.
func (ss StringSet) Delete(strs ...string) error {
	if ss == nil {
		return nil
	}
	for _, s := range strs {
		delete(ss, s)
	}
	return nil
}

// Length of the set elements number.
func (ss StringSet) Length() int {
	if ss == nil {
		return 0
	}
	return len(ss)
}

// Slice returns a slice of elements.
func (ss StringSet) Slice() []string {
	if ss == nil {
		return nil
	}
	var ret []string
	for s := range ss {
		ret = append(ret, s)
	}
	return ret
}

// SliceContainString returns true if given string is in the slice.
func SliceContainString(s []string, s1 string) bool {
	for _, item := range s {
		if item == s1 {
			return true
		}
	}
	return false
}
