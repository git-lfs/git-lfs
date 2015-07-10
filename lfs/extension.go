package lfs

import (
	"fmt"
	"sort"
)

// An Extension describes how to manipulate files during smudge and clean.
type Extension struct {
	Name     string
	Clean    string
	Smudge   string
	Priority int
}

// SortExtensions sorts a map of extensions in ascending order by Priority
func SortExtensions(m map[string]Extension) ([]Extension, error) {
	pMap := make(map[int]Extension)
	for n, ext := range m {
		p := ext.Priority
		if _, exist := pMap[p]; exist {
			err := fmt.Errorf("duplicate priority %d on %s", p, n)
			return nil, err
		}
		pMap[p] = ext
	}

	var priorities []int
	for p := range pMap {
		priorities = append(priorities, p)
	}

	sort.Ints(priorities)

	var result []Extension
	for _, p := range priorities {
		result = append(result, pMap[p])
	}

	return result, nil
}
