package commands

import (
	"os"
	"os/exec"
	"time"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

var (
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "get",
		Run:   getCommand,
	}
)

func getCommand(cmd *cobra.Command, args []string) {
	var ref string
	var err error

	if len(args) == 1 {
		ref = args[0]
	} else {
		ref, err = git.CurrentRef()
		if err != nil {
			Panic(err, "Could not get")
		}
	}

	pointers, err := lfs.ScanRefs(ref, "")
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	q := lfs.NewDownloadQueue(lfs.Config.ConcurrentTransfers(), len(pointers))

	for _, p := range pointers {
		q.Add(lfs.NewDownloadable(p))
	}

	processQueue := time.Now()
	q.Process()
	tracerx.PerformanceSince("process queue", processQueue)

	target, err := git.ResolveRef(ref)
	if err != nil {
	}

	current, err := git.CurrentRef()
	if err != nil {
	}

	if target == current {
		// We just downloaded the files for the current ref, we can copy them into
		// the working directory and update the git index
		updateWd := time.Now()
		for _, pointer := range pointers {
			file, err := os.Create(pointer.Name)
			if err != nil {
				Panic(err, "Could not create working directory file")
			}

			if err := lfs.PointerSmudge(file, pointer.Pointer, pointer.Name, nil); err != nil {
				Panic(err, "Could not write working directory file")
			}
		}
		tracerx.PerformanceSince("update working directory", updateWd)

		updateIndex := time.Now()
		cmd := exec.Command("git", "update-index", "-q", "--refresh", "--stdin")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			Panic(err, "Could not update the index")
		}

		if err := cmd.Start(); err != nil {
			Panic(err, "Could not update the index")
		}

		for _, pointer := range pointers {
			stdin.Write([]byte(pointer.Name + "\n"))
		}
		stdin.Close()
		cmd.Wait()
		tracerx.PerformanceSince("update index", updateIndex)
	}

}

func init() {
	RootCmd.AddCommand(getCmd)
}
