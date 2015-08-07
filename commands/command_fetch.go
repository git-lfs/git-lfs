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
	fetchIncludeArg string
	fetchExcludeArg string
)

func fetchCommand(cmd *cobra.Command, args []string) {
	var refs []string

	if len(args) > 0 {
		refs = args
	} else {
		ref, err := git.CurrentRef()
		if err != nil {
			Panic(err, "Could not fetch")
		}
		refs = []string{ref}
	}

	includePaths, excludePaths := determineIncludeExcludePaths(fetchIncludeArg, fetchExcludeArg)

	// Fetch refs sequentially per arg order; duplicates in later refs will be ignored
	for _, ref := range refs {
		fetchRef(ref, includePaths, excludePaths)
	}

}

func init() {
	fetchCmd.Flags().StringVarP(&fetchIncludeArg, "include", "I", "", "Include a list of paths")
	fetchCmd.Flags().StringVarP(&fetchExcludeArg, "exclude", "X", "", "Exclude a list of paths")
	RootCmd.AddCommand(fetchCmd)
}

func fetchRefToChan(ref string, include, exclude []string) chan *lfs.WrappedPointer {
	c := make(chan *lfs.WrappedPointer)
	pointers, err := lfs.ScanRefs(ref, "", nil)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	go fetchAndReportToChan(pointers, include, exclude, c)

	return c
}

// Fetch all binaries for a given ref (that we don't have already)
func fetchRef(ref string, include, exclude []string) {
	pointers, err := lfs.ScanRefs(ref, "", nil)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}
	fetchPointers(pointers, include, exclude)
}

func fetchPointers(pointers []*lfs.WrappedPointer, include, exclude []string) {
	fetchAndReportToChan(pointers, include, exclude, nil)
}

// Fetch and report completion of each OID to a channel (optional, pass nil to skip)
func fetchAndReportToChan(pointers []*lfs.WrappedPointer, include, exclude []string, out chan<- *lfs.WrappedPointer) {

	totalSize := int64(0)
	for _, p := range pointers {
		totalSize += p.Size
	}
	q := lfs.NewDownloadQueue(len(pointers), totalSize, false)

	for _, p := range pointers {
		// Only add to download queue if local file is not the right size already
		// This avoids previous case of over-reporting a requirement for files we already have
		// which would only be skipped by PointerSmudgeObject later
		passFilter := lfs.FilenamePassesIncludeExcludeFilter(p.Name, include, exclude)
		if !lfs.ObjectExistsOfSize(p.Oid, p.Size) && passFilter {
			q.Add(lfs.NewDownloadable(p))
		} else {
			// If we already have it, or it won't be fetched
			// report it to chan immediately to support pull/checkout
			if out != nil {
				out <- p
			}

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
	q.Wait()
	tracerx.PerformanceSince("process queue", processQueue)
}
