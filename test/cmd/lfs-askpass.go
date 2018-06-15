// +build testtools

package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	prompt := strings.Join(os.Args[1:], " ")

	var answer string

	if strings.Contains(prompt, "Username") {
		answer = "user"
		if env, ok := os.LookupEnv("LFS_ASKPASS_USERNAME"); ok {
			answer = env
		}
	} else if strings.Contains(prompt, "Password") {
		answer = "pass"
		if env, ok := os.LookupEnv("LFS_ASKPASS_PASSWORD"); ok {
			answer = env
		}
	}

	fmt.Println(answer)
}
