package commands

import (
	"strings"

	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/locking"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func postMergeCommand(cmd *cobra.Command, args []string) {
	gitscanner := lfs.NewGitScanner(nil)
	if err := gitscanner.RemoteForPush("master@{1}"); err != nil {
		ExitWithError(err)
	}

	defer gitscanner.Close()

	pointers, err := scanLeftOrAll(gitscanner, "master")
	if err != nil {
		Exit("Error scanning Git LFS files: %+v", err)
	}

	lc, err := locking.NewClient(cfg)
	if err != nil {
		Exit("Unable to create lock system: %v", err.Error())
	}
	defer lc.Close()

	lockSet, err := findLocks(lc, nil, 0, false)
	if err != nil && !strings.Contains(err.Error(), "Not Found") {
		Exit("error finding locks: %s", err)
	}

	name, email := cfg.CurrentCommitter()

	myLocks := make(map[string]struct{})
	for _, p := range pointers {
		if l, ok := lockSet[p.Name]; ok && l.Name == name && l.Email == email {
			myLocks[p.Name] = struct{}{}
		}
	}

	if len(myLocks) == 0 {
		return
	}

	Print("Unlocking %d files", len(myLocks))
	for filename, _ := range myLocks {
		if err := lc.UnlockFile(filename, false); err != nil {
			ExitWithError(errors.Wrap(err, "unable to unlock file"))
		}

		Print("* %s", filename)
	}
}

func init() {
	RegisterCommand("post-merge", postMergeCommand, nil)
}
