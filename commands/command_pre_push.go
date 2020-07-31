package commands

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/git"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

var (
	prePushDryRun = false
)

// prePushCommand is run through Git's pre-push hook. The pre-push hook passes
// two arguments on the command line:
//
//   1. Name of the remote to which the push is being done
//   2. URL to which the push is being done
//
// The hook receives commit information on stdin in the form:
//   <local ref> <local sha1> <remote ref> <remote sha1>
//
// In the typical case, prePushCommand will get a list of git objects being
// pushed by using the following:
//
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
func prePushCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		Print("This should be run through Git's pre-push hook.  Run `git lfs update` to install it.")
		os.Exit(1)
	}

	if cfg.Os.Bool("GIT_LFS_SKIP_PUSH", false) {
		return
	}

	requireGitVersion()

	// Remote is first arg
	remote, _ := git.MapRemoteURL(args[0], true)
	if err := cfg.SetValidPushRemote(remote); err != nil {
		Exit("Invalid remote name %q: %s", args[0], err)
	}

	ctx := newUploadContext(prePushDryRun)
	updates := prePushRefs(os.Stdin)
	if err := uploadForRefUpdates(ctx, updates, false); err != nil {
		ExitWithError(err)
	}
}

// prePushRefs parses commit information that the pre-push git hook receives:
//
//   <local ref> <local sha1> <remote ref> <remote sha1>
//
// Each line describes a proposed update of the remote ref at the remote sha to
// the local sha. Multiple updates can be received on multiple lines (such as
// from 'git push --all'). These updates are typically received over STDIN.
func prePushRefs(r io.Reader) []*git.RefUpdate {
	scanner := bufio.NewScanner(r)
	refs := make([]*git.RefUpdate, 0, 1)

	// We can be passed multiple lines of refs
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		tracerx.Printf("pre-push: %s", line)

		left, right := decodeRefs(line)
		if git.IsZeroObjectID(left.Sha) {
			continue
		}

		refs = append(refs, git.NewRefUpdate(cfg.Git, cfg.PushRemote(), left, right))
	}

	return refs
}

// decodeRefs pulls the sha1s out of the line read from the pre-push
// hook's stdin.
func decodeRefs(input string) (*git.Ref, *git.Ref) {
	refs := strings.Split(strings.TrimSpace(input), " ")
	for len(refs) < 4 {
		refs = append(refs, "")
	}

	leftRef := git.ParseRef(refs[0], refs[1])
	rightRef := git.ParseRef(refs[2], refs[3])
	return leftRef, rightRef
}

func init() {
	RegisterCommand("pre-push", prePushCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&prePushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	})
}
