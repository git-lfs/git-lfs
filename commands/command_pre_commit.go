package commands

import (
	"os"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/locking"
	"github.com/spf13/cobra"
)

func preCommitCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()

	name, email := cfg.CurrentCommitter()
	lc, err := locking.NewClient(cfg)
	if err != nil {
		Exit("Unable to create lock system: %v", err.Error())
	}
	defer lc.Close()

	lockSet, err := findLocks(lc, nil, 0, false)
	if err != nil {
		Print("could not find locks, proceeding anyway: %s", err)
		os.Exit(0)
	}

	files, err := git.StagedFiles()
	if err != nil {
		Exit("error finding staged files: %s", err)
	}

	lockConflicts := make([]string, 0, len(lockSet))

	for _, f := range files {
		if l, ok := lockSet[f]; ok && !(l.Name == name && l.Email == email) {
			lockConflicts = append(lockConflicts, f)
		}
	}

	if len(lockConflicts) > 0 {
		Error("Some files are locked")
		for _, file := range lockConflicts {
			Error("* %s", file)
		}
		os.Exit(1)
	}
}

func init() {
	RegisterCommand("pre-commit", preCommitCommand, nil)
}
