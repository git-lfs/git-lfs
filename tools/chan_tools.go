package tools

import "fmt"

func CollectErrsFromChan(errs <-chan error) error {
	var err error

	for e := range errs {
		if err != nil {
			// Combine in case multiple errors
			err = fmt.Errorf("%v\n%v", err, e)
		} else {
			err = e
		}
	}

	return err
}
