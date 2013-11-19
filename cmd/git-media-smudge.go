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
		fmt.Fprintln(os.Stderr, "Error reading git-media meta data from stdin:")
		panic(err)
	}

	fmt.Fprintf(os.Stderr, "Downloading media: %s\n", sha)

	err = gitmediafilters.Smudge(os.Stdout, sha)
	if err != nil {
		smudgerr := err.(*gitmediafilters.SmudgeError)
		fmt.Fprintf(os.Stderr, "Error reading file from local media dir: %s\n", smudgerr.Filename)
		panic(err)
	}
}
