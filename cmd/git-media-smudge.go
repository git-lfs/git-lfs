package main

import (
	".."
	"fmt"
	"io"
	"os"
)

func main() {
	meta := &gitmedia.LargeAsset{}
	dec := gitmedia.NewDecoder(os.Stdin)
	dec.Decode(meta)

	mediafile := gitmedia.LocalMediaPath(meta.SHA1)
	file, err := os.Open(mediafile)
	if err != nil {
		fmt.Printf("Error reading file from local media dir: %s\n", mediafile)
		panic(err)
	}
	defer file.Close()

	io.Copy(os.Stdout, file)
}
