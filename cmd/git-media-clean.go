package main

import (
	".."
	"../clean"
	"fmt"
	"io"
	"os"
)

func main() {
	cleaned, err := gitmediaclean.Clean(os.Stdin)
	if err != nil {
		fmt.Println("Error cleaning asset")
		panic(err)
	}

	writer := ChooseWriter(cleaned)
	defer writer.Close()
	gitmedia.Encode(writer, cleaned.Sha)
}

func ChooseWriter(cleaned *gitmediaclean.CleanedAsset) io.WriteCloser {
	mediafile := gitmedia.LocalMediaPath(cleaned.Sha)
	if stat, _ := os.Stat(mediafile); stat == nil {
		return gitmediaclean.NewWriter(cleaned)
	} else {
		return gitmediaclean.NewExistingWriter(cleaned)
	}
}
