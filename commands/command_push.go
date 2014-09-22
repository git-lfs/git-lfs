package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/gitmediaclient"
	"github.com/github/git-media/pointer"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push files to the media endpoint",
		Run:   pushCommand,
	}
	dryRun       = false
	z40          = "0000000000000000000000000000000000000000"
	deleteBranch = "(delete)"
)

func pushCommand(cmd *cobra.Command, args []string) {
	// TODO handle (delete) case, not sending anything
	refsData, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		Panic(err, "Error reading refs on stdin")
	}

	if len(refsData) == 0 {
		return
	}

	// TODO let's pull this into a nice iteratable thing like the queue provides
	refs := strings.Split(strings.TrimSpace(string(refsData)), " ")

	if refs[0] == deleteBranch {
		return
	}

	refArgs := []string{"rev-list", "--objects"}
	if len(refs) > 1 {
		refArgs = append(refArgs, refs[1])
	}
	if len(refs) > 3 && refs[3] != z40 {
		refArgs = append(refArgs, "^"+refs[3])
	}

	output, err := exec.Command("git", refArgs...).Output()
	if err != nil {
		Panic(err, "Error running git rev-list --objects %v", refArgs)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(output))
	links := make([]*pointer.Link, 0)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		sha1 := line[0]

		linkPath := filepath.Join(gitmedia.LocalLinkDir, sha1[0:2], sha1[2:len(sha1)])

		linkFile, err := os.Open(linkPath)
		if err != nil {
			continue
		}

		link, err := pointer.DecodeLink(linkFile)
		if err != nil {
			Panic(err, "Error decoding link file") // don't panic
		}

		links = append(links, link)
	}

	for i, link := range links {
		if dryRun {
			fmt.Println("push", link.Oid, link.Name)
			continue
		}
		if wErr := pushAsset(link.Oid, link.Name, i+1, len(links)); wErr != nil {
			Panic(wErr.Err, wErr.Error())
		}
		fmt.Printf("\n")
	}
}

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

func init() {
	pushCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	RootCmd.AddCommand(pushCmd)
}
