package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/github/git-media/gitmedia"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var (
	lsCmd = &cobra.Command{
		Use:   "ls",
		Short: "List details about Git Media pointers",
		Run:   lsCommand,
	}
)

func lsCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		Exit("Specify a Git Media file.")
	}

	mediaPath, err := getMediaPath(args[0])
	if err != nil {
		Exit("Not a Git Media file: %s", err)
	}

	Print(mediaPath)
}

func getMediaPath(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	oidHash := sha256.New()
	_, err = io.Copy(oidHash, file)
	if err != nil {
		return "", err
	}

	oid := hex.EncodeToString(oidHash.Sum(nil))
	return gitmedia.LocalMediaPath(oid)
}

func init() {
	RootCmd.AddCommand(lsCmd)
}
