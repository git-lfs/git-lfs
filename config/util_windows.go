//go:build windows
// +build windows

package config

// Windows doesn't provide the umask syscall, so return something sane as a
// default.  os.Chmod will only care about the owner bits anyway.
func umask() int {
	return 077
}
