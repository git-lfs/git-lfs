package commands

import (
	"bufio"
	"os"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tq"
	"github.com/git-lfs/git-lfs/v3/tr"
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

// pushCommand pushes local objects to a Git LFS server.  It has four forms:
//
//	`<remote> <ref>...`
//	`<remote> --stdin`              (reads refs from stdin)
//	`<remote> --object-id <oid>...`
//	`<remote> --object-id --stdin`  (reads oids from stdin)
//
// Remote must be a remote name, not a URL. With --stdin, values are newline
// separated.
//
// pushCommand calculates the git objects to send by comparing the range
// of commits between the local and remote git servers.
func pushCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		Print(tr.Tr.Get("Specify a remote and a remote branch name (`git lfs push origin main`)"))
		os.Exit(1)
	}

	requireGitVersion()

	// Remote is first arg
	if err := cfg.SetValidPushRemote(args[0]); err != nil {
		Exit(tr.Tr.Get("Invalid remote name %q: %s", args[0], err))
	}

	ctx := newUploadContext(pushDryRun)

	var argList []string
	if useStdin {
		if len(args) > 1 {
			Print(tr.Tr.Get("Further command line arguments are ignored with --stdin"))
			os.Exit(1)
		}

		scanner := bufio.NewScanner(os.Stdin) // line-delimited
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				argList = append(argList, line)
			}
		}
		if err := scanner.Err(); err != nil {
			ExitWithError(errors.Wrap(err, tr.Tr.Get("Error reading from stdin:")))
		}
	} else {
		argList = args[1:]
	}

	if pushObjectIDs {
		// We allow no object IDs with `--stdin` to make scripting
		// easier and avoid having to special-case this in scripts.
		if !useStdin && len(argList) < 1 {
			Print(tr.Tr.Get("At least one object ID must be supplied with --object-id"))
			os.Exit(1)
		}
		uploadsWithObjectIDs(ctx, argList)
	} else {
		if !useStdin && !pushAll && len(argList) < 1 {
			Print(tr.Tr.Get("At least one ref must be supplied without --all"))
			os.Exit(1)
		}
		uploadsBetweenRefAndRemote(ctx, argList)
	}
}

func uploadsBetweenRefAndRemote(ctx *uploadContext, refnames []string) {
	tracerx.Printf("Upload refs %v to remote %v", refnames, ctx.Remote)

	updates, err := lfsPushRefs(refnames, pushAll)
	if err != nil {
		Error(err.Error())
		Exit(tr.Tr.Get("Error getting local refs."))
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
			ExitWithError(errors.Wrap(err, tr.Tr.Get("Unable to find local media path:")))
		}

		stat, err := os.Stat(mp)
		if err != nil {
			ExitWithError(errors.Wrap(err, tr.Tr.Get("Unable to stat local media path")))
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
		if ref, ok := reflookup[name]; ok {
			refs[i] = git.NewRefUpdate(cfg.Git, cfg.PushRemote(), ref, nil)
		} else {
			ref := &git.Ref{Name: name, Type: git.RefTypeOther, Sha: name}
			if _, err := git.ResolveRef(name); err != nil {
				return nil, errors.New(tr.Tr.Get("Invalid ref argument: %v", name))
			}
			refs[i] = git.NewRefUpdate(cfg.Git, cfg.PushRemote(), ref, nil)
		}
	}

	return refs, nil
}

func init() {
	RegisterCommand("push", pushCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&pushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
		cmd.Flags().BoolVarP(&pushObjectIDs, "object-id", "o", false, "Push LFS object ID(s)")
		cmd.Flags().BoolVarP(&useStdin, "stdin", "", false, "Read object IDs or refs from stdin")
		cmd.Flags().BoolVarP(&pushAll, "all", "a", false, "Push all objects for the current ref to the remote.")
	})
}
