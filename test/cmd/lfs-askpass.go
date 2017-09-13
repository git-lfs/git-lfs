// +build testtools

package main

import (
	"fmt"
	"os"
)

func main() {
	var password string = "pass"
	if env, ok := os.LookupEnv("LFS_ASKPASS_PASSWORD"); ok {
		password = env
	}

	fmt.Println(password)
}
