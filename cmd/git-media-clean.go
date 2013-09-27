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

	enc := gitmedia.NewEncoder(writer)
	enc.Encode(cleaned.LargeAsset)
}

func ChooseWriter(cleaned *gitmediaclean.CleanedAsset) io.WriteCloser {
	mediafile := gitmedia.LocalMediaPath(cleaned.SHA1)
	if stat, _ := os.Stat(mediafile); stat == nil {
		return gitmediaclean.NewWriter(cleaned.LargeAsset, cleaned.File)
	} else {
		return gitmediaclean.NewExistingWriter(cleaned.LargeAsset, cleaned.File)
	}
}
