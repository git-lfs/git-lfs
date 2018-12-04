// +build windows

package tools

func doWithUmask(mask int, f func() error) error {
	return f()
}
