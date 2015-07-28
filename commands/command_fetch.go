package commands

import (
	"time"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	fetchCmd = &cobra.Command{
		Use:   "fetch",
		Short: "Downloads LFS files",
		Run:   fetchCommand,
	}
)

func fetchCommand(cmd *cobra.Command, args []string) {
	var ref string
	var err error

	if len(args) == 1 {
		ref = args[0]
	} else {
		ref, err = git.CurrentRef()
		if err != nil {
			Panic(err, "Could not fetch")
		}
	}

	fetchRef(ref)
}

func init() {
	RootCmd.AddCommand(fetchCmd)
}

// Fetch all binaries for a given ref (that we don't have already)
func fetchRef(ref string) {
	pointers, err := lfs.ScanRefs(ref, "", nil)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}
	fetchPointers(pointers)
}

func fetchPointers(pointers []*lfs.WrappedPointer) {
	fetchAndReportToChan(pointers, nil)
}

// Fetch and report completion of each OID to a channel (optional, pass nil to skip)
func fetchAndReportToChan(pointers []*lfs.WrappedPointer, out chan<- *lfs.WrappedPointer) {
	q := lfs.NewDownloadQueue(lfs.Config.ConcurrentTransfers(), len(pointers))

	for _, p := range pointers {
		// Only add to download queue if local file is not the right size already
		// This avoids previous case of over-reporting a requirement for files we already have
		// which would only be skipped by PointerSmudgeObject later
		if !lfs.ObjectExistsOfSize(p.Oid, p.Size) {
			q.Add(lfs.NewDownloadable(p))
		}
	}

	if out != nil {
		dlwatch := q.Watch()

		go func() {
			// fetch only reports single OID, but OID *might* be referenced by multiple
			// WrappedPointers if same content is at multiple paths, so map oid->slice
			oidToPointers := make(map[string][]*lfs.WrappedPointer, len(pointers))
			for _, pointer := range pointers {
				plist := oidToPointers[pointer.Oid]
				oidToPointers[pointer.Oid] = append(plist, pointer)
			}

			for oid := range dlwatch {
				plist, ok := oidToPointers[oid]
				if !ok {
					continue
				}
				for _, p := range plist {
					out <- p
				}
			}
			close(out)
		}()

	}
	processQueue := time.Now()
	q.Process()
	tracerx.PerformanceSince("process queue", processQueue)
}
