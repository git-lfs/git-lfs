package commands

import (
	"github.com/hawser/git-hawser/git"
	"github.com/hawser/git-hawser/hawser"
	"github.com/hawser/git-hawser/hawserclient"
	"github.com/hawser/git-hawser/scanner"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var (
	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push files to the hawser endpoint",
		Run:   pushCommand,
	}
	dryRun       = false
	useStdin     = false
	deleteBranch = "(delete)"
)

// pushCommand is the command that's run via `git hawser push`. It has two modes
// of operation. The primary mode is run via the git pre-push hook. The pre-push
// hook passes two arguments on the command line:
//   1. Name of the remote to which the push is being done
//   2. URL to which the push is being done
//
// The hook receives commit information on stdin in the form:
//   <local ref> <local sha1> <remote ref> <remote sha1>
//
// In the typical case, pushCommand will get a list of git objects being pushed
// by using the following:
//    git rev-list --objects <local sha1> ^<remote sha1>
//
// If any of those git objects are associated with hawser objects, those hawser
// objects will be pushed to the hawser endpoint.
//
// In the case of pushing a new branch, the list of git objects will be all of
// the git objects in this branch.
//
// In the case of deleting a branch, no attempts to push hawser objects will be
// made.
//
// When pushing hawser objects, the client will first perform an OPTIONS command
// which will determine not only whether or not the client is authorized, but also
// whether or not that hawser endpoint already has the hawser object. If it
// does, the object will not be pushed.
//
// The other mode of operation is the dry run mode. In this mode, the repo
// and refspec are passed on the command line. pushCommand will calculate the
// git objects that would be pushed in a similar manner as above and will print
// out each file name.
func pushCommand(cmd *cobra.Command, args []string) {
	var left, right string

	if len(args) == 0 {
		Print("The git hawser pre-push hook is out of date. Please run `git hawser update`")
		os.Exit(1)
	}

	if useStdin {
		refsData, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			Panic(err, "Error reading refs on stdin")
		}

		if len(refsData) == 0 {
			return
		}

		left, right = decodeRefs(string(refsData))
		if left == deleteBranch {
			return
		}
	} else {
		var repo, refspec string

		if len(args) < 1 {
			Print("Usage: git hawser push --dry-run <repo> [refspec]")
			return
		}

		repo = args[0]
		if len(args) == 2 {
			refspec = args[1]
		}

		localRef, err := git.CurrentRef()
		if err != nil {
			Panic(err, "Error getting local ref")
		}
		left = localRef

		remoteRef, err := git.LsRemote(repo, refspec)
		if err != nil {
			Panic(err, "Error getting remote ref")
		}

		if remoteRef != "" {
			right = "^" + strings.Split(remoteRef, "\t")[0]
		}
	}

	// Just use scanner here
	pointers, err := scanner.Scan(left, right)
	if err != nil {
		Panic(err, "Error scanning for hawser files")
	}

	for i, pointer := range pointers {
		if dryRun {
			Print("push %s", pointer.Name)
			continue
		}
		if wErr := pushAsset(pointer.Oid, pointer.Name, i+1, len(pointers)); wErr != nil {
			Panic(wErr.Err, wErr.Error())
		}
	}
}

// pushAsset pushes the asset with the given oid to the hawser endpoint. It will
// first make an OPTIONS call. If OPTIONS returns a 200 status, it indicates that the
// hawser endpoint already has a hawser object for that oid. The object will
// not be pushed again.
func pushAsset(oid, filename string, index, totalFiles int) *hawser.WrappedError {
	tracerx.Printf("checking_asset: %s %s %d/%d", oid, filename, index, totalFiles)
	path, err := hawser.LocalMediaPath(oid)
	if err != nil {
		return hawser.Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	cb, file, cbErr := hawser.CopyCallbackFile("push", filename, index, totalFiles)
	if cbErr != nil {
		Error(cbErr.Error())
	}
	if file != nil {
		defer file.Close()
	}

	return hawserclient.Upload(&hawserclient.UploadRequest{
		OidPath:      path,
		Filename:     filename,
		CopyCallback: cb,
	})
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
	pushCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	pushCmd.Flags().BoolVarP(&useStdin, "stdin", "s", false, "Take refs on stdin (for pre-push hook)")
	RootCmd.AddCommand(pushCmd)
}
