package main

import (
	".."
	"../filters"
	"fmt"
	"os"
)

func main() {
	cleaned, err := gitmediafilters.Clean(os.Stdin)
	if err != nil {
		fmt.Println("Error cleaning asset")
		panic(err)
	}
	defer cleaned.Close()

	tmpfile := cleaned.File.Name()
	mediafile := gitmedia.LocalMediaPath(cleaned.Sha)
	if stat, _ := os.Stat(mediafile); stat != nil {
		if stat.Size() != cleaned.Size {
			panic(fmt.Sprintf("Files don't match:\n%s\n%s", mediafile, tmpfile))
		}
	} else {
		if err := os.Rename(tmpfile, mediafile); err != nil {
			fmt.Printf("Unable to move %s to %s\n", tmpfile, mediafile)
			panic(err)
		}
	}

	gitmedia.Encode(os.Stdout, cleaned.Sha)
}
