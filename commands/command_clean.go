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
	if len(c.Args) > 1 {
		filename = c.Args[1]
	} else {
		filename = ""
	}

	cleaned, err := filters.Clean(os.Stdin)
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

		gitmedia.QueueUpload(cleaned.Sha, filename)
		gitmedia.Debug("Writing %s", mediafile)
	}

	metafile.Encode(os.Stdout, cleaned.Sha)
}

func init() {
	registerCommand("clean", func(c *Command) RunnableCommand {
		return &CleanCommand{Command: c}
	})
}
