package commands

import (
	"bufio"
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

var (
	prePushDryRun       = false
	prePushDeleteBranch = strings.Repeat("0", 40)
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

	requireGitVersion()

	// Remote is first arg
	if err := git.ValidateRemote(args[0]); err != nil {
		Exit("Invalid remote name %q", args[0])
	}

	ctx := newUploadContext(args[0], prePushDryRun)

	gitscanner := lfs.NewGitScanner(nil)
	if err := gitscanner.RemoteForPush(ctx.Remote); err != nil {
		ExitWithError(err)
	}

	defer gitscanner.Close()

	// We can be passed multiple lines of refs
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 {
			continue
		}

		tracerx.Printf("pre-push: %s", line)

		left, _ := decodeRefs(line)
		if left == prePushDeleteBranch {
			continue
		}

		if err := uploadLeftOrAll(gitscanner, ctx, left); err != nil {
			Print("Error scanning for Git LFS files in %q", left)
			ExitWithError(err)
		}
	}

	ctx.Await()
}

// decodeRefs pulls the sha1s out of the line read from the pre-push
// hook's stdin.
func decodeRefs(input string) (string, string) {
	refs := strings.Split(strings.TrimSpace(input), " ")
	var left, right string

	if len(refs) > 1 {
		left = refs[1]
	}

	if len(refs) > 3 {
		right = "^" + refs[3]
	}

	return left, right
}

func init() {
	RegisterCommand("pre-push", prePushCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&prePushDryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	})
}
