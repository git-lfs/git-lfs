package commands

import (
	"fmt"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/pointer"
	"github.com/github/git-lfs/scanner"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push files to the Git LFS endpoint",
		Run:   pushCommand,
	}
	dryRun       = false
	useStdin     = false
	deleteBranch = "(delete)"
)

// pushCommand is the command that's run via `git lfs push`. It has two modes
// of operation. The primary mode is run via the git pre-push hook. The pre-push
// hook passes two arguments on the command line:
//   1. Name of the remote to which the push is being done
//   2. URL to which the push is being done
//
// The hook receives commit information on stdin in the form:
//   <local ref> <local sha1> <remote ref> <remote sha1>
//
// In the typical case, pushCommand will get a list of git objects being pushed
// by using the following:
//    git rev-list --objects <local sha1> ^<remote sha1>
//
// If any of those git objects are associated with Git LFS objects, those
// objects will be pushed to the Git LFS API.
//
// In the case of pushing a new branch, the list of git objects will be all of
// the git objects in this branch.
//
// In the case of deleting a branch, no attempts to push Git LFS objects will be
// made.
//
// The other mode of operation is the dry run mode. In this mode, the repo
// and refspec are passed on the command line. pushCommand will calculate the
// git objects that would be pushed in a similar manner as above and will print
// out each file name.
func pushCommand(cmd *cobra.Command, args []string) {
	var left, right string

	if len(args) == 0 {
		Print("The git lfs pre-push hook is out of date. Please run `git lfs update`")
		os.Exit(1)
	}

	lfs.Config.CurrentRemote = args[0]

	if useStdin {
		refsData, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			Panic(err, "Error reading refs on stdin")
		}

		if len(refsData) == 0 {
			return
		}

		left, right = decodeRefs(string(refsData))
		if left == deleteBranch {
			return
		}
	} else {
		var repo, refspec string

		if len(args) < 1 {
			Print("Usage: git lfs push --dry-run <repo> [refspec]")
			return
		}

		repo = args[0]
		if len(args) == 2 {
			refspec = args[1]
		}

		localRef, err := git.CurrentRef()
		if err != nil {
			Panic(err, "Error getting local ref")
		}
		left = localRef

		remoteRef, err := git.LsRemote(repo, refspec)
		if err != nil {
			Panic(err, "Error getting remote ref")
		}

		if remoteRef != "" {
			right = "^" + strings.Split(remoteRef, "\t")[0]
		}
	}

	// Just use scanner here
	pointers, err := scanner.Scan(left, right)
	if err != nil {
		Panic(err, "Error scanning for Git LFS files")
	}

	for i, pointer := range pointers {
		if dryRun {
			Print("push %s", pointer.Name)
			continue
		}
		if wErr := pushAsset(pointer.Oid, pointer.Name, i+1, len(pointers)); wErr != nil {
			if Debugging || wErr.Panic {
				Panic(wErr.Err, wErr.Error())
			} else {
				Exit(wErr.Error())
			}
		}
	}
}

// pushAsset pushes the asset with the given oid to the Git LFS API.
func pushAsset(oid, filename string, index, totalFiles int) *lfs.WrappedError {
	tracerx.Printf("checking_asset: %s %s %d/%d", oid, filename, index, totalFiles)
	path, err := lfs.LocalMediaPath(oid)
	if err != nil {
		return lfs.Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	if err := ensureFile(filename, path); err != nil {
		return lfs.Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	cb, file, cbErr := lfs.CopyCallbackFile("push", filename, index, totalFiles)
	if cbErr != nil {
		Error(cbErr.Error())
	}

	if file != nil {
		defer file.Close()
	}

	fmt.Fprintf(os.Stderr, "Uploading %s\n", filename)
	return lfs.Upload(path, filename, cb)
}

// ensureFile makes sure that the cleanPath exists before pushing it.  If it
// does not exist, it attempts to clean it by reading the file at smudgePath.
func ensureFile(smudgePath, cleanPath string) error {
	if _, err := os.Stat(cleanPath); err == nil {
		return nil
	}

	expectedOid := filepath.Base(cleanPath)
	localPath := filepath.Join(lfs.LocalWorkingDir, smudgePath)
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	cleaned, err := pointer.Clean(file, stat.Size(), nil)
	if err != nil {
		return err
	}

	cleaned.Close()

	if expectedOid != cleaned.Oid {
		return fmt.Errorf("Expected %s to have an OID of %s, got %s", smudgePath, expectedOid, cleaned.Oid)
	}

	return nil
}

// decodeRefs pulls the sha1s out of the line read from the pre-push
// hook's stdin.
func decodeRefs(input string) (string, string) {
	refs := strings.Split(strings.TrimSpace(input), " ")
	var left, right string

	if len(refs) > 1 {
		left = refs[1]
	}

	if len(refs) > 3 {
		right = "^" + refs[3]
	}

	return left, right
}

func init() {
	pushCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	pushCmd.Flags().BoolVarP(&useStdin, "stdin", "s", false, "Take refs on stdin (for pre-push hook)")
	RootCmd.AddCommand(pushCmd)
}
