package commands

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tasklog"
	"github.com/git-lfs/git-lfs/tq"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

func pullCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()
	setupRepository()

	if len(args) > 0 {
		// Remote is first arg
		if err := cfg.SetValidRemote(args[0]); err != nil {
			Exit("Invalid remote name %q: %s", args[0], err)
		}
	}

	includeArg, excludeArg := getIncludeExcludeArgs(cmd)
	filter := buildFilepathFilter(cfg, includeArg, excludeArg, true)
	pull(filter)
}

func pull(filter *filepathfilter.Filter) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not pull")
	}

	pointers := newPointerMap()
	logger := tasklog.NewLogger(os.Stdout,
		tasklog.ForceProgress(cfg.ForceProgress()),
	)
	meter := tq.NewMeter(cfg)
	meter.Logger = meter.LoggerFromEnv(cfg.Os)
	logger.Enqueue(meter)
	remote := cfg.Remote()
	singleCheckout := newSingleCheckout(cfg.Git, remote)
	q := newDownloadQueue(singleCheckout.Manifest(), remote, tq.WithProgress(meter))
	gitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			LoggedError(err, "Scanner error: %s", err)
			return
		}

		if pointers.Seen(p) {
			return
		}

		// no need to download objects that exist locally already
		lfs.LinkOrCopyFromReference(cfg, p.Oid, p.Size)
		if cfg.LFSObjectExists(p.Oid, p.Size) {
			singleCheckout.Run(p)
			return
		}

		meter.Add(p.Size)
		tracerx.Printf("fetch %v [%v]", p.Name, p.Oid)
		pointers.Add(p)
		q.Add(downloadTransfer(p))
	})

	gitscanner.Filter = filter

	dlwatch := q.Watch()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for t := range dlwatch {
			for _, p := range pointers.All(t.Oid) {
				singleCheckout.Run(p)
			}
		}
		wg.Done()
	}()

	processQueue := time.Now()
	if err := gitscanner.ScanTree(ref.Sha); err != nil {
		singleCheckout.Close()
		ExitWithError(err)
	}

	meter.Start()
	gitscanner.Close()
	q.Wait()
	wg.Wait()
	tracerx.PerformanceSince("process queue", processQueue)

	singleCheckout.Close()

	success := true
	for _, err := range q.Errors() {
		success = false
		FullError(err)
	}

	if !success {
		c := getAPIClient()
		e := c.Endpoints.Endpoint("download", remote)
		Exit("error: failed to fetch some objects from '%s'", e.Url)
	}

	if singleCheckout.Skip() {
		fmt.Println("Skipping object checkout, Git LFS is not installed.")
	}
}

// tracks LFS objects being downloaded, according to their unique OIDs.
type pointerMap struct {
	pointers map[string][]*lfs.WrappedPointer
	mu       sync.Mutex
}

func newPointerMap() *pointerMap {
	return &pointerMap{pointers: make(map[string][]*lfs.WrappedPointer)}
}

func (m *pointerMap) Seen(p *lfs.WrappedPointer) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.pointers[p.Oid]; ok {
		m.pointers[p.Oid] = append(existing, p)
		return true
	}
	return false
}

func (m *pointerMap) Add(p *lfs.WrappedPointer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pointers[p.Oid] = append(m.pointers[p.Oid], p)
}

func (m *pointerMap) All(oid string) []*lfs.WrappedPointer {
	m.mu.Lock()
	defer m.mu.Unlock()
	pointers := m.pointers[oid]
	delete(m.pointers, oid)
	return pointers
}

func init() {
	RegisterCommand("pull", pullCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")
	})
}
