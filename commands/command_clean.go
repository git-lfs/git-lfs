package commands

import (
	"github.com/github/git-media/filters"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/metafile"
	"github.com/spf13/cobra"
	"os"
)

var (
	cleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Implements the Git clean filter",
		Run:   cleanCommand,
	}
)

func cleanCommand(cmd *cobra.Command, args []string) {
	gitmedia.InstallHooks()

	var filename string
	if len(args) > 0 {
		filename = args[0]
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
	RootCmd.AddCommand(cleanCmd)
}
