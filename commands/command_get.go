package commands

import (
	"os"
	"os/exec"
	"time"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
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

	target, err := git.ResolveRef(ref)
	if err != nil {
		Panic(err, "Could not resolve git ref")
	}

	current, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not get the current git ref")
	}

	if target == current {
		// We just downloaded the files for the current ref, we can copy them into
		// the working directory and update the git index. We're doing this in a
		// goroutine so they can be copied as they come in, for efficiency.
		watch := q.Watch()

		go func() {
			files := make(map[string]*lfs.WrappedPointer, len(pointers))
			for _, pointer := range pointers {
				files[pointer.Oid] = pointer
			}

			// Fire up the update-index command
			cmd := exec.Command("git", "update-index", "-q", "--refresh", "--stdin")
			stdin, err := cmd.StdinPipe()
			if err != nil {
				Panic(err, "Could not update the index")
			}

			if err := cmd.Start(); err != nil {
				Panic(err, "Could not update the index")
			}

			// As files come in, write them to the wd and update the index
			for oid := range watch {
				pointer, ok := files[oid]
				if !ok {
					continue
				}

				file, err := os.Create(pointer.Name)
				if err != nil {
					Panic(err, "Could not create working directory file")
				}

				if err := lfs.PointerSmudge(file, pointer.Pointer, pointer.Name, nil); err != nil {
					Panic(err, "Could not write working directory file")
				}
				file.Close()

				stdin.Write([]byte(pointer.Name + "\n"))
			}

			stdin.Close()
			if err := cmd.Wait(); err != nil {
				Panic(err, "Error updating the git index")
			}
		}()

		processQueue := time.Now()
		q.Process()
		tracerx.PerformanceSince("process queue", processQueue)
	}
}

func init() {
	RootCmd.AddCommand(getCmd)
}
