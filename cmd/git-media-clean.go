package main

import (
	".."
	"../filters"
	"fmt"
	"os"
)

func main() {
	gitmedia.SetupDebugging()

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
			gitmedia.Panic(nil, "Files don't match:\n%s\n%s", mediafile, tmpfile)
		}
		gitmedia.Debug("%s exists", mediafile)
	} else {
		if err := os.Rename(tmpfile, mediafile); err != nil {
			gitmedia.Panic(err, "Unable to move %s to %s\n", tmpfile, mediafile)
		}

		gitmedia.QueueUpload(cleaned.Sha)
		gitmedia.Debug("Writing %s", mediafile)
	}

	gitmedia.Encode(os.Stdout, cleaned.Sha)
}
