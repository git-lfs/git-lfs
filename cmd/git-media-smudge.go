package main

import (
	".."
	"../filters"
	"os"
)

func main() {
	sha, err := gitmedia.Decode(os.Stdin)
	if err != nil {
		gitmedia.Panic(err, "Error reading git-media meta data from stdin:")
	}

	err = gitmediafilters.Smudge(os.Stdout, sha)
	if err != nil {
		smudgerr := err.(*gitmediafilters.SmudgeError)
		gitmedia.Panic(err, "Error reading file from local media dir: %s", smudgerr.Filename)
	}
}
