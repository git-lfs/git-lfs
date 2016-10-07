// +build !windows

package git

func cleanRootPath(pattern string) string {
	return pattern
}
