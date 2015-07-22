package commands

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
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

func uploadsBetweenRefs(left string, right string) *lfs.TransferQueue {
	// Just use scanner here
	pointers, err := lfs.ScanRefs(left, right, nil)
	if err != nil {
		Panic(err, "Error scanning for Git LFS files")
	}

	uploadQueue := lfs.NewUploadQueue(lfs.Config.ConcurrentTransfers(), len(pointers))
	pointerLen := len(pointers)
	for i, pointer := range pointers {
		if pushDryRun {
			Print("push %s", pointer.Name)
			continue
		}
		tracerx.Printf("prepare upload: %s %s %d/%d", pointer.Oid, pointer.Name, i+1, len(pointers))

		u, wErr := lfs.NewUploadable(pointer.Oid, pointer.Name, i+1, pointerLen)
		if wErr != nil {
			if Debugging || wErr.Panic {
				Panic(wErr.Err, wErr.Error())
			} else {
				Exit(wErr.Error())
			}
		}
		uploadQueue.Add(u)
	}

	return uploadQueue
}

func uploadsWithObjectIDs(oids []string) *lfs.TransferQueue {
	uploadQueue := lfs.NewUploadQueue(lfs.Config.ConcurrentTransfers(), len(oids))
	oidLen := len(oids)

	for i, oid := range oids {
		if pushDryRun {
			Print("push object ID %s", oid)
			continue
		}
		tracerx.Printf("prepare upload: %s %d/%d", oid, i+1, len(oids))

		u, wErr := lfs.NewUploadable(oid, "", i+1, oidLen)
		if wErr != nil {
			if Debugging || wErr.Panic {
				Panic(wErr.Err, wErr.Error())
			} else {
				Exit(wErr.Error())
			}
		}
		uploadQueue.Add(u)
	}

	return uploadQueue
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
	var left, right string

	if len(args) == 0 {
		Print("Specify a remote and a remote branch name (`git lfs push origin master`)")
		os.Exit(1)
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
	pointers, err := lfs.ScanRefs(left, right, nil)
	if err != nil {
		Panic(err, "Error scanning for Git LFS files")
	}

	uploadQueue := lfs.NewUploadQueue(lfs.Config.ConcurrentTransfers(), len(pointers))

	for i, pointer := range pointers {
		if pushDryRun {
			Print("push %s", pointer.Name)
			continue
		}
		tracerx.Printf("checking_asset: %s %s %d/%d", pointer.Oid, pointer.Name, i+1, len(pointers))

		u, wErr := lfs.NewUploadable(pointer.Oid, pointer.Name, i+1, len(pointers))
		if wErr != nil {
			if Debugging || wErr.Panic {
				Panic(wErr.Err, wErr.Error())
			} else {
				Exit(wErr.Error())
			}
		}
		uploadQueue.Add(u)
	}

	if !pushDryRun {
		uploadQueue.Process()
		for _, err := range uploadQueue.Errors() {
			if Debugging || err.Panic {
				LoggedError(err.Err, err.Error())
			} else {
				Error(err.Error())
			}
		}

		if len(uploadQueue.Errors()) > 0 {
			os.Exit(2)
		}
	}
}

func init() {
	pushCmd.Flags().BoolVarP(&pushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	pushCmd.Flags().BoolVarP(&useStdin, "stdin", "s", false, "Take refs on stdin (for pre-push hook)")
	RootCmd.AddCommand(pushCmd)
}
