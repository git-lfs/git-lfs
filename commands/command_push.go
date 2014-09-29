package commands

import (
	"fmt"
	"github.com/github/git-media/git"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/gitmediaclient"
	"github.com/github/git-media/pointer"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var (
	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push files to the media endpoint",
		Run:   pushCommand,
	}
	dryRun       = false
	deleteBranch = "(delete)"
)

// pushCommand is the command that's run via `git media push`. It has two modes
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
// If any of those git objects are associated with git media objects, those git
// media objects will be pushed to the git media endpoint.
//
// In the case of pushing a new branch, the list of git objects will be all of
// the git objects in this branch.
//
// In the case of deleting a branch, no attempts to push git media objects will be
// made.
//
// When pushing git media objects, the client will first perform an OPTIONS command
// which will determine not only whether or not the client is authorized, but also
// whether or not that git media endpoint already has the git media object. If it
// does, the object will not be pushed.
//
// The other mode of operation is the dry run mode. In this mode, the repo
// and refspec are passed on the command line. pushCommand will calculate the
// git objects that would be pushed in a similar manner as above and will print
// out each file name.
func pushCommand(cmd *cobra.Command, args []string) {
	var left, right string

	if dryRun {
		if len(args) < 1 {
			Print("Usage: git media push --dry-run <repo> [refspec]")
			return
		}

		ref, err := gitmedia.CurrentRef()
		if err != nil {
			Panic(err, "Error getting current ref")
		}
		left = ref
		if len(args) == 2 {
			right = fmt.Sprintf("^%s/%s", args[0], args[1])
		}
	} else {
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
	}

	links := linksFromRefs(left, right)

	for i, link := range links {
		if dryRun {
			Print("push %s", link.Name)
			continue
		}
		if wErr := pushAsset(link.Oid, link.Name, i+1, len(links)); wErr != nil {
			Panic(wErr.Err, wErr.Error())
		}
	}
}

// pushAsset pushes the asset with the given oid to the git media endpoint. It will
// first make an OPTIONS call. If OPTIONS returns a 200 status, it indicates that the
// git media endpoint already has a git media object for that oid. The object will
// not be pushed again.
func pushAsset(oid, filename string, index, totalFiles int) *gitmedia.WrappedError {
	path, err := gitmedia.LocalMediaPath(oid)
	if err != nil {
		return gitmedia.Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	status, err := gitmediaclient.Options(path)
	if err != nil {
		return gitmedia.Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	if status == 200 {
		return nil
	}

	cb, file, cbErr := gitmedia.CopyCallbackFile("push", filename, index, totalFiles)
	if cbErr != nil {
		Error(cbErr.Error())
	}

	err = gitmediaclient.Put(path, filename, cb)
	if file != nil {
		file.Close()
	}

	return nil
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

// linksFromRefs runs git.RevListObjects for the passed refs and builds
// a slice of pointer.Link's for any object that has an associated git media file.
func linksFromRefs(left, right string) []*pointer.Link {
	revList, err := git.RevListObjects(left, right, false)
	if err != nil {
		Panic(err, "Error running git rev-list --objects %s %s", left, right)
	}

	links := make([]*pointer.Link, 0)
	for _, object := range revList {
		link, err := pointer.FindLink(object.Sha1)
		if err != nil {
			continue
		}

		links = append(links, link)
	}

	return links
}

func init() {
	pushCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	RootCmd.AddCommand(pushCmd)
}
