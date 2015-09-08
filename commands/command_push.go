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
		Use:   "push",
		Short: "Push files to the Git LFS server",
		Run:   pushCommand,
	}
	pushDryRun       = false
	pushDeleteBranch = "(delete)"
	pushObjectIDs    = false
	pushAllToRemote  = false
	useStdin         = false

	// shares some global vars and functions with commmands_pre_push.go
)

func uploadsBetweenRefs(left string, right string) *lfs.TransferQueue {
	tracerx.Printf("Upload between %v and %v", left, right)

	// Just use scanner here
	pointers, err := lfs.ScanRefs(left, right, nil)
	if err != nil {
		Panic(err, "Error scanning for Git LFS files")
	}
	return uploadPointers(pointers)
}

func uploadsBetweenRefAndRemote(ref, remote string) *lfs.TransferQueue {
	tracerx.Printf("Upload between %v and remote %v", ref, remote)

	scanOpt := &lfs.ScanRefsOptions{ScanMode: lfs.ScanLeftToRemoteMode, RemoteName: remote}
	if pushAllToRemote {
		scanOpt.ScanMode = lfs.ScanRefsMode
	}
	pointers, err := lfs.ScanRefs(ref, "", scanOpt)
	if err != nil {
		Panic(err, "Error scanning for Git LFS files")
	}
	return uploadPointers(pointers)
}

func uploadPointers(pointers []*lfs.WrappedPointer) *lfs.TransferQueue {
	totalSize := int64(0)
	for _, p := range pointers {
		totalSize += p.Size
	}

	uploadQueue := lfs.NewUploadQueue(len(pointers), totalSize, pushDryRun)
	for i, pointer := range pointers {
		if pushDryRun {
			Print("push %s => %s", pointer.Oid, pointer.Name)
			continue
		}

		tracerx.Printf("prepare upload: %s %s %d/%d", pointer.Oid, pointer.Name, i+1, len(pointers))

		u, err := lfs.NewUploadable(pointer.Oid, pointer.Name)
		if err != nil {
			if Debugging || lfs.IsFatalError(err) {
				Panic(err, err.Error())
			} else {
				Exit(err.Error())
			}
		}
		uploadQueue.Add(u)
	}

	return uploadQueue
}

func uploadsWithObjectIDs(oids []string) *lfs.TransferQueue {
	uploads := []*lfs.Uploadable{}
	totalSize := int64(0)

	for i, oid := range oids {
		if pushDryRun {
			Print("push object ID %s", oid)
			continue
		}
		tracerx.Printf("prepare upload: %s %d/%d", oid, i+1, len(oids))

		u, err := lfs.NewUploadable(oid, "")
		if err != nil {
			if Debugging || lfs.IsFatalError(err) {
				Panic(err, err.Error())
			} else {
				Exit(err.Error())
			}
		}
		uploads = append(uploads, u)
	}

	uploadQueue := lfs.NewUploadQueue(len(oids), totalSize, pushDryRun)

	for _, u := range uploads {
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
	var uploadQueue *lfs.TransferQueue

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

		left, right := decodeRefs(string(refsData))
		if left == pushDeleteBranch {
			return
		}

		uploadQueue = uploadsBetweenRefs(left, right)
	} else if pushObjectIDs {
		if len(args) < 2 {
			Print("Usage: git lfs push --object-id <remote> <lfs-object-id> [lfs-object-id] ...")
			return
		}

		uploadQueue = uploadsWithObjectIDs(args[1:])
	} else {

		if len(args) < 1 {
			Print("Usage: git lfs push --dry-run <remote> [ref]")
			return
		}

		remote := args[0]
		var ref string
		if len(args) == 2 {
			ref = args[1]
		}

		if ref == "" {
			localRef, err := git.CurrentRef()
			if err != nil {
				Panic(err, "Error getting local ref")
			}
			ref = localRef.Sha
		}

		uploadQueue = uploadsBetweenRefAndRemote(ref, remote)
	}

	if !pushDryRun {
		uploadQueue.Wait()
		for _, err := range uploadQueue.Errors() {
			if Debugging || lfs.IsFatalError(err) {
				LoggedError(err, err.Error())
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
	pushCmd.Flags().BoolVarP(&pushObjectIDs, "object-id", "o", false, "Push LFS object ID(s)")
	pushCmd.Flags().BoolVarP(&pushAllToRemote, "all", "a", false, "Push all objects for the current ref to the remote.")

	RootCmd.AddCommand(pushCmd)
}
