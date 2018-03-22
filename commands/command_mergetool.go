package commands

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/git/gitattributes"
	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/spf13/cobra"
)

var (
	mergeToolTool string

	mergeToolPrompt   bool
	mergeToolNoPrompt bool
)

func mergetoolCommand(cmd *cobra.Command, args []string) {
	var (
		writeToTemp     = cfg.Git.Bool("mergetool.writeToTemp", false)
		keepBackup      = cfg.Git.Bool("mergetool.keepBackup", true)
		keepTemporaries = cfg.Git.Bool("mergetool.keepTemporaries", false)
	)

	root, err := git.RootDir()
	if err != nil {
		ExitWithError(errors.Wrap(err, "fatal: could not find root"))
	}

	var dirname string
	if writeToTemp {
		dirname = cfg.TempDir()
	} else {
		dirname = root
	}

	conflicts, err := conflicts(args)
	if err != nil {
		ExitWithError(errors.Wrap(err, "fatal: unable to find conflicts"))
	}

	db, err := getObjectDatabase()
	if err != nil {
		ExitWithError(errors.Wrap(err, "fatal: could not open objects"))
	}

	attrs := gitattributes.NewRepository(root)

	for _, conflict := range conflicts {
		backup, err := createMergeBackup(conflict.Path)
		if err != nil {
			ExitWithError(errors.Wrapf(err,
				"fatal: could not create backup of %s",
				conflict.Path))
		}

		common, err := mergeSmudge(db, dirname, conflict.Common)
		if err != nil {
			ExitWithError(errors.Wrap(err, "fatal: could not prepare common for merge"))
		}

		head, err := mergeSmudge(db, dirname, conflict.Head)
		if err != nil {
			ExitWithError(errors.Wrap(err, "fatal: could not prepare head for merge"))
		}

		mergeHead, err := mergeSmudge(db, dirname, conflict.MergeHead)
		if err != nil {
			ExitWithError(errors.Wrap(err, "fatal: could not prepare merge head for merge"))
		}

		tool := findMergeTool(attrs, conflict.Path)
		cmd := tool.Cmd(conflict)

		if !mergeToolNoPrompt || mergeToolPrompt {
			fmt.Printf("Hit return to start merge resolution tool (%s)", tool.Cmd)
			fmt.Scanln()
			fmt.Println()
		}

		if err := cmd.Run(); err != nil {
			ExitWithError(err)
		}

		if writeToTemp && !keepTemporaries {
			if common != nil {
				os.Remove(common.Name())
			}
			os.Remove(head.Name())
			os.Remove(mergeHead.Name())
		}

		if !keepBackup {
			os.Remove(backup.Name())
		}

		// $BASE   <- conflict.Common    : (stage 1)
		// $LOCAL  <- conflict.Head      : (stage 2)
		// $REMOTE <- conflict.MergeHead : (stage 3)
	}
}

type MergeTool struct {
	Path  string
	Name  string
	Trust bool
}

func (m *MergeTool) Cmd(conflict *git.Conflict) *exec.Cmd {
	fields := tools.QuotedFields(m.Name)

	cmd := exec.Command(fields[0], fields[1:]...)
	if len(m.Path) > 0 {
		cmd.Path = m.Path
	}
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("LOCAL=%s", conflict.Head),
		fmt.Sprintf("REMOTE=%s", conflict.MergeHead),
		fmt.Sprintf("MERGE=%s", conflict.Path))

	if conflict.Common != missing {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("BASE=%s", conflict.Common))
	}

	return cmd
}

func findMergeTool(r *gitattributes.Repository, path string) *MergeTool {
	var configured bool
	var tool string = mergeToolTool

	if len(tool) == 0 {
		fmt.Println("A")
		tool, configured = cfg.Git.Get("merge.tool")
		if !configured {
			fmt.Println("B")
			tool, _ = r.Applied(path)["mergetool"]
			fmt.Println("C", tool)
		}
	}

	tpath, _ := cfg.Git.Get(fmt.Sprintf("mergetool.%s.path", tool))
	name, _ := cfg.Git.Get(fmt.Sprintf("mergetool.%s.cmd", tool))
	trust := cfg.Git.Bool(fmt.Sprintf("mergetool.%s.trustExitCode", tool), configured)

	panic(fmt.Sprintf("%s %s %s", tpath, name, trust))

	return &MergeTool{
		Path:  tpath,
		Name:  name,
		Trust: trust,
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
		os.O_CREATE|os.O_EXCL|os.O_RDWR,
		fi.Mode())
}

var (
	missing = "0000000000000000000000000000000000000000"
)

func mergeSmudge(db *odb.ObjectDatabase, dirname, oid string) (*os.File, error) {
	if oid == missing {
		return nil, nil
	}

	sha, _ := hex.DecodeString(oid)
	blob, err := db.Blob(sha)
	if err != nil {
		return nil, err
	}
	defer blob.Close()

	ptr, err := lfs.DecodePointer(blob.Contents)
	if err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile(dirname, "")
	if err != nil {
		return nil, err
	}

	gf := lfs.NewGitFilter(cfg)
	if err := gf.SmudgeToFile(f.Name(), ptr, true, getTransferManifest(), nil); err != nil {
		return nil, err
	}
	return f, nil
}

func init() {
	RegisterCommand("mergetool", mergetoolCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&mergeToolTool, "tool", "t", "", "")

		cmd.Flags().BoolVar(&mergeToolPrompt, "prompt", true, "")
		cmd.Flags().BoolVarP(&mergeToolNoPrompt, "no-prompt", "y", false, "")
	})
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
