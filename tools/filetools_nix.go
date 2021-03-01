// +build !windows

package tools

import "path/filepath"

func CanonicalizeSystemPath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}
