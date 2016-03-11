package commands

import (
	"io/ioutil"
	"os"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	pushCmd = &cobra.Command{
		Use: "push",
		Run: pushCommand,
	}
	pushDryRun       = false
	pushDeleteBranch = "(delete)"
	pushObjectIDs    = false
	pushAll          = false
	useStdin         = false

	// shares some global vars and functions with command_pre_push.go
)

func uploadsBetweenRefs(left string, right string) {
	tracerx.Printf("Upload between %v and %v", left, right)

	// Just use scanner here
	pointers, err := scanObjectsLeftToRight(lfs.Config, left, right, lfs.ScanRefsMode)
	if err != nil {
		Panic(err, "Error scanning for Git LFS files")
	}

	uploadObjects(lfs.Config, pointers, pushDryRun)
}

func uploadsBetweenRefAndRemote(remote string, refs []string) {
	tracerx.Printf("Upload refs %v to remote %v", remote, refs)

	scanMode := lfs.ScanLeftToRemoteMode

	if pushAll {
		if len(refs) == 0 {
			gitrefs, err := git.LocalRefs()
			if err != nil {
				Error(err.Error())
				Exit("Error getting local refs.")
			}

			refs = make([]string, len(gitrefs))
			for idx, gitref := range gitrefs {
				refs[idx] = gitref.Name
			}
		}

		scanMode = lfs.ScanRefsMode
	}

	for _, ref := range refs {
		pointers, err := scanObjectsLeftToRight(lfs.Config, ref, "", scanMode)
		if err != nil {
			Panic(err, "Error scanning for Git LFS files in the %q ref", ref)
		}

		uploadObjects(lfs.Config, pointers, pushDryRun)
	}
}

func uploadsWithObjectIDs(oids []string) {
	pointers := make([]*lfs.WrappedPointer, len(oids))
	for idx, oid := range oids {
		pointers[idx] = &lfs.WrappedPointer{Pointer: &lfs.Pointer{Oid: oid}}
	}
	uploadObjects(lfs.Config, pointers, pushDryRun)
}

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
	if len(args) == 0 {
		Print("Specify a remote and a remote branch name (`git lfs push origin master`)")
		os.Exit(1)
	}

	// Remote is first arg
	if err := git.ValidateRemote(args[0]); err != nil {
		Exit("Invalid remote name %q", args[0])
	}
	lfs.Config.CurrentRemote = args[0]

	if useStdin {
		requireStdin("Run this command from the Git pre-push hook, or leave the --stdin flag off.")

		// called from a pre-push hook!  Update the existing pre-push hook if it's
		// one that git-lfs set.
		lfs.InstallHooks(false)

		refsData, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			Panic(err, "Error reading refs on stdin")
		}

		if len(refsData) == 0 {
			return
		}

		left, right := decodeRefs(string(refsData))
		if left == pushDeleteBranch {
			return
		}

		uploadsBetweenRefs(left, right)
	} else if pushObjectIDs {
		if len(args) < 2 {
			Print("Usage: git lfs push --object-id <remote> <lfs-object-id> [lfs-object-id] ...")
			return
		}

		uploadsWithObjectIDs(args[1:])
	} else {
		if len(args) < 1 {
			Print("Usage: git lfs push --dry-run <remote> [ref]")
			return
		}

		uploadsBetweenRefAndRemote(args[0], args[1:])
	}
}

func init() {
	pushCmd.Flags().BoolVarP(&pushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	pushCmd.Flags().BoolVarP(&useStdin, "stdin", "s", false, "Take refs on stdin (for pre-push hook)")
	pushCmd.Flags().BoolVarP(&pushObjectIDs, "object-id", "o", false, "Push LFS object ID(s)")
	pushCmd.Flags().BoolVarP(&pushAll, "all", "a", false, "Push all objects for the current ref to the remote.")

	RootCmd.AddCommand(pushCmd)
}
