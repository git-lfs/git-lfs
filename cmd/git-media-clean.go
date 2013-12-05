package main

import (
	".."
	"../filters"
	"flag"
	"os"
)

func main() {
	gitmedia.SetupDebugging(nil)
	flag.Parse()

	cleaned, err := gitmediafilters.Clean(os.Stdin)
	if err != nil {
		gitmedia.Panic(err, "Error cleaning asset")
	}
	defer cleaned.Close()

	tmpfile := cleaned.File.Name()
	mediafile := gitmedia.LocalMediaPath(cleaned.Sha)
	if stat, _ := os.Stat(mediafile); stat != nil {
		if stat.Size() != cleaned.Size {
			gitmedia.Exit("Files don't match:\n%s\n%s", mediafile, tmpfile)
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
