package main

import (
	".."
	"fmt"
	"io"
	"os"
)

func main() {
	sha, err := gitmedia.Decode(os.Stdin)
	if err != nil {
		fmt.Println("Error reading git-media meta data from stdin:")
		panic(err)
	}

	mediafile := gitmedia.LocalMediaPath(sha)
	file, err := os.Open(mediafile)
	if err != nil {
		fmt.Printf("Error reading file from local media dir: %s\n", mediafile)
		panic(err)
	}
	defer file.Close()

	io.Copy(os.Stdout, file)
}
