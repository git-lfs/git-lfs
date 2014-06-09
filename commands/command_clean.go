package commands

import (
	"github.com/github/git-media/filters"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/metafile"
	"os"
)

type CleanCommand struct {
	*Command
}

func (c *CleanCommand) Run() {
	gitmedia.InstallHooks()

	var filename string
	if len(c.Args) > 0 {
		filename = c.Args[0]
	} else {
		filename = ""
	}

	cleaned, err := filters.Clean(os.Stdin)
	if err != nil {
		Panic(err, "Error cleaning asset")
	}
	defer cleaned.Close()

	tmpfile := cleaned.File.Name()
	mediafile, err := gitmedia.LocalMediaPath(cleaned.Sha)
	if err != nil {
		Panic(err, "Unable to get local media path.")
	}

	if stat, _ := os.Stat(mediafile); stat != nil {
		if stat.Size() != cleaned.Size {
			Exit("Files don't match:\n%s\n%s", mediafile, tmpfile)
		}
		Debug("%s exists", mediafile)
	} else {
		if err := os.Rename(tmpfile, mediafile); err != nil {
			Panic(err, "Unable to move %s to %s\n", tmpfile, mediafile)
		}

		if err = gitmedia.QueueUpload(cleaned.Sha, filename); err != nil {
			Panic(err, "Unable to add %s to queue", cleaned.Sha)
		}
		Debug("Writing %s", mediafile)
	}

	metafile.Encode(os.Stdout, cleaned.Sha)
}

func init() {
	registerCommand("clean", func(c *Command) RunnableCommand {
		return &CleanCommand{Command: c}
	})
}
