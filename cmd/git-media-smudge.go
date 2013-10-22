package main

import (
	".."
	"../filters"
	"fmt"
	"os"
)

func main() {
	sha, err := gitmedia.Decode(os.Stdin)
	if err != nil {
		fmt.Println("Error reading git-media meta data from stdin:")
		panic(err)
	}

	err = gitmediafilters.Smudge(os.Stdout, sha)
	if err != nil {
		smudgerr := err.(*gitmediafilters.SmudgeError)
		fmt.Printf("Error reading file from local media dir: %s\n", smudgerr.Filename)
		panic(err)
	}
}
