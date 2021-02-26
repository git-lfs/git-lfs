// +build windows

package tools

import "fmt"

func GetMaxFileDescriptors() (uint64, error) {
	return 0, fmt.Errorf("Not implemented")
}
