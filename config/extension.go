package config

import (
	"fmt"
	"sort"
)

// An Extension describes how to manipulate files during smudge and clean.
// Extensions are parsed from the Git config.
type Extension struct {
	Name     string
	Clean    string
	Smudge   string
	Priority int
}

// SortExtensions sorts a map of extensions in ascending order by Priority
func SortExtensions(m map[string]Extension) ([]Extension, error) {
	pMap := make(map[int]Extension)
	priorities := make([]int, 0, len(m))
	for n, ext := range m {
		p := ext.Priority
		if _, exist := pMap[p]; exist {
			err := fmt.Errorf("duplicate priority %d on %s", p, n)
			return nil, err
		}
		pMap[p] = ext
		priorities = append(priorities, p)
	}

	sort.Ints(priorities)

	result := make([]Extension, len(priorities))
	for i, p := range priorities {
		result[i] = pMap[p]
	}

	return result, nil
}
