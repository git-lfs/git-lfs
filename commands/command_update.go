package commands

import (
	"bytes"
	"github.com/github/git-media/git"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/pointer"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update local git media configuration",
		Run:   updateCommand,
	}
)

// updateCommand is used for updating parts of git media that reside
// under .git/media.
func updateCommand(cmd *cobra.Command, args []string) {
	updatePrePushHook()
	removeSyncQueue()
}

// updatePrePushHook will force an update of the pre-push hook.
func updatePrePushHook() {
	gitmedia.InstallHooks(true)
	Print("Updated pre-push hook")
}

// removeSyncQueue is intended to update git media repositories that
// used the upload queue. It will walk all git media objects under
// .git/media and create pointer links for them under .git/media/objects.
// After doing so it will remove the upload queue directory.
func removeSyncQueue() {
	queuePath := filepath.Join(gitmedia.LocalMediaDir, "queue")
	if _, err := os.Stat(queuePath); os.IsNotExist(err) {
		return
	}

	objects, err := git.RevListObjects("", "", true)
	if err != nil {
		Panic(err, "Error migrating upload queue")
	}

	Print("Migrating git media objects")

	filepath.Walk(gitmedia.LocalMediaDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			switch info.Name() {
			case "objects", "queue", "tmp", "logs":
				return filepath.SkipDir
			default:
				return nil
			}
		}

		oid := filepath.Base(path)

		file, err := git.Grep(oid)
		if err != nil {
			Panic(err, "Error processing file: %s", path)
		}

		if file != "" {
			for _, obj := range objects {
				if file == obj.Name {
					contents, err := git.CatFile(obj.Sha1)
					if err != nil {
						Panic(err, "Error processing file: %s", path)
					}

					buf := bytes.NewBufferString(contents)
					ptr, err := pointer.Decode(buf)
					if err != nil {
						Panic(err, "Error processing file: %s", path)
					}

					err = ptr.CreateLink(obj.Name)
					if err != nil {
						Panic(err, "Error processing file: %s", path)
					}
				}
			}
		}
		return nil
	})

	os.RemoveAll(queuePath)
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
