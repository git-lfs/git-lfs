package commands

import (
	"bufio"
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/git"
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
	if err := cfg.SetValidRemote(args[0]); err != nil {
		Exit("Invalid remote name %q: %s", args[0], err)
	}

	ctx := newUploadContext(prePushDryRun)

	gitscanner, err := ctx.buildGitScanner()
	if err != nil {
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

		left, right := decodeRefs(line)
		if left.Sha == prePushDeleteBranch {
			continue
		}

		if err := uploadLeftOrAll(gitscanner, ctx, left, right); err != nil {
			Print("Error scanning for Git LFS files in %+v", left)
			ExitWithError(err)
		}
	}

	ctx.Await()
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
