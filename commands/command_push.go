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

func uploadsBetweenRefs(cli *clientContext, left string, right string) {
	tracerx.Printf("Upload between %v and %v", left, right)

	scanOpt := lfs.NewScanRefsOptions()
	scanOpt.ScanMode = lfs.ScanRefsMode
	scanOpt.RemoteName = cli.RemoteName

	pointers, err := lfs.ScanRefs(left, right, scanOpt)
	if err != nil {
		Panic(err, "Error scanning for Git LFS files")
	}

	cli.Upload(nil, pointers)
}

func uploadsBetweenRefAndRemote(cli *clientContext, refs []string) {
	tracerx.Printf("Upload refs %v to remote %v", cli.RemoteName, refs)

	scanOpt := lfs.NewScanRefsOptions()
	scanOpt.ScanMode = lfs.ScanLeftToRemoteMode
	scanOpt.RemoteName = cli.RemoteName

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

		scanOpt.ScanMode = lfs.ScanRefsMode
	}

	for _, ref := range refs {
		pointers, err := lfs.ScanRefs(ref, "", scanOpt)

		if err != nil {
			Panic(err, "Error scanning for Git LFS files in the %q ref", ref)
		}

		cli.Upload(nil, pointers)
	}
}

func uploadsWithObjectIDs(cli *clientContext, oids []string) {
	pointers := make([]*lfs.WrappedPointer, len(oids))

	for idx, oid := range oids {
		pointers[idx] = &lfs.WrappedPointer{Pointer: &lfs.Pointer{Oid: oid}}
	}

	cli.Upload(nil, pointers)
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

	cli := newClient()
	cli.RemoteName = lfs.Config.CurrentRemote
	cli.DryRun = pushDryRun

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

		left, right, _ := decodeRefs(string(refsData))
		if left == pushDeleteBranch {
			return
		}

		uploadsBetweenRefs(cli, left, right)
	} else if pushObjectIDs {
		if len(args) < 2 {
			Print("Usage: git lfs push --object-id <remote> <lfs-object-id> [lfs-object-id] ...")
			return
		}

		uploadsWithObjectIDs(cli, args[1:])
	} else {
		if len(args) < 1 {
			Print("Usage: git lfs push --dry-run <remote> [ref]")
			return
		}

		uploadsBetweenRefAndRemote(cli, args[1:])
	}
}

func init() {
	pushCmd.Flags().BoolVarP(&pushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	pushCmd.Flags().BoolVarP(&useStdin, "stdin", "s", false, "Take refs on stdin (for pre-push hook)")
	pushCmd.Flags().BoolVarP(&pushObjectIDs, "object-id", "o", false, "Push LFS object ID(s)")
	pushCmd.Flags().BoolVarP(&pushAll, "all", "a", false, "Push all objects for the current ref to the remote.")

	RootCmd.AddCommand(pushCmd)
}
