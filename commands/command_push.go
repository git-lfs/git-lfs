package commands

import (
	"io/ioutil"
	"os"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

var (
	pushDryRun    = false
	pushObjectIDs = false
	pushAll       = false
	useStdin      = false

	// shares some global vars and functions with command_pre_push.go
)

func uploadsBetweenRefs(ctx *uploadContext, left string, right string) {
	tracerx.Printf("Upload between %v and %v", left, right)

	scanOpt := lfs.NewScanRefsOptions()
	scanOpt.ScanMode = lfs.ScanRefsMode
	scanOpt.RemoteName = cfg.CurrentRemote

	pointers, err := lfs.ScanRefs(left, right, scanOpt)
	if err != nil {
		Panic(err, "Error scanning for Git LFS files")
	}

	upload(ctx, pointers)
}

func uploadsBetweenRefAndRemote(ctx *uploadContext, refnames []string) {
	tracerx.Printf("Upload refs %v to remote %v", refnames, cfg.CurrentRemote)

	scanOpt := lfs.NewScanRefsOptions()
	scanOpt.ScanMode = lfs.ScanLeftToRemoteMode
	scanOpt.RemoteName = cfg.CurrentRemote

	if pushAll {
		scanOpt.ScanMode = lfs.ScanRefsMode
	}

	refs, err := refsByNames(refnames)
	if err != nil {
		Error(err.Error())
		Exit("Error getting local refs.")
	}

	for _, ref := range refs {
		pointers, err := lfs.ScanRefs(ref.Name, "", scanOpt)
		if err != nil {
			Panic(err, "Error scanning for Git LFS files in the %q ref", ref.Name)
		}

		upload(ctx, pointers)
	}
}

func uploadsWithObjectIDs(ctx *uploadContext, oids []string) {
	pointers := make([]*lfs.WrappedPointer, len(oids))

	for idx, oid := range oids {
		pointers[idx] = &lfs.WrappedPointer{Pointer: &lfs.Pointer{Oid: oid}}
	}

	upload(ctx, pointers)
}

func refsByNames(refnames []string) ([]*git.Ref, error) {
	localrefs, err := git.LocalRefs()
	if err != nil {
		return nil, err
	}

	if pushAll && len(refnames) == 0 {
		return localrefs, nil
	}

	reflookup := make(map[string]*git.Ref, len(localrefs))
	for _, ref := range localrefs {
		reflookup[ref.Name] = ref
	}

	refs := make([]*git.Ref, len(refnames))
	for i, name := range refnames {
		if ref, ok := reflookup[name]; ok {
			refs[i] = ref
		} else {
			refs[i] = &git.Ref{name, git.RefTypeOther, name}
		}
	}

	return refs, nil
}

// pushCommand pushes local objects to a Git LFS server.  It takes two
// arguments:
//
//   `<remote> <remote ref>`
//
// Remote must be a remote name, not a URL
//
// pushCommand calculates the git objects to send by looking comparing the range
// of commits between the local and remote git servers.
func pushCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		Print("Specify a remote and a remote branch name (`git lfs push origin master`)")
		os.Exit(1)
	}

	requireGitVersion()

	// Remote is first arg
	if err := git.ValidateRemote(args[0]); err != nil {
		Exit("Invalid remote name %q", args[0])
	}

	cfg.CurrentRemote = args[0]
	ctx := newUploadContext(pushDryRun)

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
		if left == prePushDeleteBranch {
			return
		}

		uploadsBetweenRefs(ctx, left, right)
	} else if pushObjectIDs {
		if len(args) < 2 {
			Print("Usage: git lfs push --object-id <remote> <lfs-object-id> [lfs-object-id] ...")
			return
		}

		uploadsWithObjectIDs(ctx, args[1:])
	} else {
		if len(args) < 1 {
			Print("Usage: git lfs push --dry-run <remote> [ref]")
			return
		}

		uploadsBetweenRefAndRemote(ctx, args[1:])
	}
}

func init() {
	RegisterCommand("push", pushCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&pushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
		cmd.Flags().BoolVarP(&useStdin, "stdin", "s", false, "Take refs on stdin (for pre-push hook)")
		cmd.Flags().BoolVarP(&pushObjectIDs, "object-id", "o", false, "Push LFS object ID(s)")
		cmd.Flags().BoolVarP(&pushAll, "all", "a", false, "Push all objects for the current ref to the remote.")
	})
}
