package commands

import (
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tq"
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

// pushCommand pushes local objects to a Git LFS server.  It takes two
// arguments:
//
//   `<remote> <remote ref>`
//
// Remote must be a remote name, not a URL
//
// pushCommand calculates the git objects to send by comparing the range
// of commits between the local and remote git servers.
func pushCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		Print("Specify a remote and a remote branch name (`git lfs push origin master`)")
		os.Exit(1)
	}

	requireGitVersion()

	// Remote is first arg
	if err := cfg.SetValidRemote(args[0]); err != nil {
		Exit("Invalid remote name %q: %s", args[0], err)
	}

	ctx := newUploadContext(pushDryRun)
	if pushObjectIDs {
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

func uploadsBetweenRefAndRemote(ctx *uploadContext, refnames []string) {
	tracerx.Printf("Upload refs %v to remote %v", refnames, ctx.Remote)

	updates, err := lfsPushRefs(refnames, pushAll)
	if err != nil {
		Error(err.Error())
		Exit("Error getting local refs.")
	}

	if err := uploadForRefUpdates(ctx, updates, pushAll); err != nil {
		ExitWithError(err)
	}
}

func uploadsWithObjectIDs(ctx *uploadContext, oids []string) {
	pointers := make([]*lfs.WrappedPointer, len(oids))
	for i, oid := range oids {
		mp, err := ctx.gitfilter.ObjectPath(oid)
		if err != nil {
			ExitWithError(errors.Wrap(err, "Unable to find local media path:"))
		}

		stat, err := os.Stat(mp)
		if err != nil {
			ExitWithError(errors.Wrap(err, "Unable to stat local media path"))
		}

		pointers[i] = &lfs.WrappedPointer{
			Name: mp,
			Pointer: &lfs.Pointer{
				Oid:  oid,
				Size: stat.Size(),
			},
		}
	}

	q := ctx.NewQueue(tq.RemoteRef(currentRemoteRef()))
	ctx.UploadPointers(q, pointers...)
	ctx.CollectErrors(q)
	ctx.ReportErrors()
}

// lfsPushRefs returns valid ref updates from the given ref and --all arguments.
// Either one or more refs can be explicitly specified, or --all indicates all
// local refs are pushed.
func lfsPushRefs(refnames []string, pushAll bool) ([]*git.RefUpdate, error) {
	localrefs, err := git.LocalRefs()
	if err != nil {
		return nil, err
	}

	if pushAll && len(refnames) == 0 {
		refs := make([]*git.RefUpdate, len(localrefs))
		for i, lr := range localrefs {
			refs[i] = git.NewRefUpdate(cfg.Git, cfg.PushRemote(), lr, nil)
		}
		return refs, nil
	}

	reflookup := make(map[string]*git.Ref, len(localrefs))
	for _, ref := range localrefs {
		reflookup[ref.Name] = ref
	}

	refs := make([]*git.RefUpdate, len(refnames))
	for i, name := range refnames {
		if left, ok := reflookup[name]; ok {
			refs[i] = git.NewRefUpdate(cfg.Git, cfg.PushRemote(), left, nil)
		} else {
			left := &git.Ref{Name: name, Type: git.RefTypeOther, Sha: name}
			refs[i] = git.NewRefUpdate(cfg.Git, cfg.PushRemote(), left, nil)
		}
	}

	return refs, nil
}

func init() {
	RegisterCommand("push", pushCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&pushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
		cmd.Flags().BoolVarP(&pushObjectIDs, "object-id", "o", false, "Push LFS object ID(s)")
		cmd.Flags().BoolVarP(&pushAll, "all", "a", false, "Push all objects for the current ref to the remote.")
	})
}
