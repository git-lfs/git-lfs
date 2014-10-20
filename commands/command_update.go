package commands

import (
	"github.com/github/git-media/gitmedia"
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

	os.RemoveAll(queuePath)
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
