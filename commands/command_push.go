package commands

import (
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/scanner"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var (
	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push files to the Git LFS server",
		Run:   pushCommand,
	}
	pushDryRun       = false
	pushDeleteBranch = "(delete)"
	useStdin         = false

	// shares some global vars and functions with commmands_pre_push.go
)

// pushCommand pushes local objects to a Git LFS server.  It takes two
// arguments:
//
//   `<remote> <remote ref>`
//
// Both a remote name ("origin") or a remote URL are accepted.
//
// pushCommand calculates the git objects to send by looking comparing the range
// of commits between the local and remote git servers.
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
		if left == pushDeleteBranch {
			return
		}
	} else {
		var remoteArg, refArg string

		if len(args) < 1 {
			Print("Usage: git lfs push --dry-run <remote> [ref]")
			return
		}

		remoteArg = args[0]
		if len(args) == 2 {
			refArg = args[1]
		}

		localRef, err := git.CurrentRef()
		if err != nil {
			Panic(err, "Error getting local ref")
		}
		left = localRef

		remoteRef, err := git.LsRemote(remoteArg, refArg)
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
		if pushDryRun {
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

func init() {
	pushCmd.Flags().BoolVarP(&pushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	pushCmd.Flags().BoolVarP(&useStdin, "stdin", "s", false, "Take refs on stdin (for pre-push hook)")
	RootCmd.AddCommand(pushCmd)
}
