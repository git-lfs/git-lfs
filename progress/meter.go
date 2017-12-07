package progress

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/git-lfs/git-lfs/tasklog"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tools/humanize"
)

// ProgressMeter provides a progress bar type output for the TransferQueue. It
// is given an estimated file count and size up front and tracks the number of
// files and bytes transferred as well as the number of files and bytes that
// get skipped because the transfer is unnecessary.
type ProgressMeter struct {
	finishedFiles     int64 // int64s must come first for struct alignment
	skippedFiles      int64
	transferringFiles int64
	estimatedBytes    int64
	currentBytes      int64
	skippedBytes      int64
	estimatedFiles    int32
	paused            uint32
	logToFile         uint32
	logger            *tools.SyncWriter
	fileIndex         map[string]int64 // Maps a file name to its transfer number
	fileIndexMutex    *sync.Mutex
	dryRun            bool
	updates           chan *tasklog.Update
}

type env interface {
	Get(key string) (val string, ok bool)
}

type meterOption func(*ProgressMeter)

// DryRun is an option for NewMeter() that determines whether updates should be
// sent to stdout.
func DryRun(dryRun bool) meterOption {
	return func(m *ProgressMeter) {
		m.dryRun = dryRun
	}
}

// WithLogFile is an option for NewMeter() that sends updates to a text file.
func WithLogFile(name string) meterOption {
	printErr := func(err string) {
		fmt.Fprintf(os.Stderr, "Error creating progress logger: %s\n", err)
	}

	return func(m *ProgressMeter) {
		if len(name) == 0 {
			return
		}

		if !filepath.IsAbs(name) {
			printErr("GIT_LFS_PROGRESS must be an absolute path")
			return
		}

		cbDir := filepath.Dir(name)
		if err := os.MkdirAll(cbDir, 0755); err != nil {
			printErr(err.Error())
			return
		}

		file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			printErr(err.Error())
			return
		}

		m.logToFile = 1
		m.logger = tools.NewSyncWriter(file)
	}
}

// WithOSEnv is an option for NewMeter() that sends updates to the text file
// path specified in the OS Env.
func WithOSEnv(os env) meterOption {
	name, _ := os.Get("GIT_LFS_PROGRESS")
	return WithLogFile(name)
}

// NewMeter creates a new ProgressMeter.
func NewMeter(options ...meterOption) *ProgressMeter {
	m := &ProgressMeter{
		fileIndex:      make(map[string]int64),
		fileIndexMutex: &sync.Mutex{},
		updates:        make(chan *tasklog.Update),
	}

	for _, opt := range options {
		opt(m)
	}

	return m
}

// Start begins sending status updates to the optional log file, and stdout.
func (p *ProgressMeter) Start() {
	atomic.StoreUint32(&p.paused, 0)
}

// Pause stops sending status updates temporarily, until Start() is called again.
func (p *ProgressMeter) Pause() {
	atomic.StoreUint32(&p.paused, 1)
}

// Add tells the progress meter that a single file of the given size will
// possibly be transferred. If a file doesn't need to be transferred for some
// reason, be sure to call Skip(int64) with the same size.
func (p *ProgressMeter) Add(size int64) {
	defer p.update()
	atomic.AddInt32(&p.estimatedFiles, 1)
	atomic.AddInt64(&p.estimatedBytes, size)
}

// Skip tells the progress meter that a file of size `size` is being skipped
// because the transfer is unnecessary.
func (p *ProgressMeter) Skip(size int64) {
	defer p.update()
	atomic.AddInt64(&p.skippedFiles, 1)
	atomic.AddInt64(&p.skippedBytes, size)
	// Reduce bytes and files so progress easier to parse
	atomic.AddInt32(&p.estimatedFiles, -1)
	atomic.AddInt64(&p.estimatedBytes, -size)
}

// StartTransfer tells the progress meter that a transferring file is being
// added to the TransferQueue.
func (p *ProgressMeter) StartTransfer(name string) {
	defer p.update()
	idx := atomic.AddInt64(&p.transferringFiles, 1)
	p.fileIndexMutex.Lock()
	p.fileIndex[name] = idx
	p.fileIndexMutex.Unlock()
}

// TransferBytes increments the number of bytes transferred
func (p *ProgressMeter) TransferBytes(direction, name string, read, total int64, current int) {
	defer p.update()
	atomic.AddInt64(&p.currentBytes, int64(current))
	p.logBytes(direction, name, read, total)
}

// FinishTransfer increments the finished transfer count
func (p *ProgressMeter) FinishTransfer(name string) {
	defer p.update()
	atomic.AddInt64(&p.finishedFiles, 1)
	p.fileIndexMutex.Lock()
	delete(p.fileIndex, name)
	p.fileIndexMutex.Unlock()
}

// Sync sends an update now, if necessary
func (p *ProgressMeter) Sync() {
	p.update()
}

// Finish shuts down the ProgressMeter
func (p *ProgressMeter) Finish() {
	p.update()
	close(p.updates)
}

func (p *ProgressMeter) Updates() <-chan *tasklog.Update {
	return p.updates
}

func (p *ProgressMeter) Throttled() bool {
	return true
}

func (p *ProgressMeter) update() {
	if p.skipUpdate() {
		return
	}

	p.updates <- &tasklog.Update{
		S:  p.str(),
		At: time.Now(),
	}
}

func (p *ProgressMeter) skipUpdate() bool {
	return p.dryRun ||
		(p.estimatedFiles == 0 && p.skippedFiles == 0) ||
		atomic.LoadUint32(&p.paused) == 1
}

func (p *ProgressMeter) str() string {
	// (%d of %d files, %d skipped) %f B / %f B, %f B skipped
	// skipped counts only show when > 0

	out := fmt.Sprintf("\rGit LFS: (%d of %d files",
		p.finishedFiles,
		p.estimatedFiles)
	if p.skippedFiles > 0 {
		out += fmt.Sprintf(", %d skipped", p.skippedFiles)
	}
	out += fmt.Sprintf(") %s / %s",
		humanize.FormatBytes(uint64(p.currentBytes)),
		humanize.FormatBytes(uint64(p.estimatedBytes)))
	if p.skippedBytes > 0 {
		out += fmt.Sprintf(", %s skipped",
			humanize.FormatBytes(uint64(p.skippedBytes)))
	}

	return out
}

func (p *ProgressMeter) logBytes(direction, name string, read, total int64) {
	p.fileIndexMutex.Lock()
	idx := p.fileIndex[name]
	p.fileIndexMutex.Unlock()
	line := fmt.Sprintf("%s %d/%d %d/%d %s\n", direction, idx, p.estimatedFiles, read, total, name)
	if atomic.LoadUint32(&p.logToFile) == 1 {
		if err := p.logger.Write([]byte(line)); err != nil {
			atomic.StoreUint32(&p.logToFile, 0)
		}
	}
}
