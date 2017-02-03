package commands

import (
	"fmt"
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
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

func uploadsBetweenRefAndRemote(ctx *uploadContext, refnames []string) {
	tracerx.Printf("Upload refs %v to remote %v", refnames, ctx.Remote)

	gitscanner := lfs.NewGitScanner(nil)
	if err := gitscanner.RemoteForPush(ctx.Remote); err != nil {
		ExitWithError(err)
	}
	defer gitscanner.Close()

	refs, err := refsByNames(refnames)
	if err != nil {
		Error(err.Error())
		Exit("Error getting local refs.")
	}

	for _, ref := range refs {
		if err = uploadLeftOrAll(gitscanner, ctx, ref.Name); err != nil {
			Print("Error scanning for Git LFS files in the %q ref", ref.Name)
			ExitWithError(err)
		}
	}

	ctx.Await()
}

func uploadLeftOrAll(g *lfs.GitScanner, ctx *uploadContext, ref string) error {
	var multiErr error
	cb := func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			if multiErr != nil {
				multiErr = fmt.Errorf("%v\n%v", multiErr, err)
			} else {
				multiErr = err
			}
			return
		}

		uploadPointers(ctx, p)
	}

	if pushAll {
		if err := g.ScanRefWithDeleted(ref, cb); err != nil {
			return err
		}
	} else {
		if err := g.ScanLeftToRemote(ref, cb); err != nil {
			return err
		}
	}

	return multiErr
}

func uploadsWithObjectIDs(ctx *uploadContext, oids []string) {
	for _, oid := range oids {
		mp, err := lfs.LocalMediaPath(oid)
		if err != nil {
			ExitWithError(errors.Wrap(err, "Unable to find local media path:"))
		}

		stat, err := os.Stat(mp)
		if err != nil {
			ExitWithError(errors.Wrap(err, "Unable to stat local media path"))
		}

		uploadPointers(ctx, &lfs.WrappedPointer{
			Name: mp,
			Pointer: &lfs.Pointer{
				Oid:  oid,
				Size: stat.Size(),
			},
		})
	}

	ctx.Await()
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
			refs[i] = &git.Ref{Name: name, Type: git.RefTypeOther, Sha: name}
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

	ctx := newUploadContext(args[0], pushDryRun)

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

func init() {
	RegisterCommand("push", pushCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&pushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
		cmd.Flags().BoolVarP(&pushObjectIDs, "object-id", "o", false, "Push LFS object ID(s)")
		cmd.Flags().BoolVarP(&pushAll, "all", "a", false, "Push all objects for the current ref to the remote.")
	})
}
