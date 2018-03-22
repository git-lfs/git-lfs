package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/spf13/cobra"
)

func mergetoolCommand(cmd *cobra.Command, args []string) {
	conflicts, err := conflicts(args)
	if err != nil {
		ExitWithError(errors.Wrap(err, "fatal: unable to find conflicts"))
	}

	db, err := getObjectDatabase()
	if err != nil {
		ExitWithError(errors.Wrap(err, "fatal: could not open objects"))
	}

	gf := lfs.NewGitFilter(cfg)

	for _, conflict := range conflicts {
		backup, err := createMergeBackup(conflict.Path)
		if err != nil {
			ExitWithError(errors.Wrapf(err,
				"fatal: could not create backup of %s",
				conflict.Path))
		}

		common, err := mergeSmudge(conflict.Common)
		if err != nil {
			ExitWithError(errors.Wrap(err, "fatal: could not prepare common for merge"))
		}

		head, err := mergeSmudge(conflict.Head)
		if err != nil {
			ExitWithError(errors.Wrap(err, "fatal: could not prepare head for merge"))
		}

		mergeHead, err := mergeSmudge(conflict.MergeHead)
		if err != nil {
			ExitWithError(errors.Wrap(err, "fatal: could not prepare merge head for merge"))
		}

		cmd := subprocess.ExecCommand()
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("BASE=%s", common.Name()),
			fmt.Sprintf("LOCAL=%s", head.Name()),
			fmt.Sprintf("REMOTE=%s", mergeHead.Name()),
			fmt.Sprintf("MERGE=%s", conflict.Path),
		)

		if err := cmd.Run(); err != nil {
			ExitWithError(err)
		}

		// $BASE   <- conflict.Common    : (stage 1)
		// $LOCAL  <- conflict.Head      : (stage 2)
		// $REMOTE <- conflict.MergeHead : (stage 3)
	}
}

func createMergeBackup(name string) (*os.File, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(fmt.Sprintf("%s.orig", f.Name()),
		os.O_CREATE|os.O_EXCL, os.O_RDWR,
		int64(fi.Mode()))
}

func mergeSmudge(oid string) (*os.File, error) {
	blob, err := db.Blob(conflict.Common)
	if err != nil {
		return nil, err
	}

	ptr, err := lfs.DecodePointer(oid)
	if err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile(cfg.TempDir(), "")
	if err != nil {
		return nil, err
	}

	if err := gf.SmudgeToFile(f, ptr, true, getTransferManifest(), nil); err != nil {
		return nil, err
	}
	return f, nil
}

func init() {
	RegisterCommand("mergetool", mergetoolCommand, nil)
}

func conflicts(args []string) ([]*git.Conflict, error) {
	if len(args) > 0 {
		var all []*git.Conflict
		for _, arg := range args {
			stat, err := os.Stat(arg)
			if err != nil {
				return nil, err
			}

			if stat.IsDir() {
				conflicts, err := git.ConflictsInDir(stat.Name())
				if err != nil {
					return nil, err
				}

				all = append(all, conflicts...)
			} else {
				conflict, err := git.ConflictDetails(stat.Name())
				if err != nil {
					return nil, err
				}

				all = append(all, conflict)
			}
		}
		return all, nil
	}

	return git.AllConflicts()
}
