package progress

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olekukonko/ts"
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
	started           int32
	estimatedFiles    int32
	startTime         time.Time
	finished          chan interface{}
	logger            *progressLogger
	fileIndex         map[string]int64 // Maps a file name to its transfer number
	fileIndexMutex    *sync.Mutex
	dryRun            bool
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

		m.logger.writeData = true
		m.logger.log = file
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
		logger:         &progressLogger{},
		startTime:      time.Now(),
		fileIndex:      make(map[string]int64),
		fileIndexMutex: &sync.Mutex{},
		finished:       make(chan interface{}),
	}

	for _, opt := range options {
		opt(m)
	}

	return m
}

// Start begins sending status updates to the optional log file, and stdout.
func (p *ProgressMeter) Start() {
	if atomic.SwapInt32(&p.started, 1) == 0 {
		go p.writer()
	}
}

// Add tells the progress meter that a single file of the given size will
// possibly be transferred. If a file doesn't need to be transferred for some
// reason, be sure to call Skip(int64) with the same size.
func (p *ProgressMeter) Add(size int64) {
	atomic.AddInt32(&p.estimatedFiles, 1)
	atomic.AddInt64(&p.estimatedBytes, size)
}

// Skip tells the progress meter that a file of size `size` is being skipped
// because the transfer is unnecessary.
func (p *ProgressMeter) Skip(size int64) {
	atomic.AddInt64(&p.skippedFiles, 1)
	atomic.AddInt64(&p.skippedBytes, size)
	// Reduce bytes and files so progress easier to parse
	atomic.AddInt32(&p.estimatedFiles, -1)
	atomic.AddInt64(&p.estimatedBytes, -size)
}

// StartTransfer tells the progress meter that a transferring file is being
// added to the TransferQueue.
func (p *ProgressMeter) StartTransfer(name string) {
	idx := atomic.AddInt64(&p.transferringFiles, 1)
	p.fileIndexMutex.Lock()
	p.fileIndex[name] = idx
	p.fileIndexMutex.Unlock()
}

// TransferBytes increments the number of bytes transferred
func (p *ProgressMeter) TransferBytes(direction, name string, read, total int64, current int) {
	atomic.AddInt64(&p.currentBytes, int64(current))
	p.logBytes(direction, name, read, total)
}

// FinishTransfer increments the finished transfer count
func (p *ProgressMeter) FinishTransfer(name string) {
	atomic.AddInt64(&p.finishedFiles, 1)
	p.fileIndexMutex.Lock()
	delete(p.fileIndex, name)
	p.fileIndexMutex.Unlock()
}

// Finish shuts down the ProgressMeter
func (p *ProgressMeter) Finish() {
	close(p.finished)
	p.update()
	p.logger.Close()
	if !p.dryRun && p.estimatedBytes > 0 {
		fmt.Fprintf(os.Stdout, "\n")
	}
}

func (p *ProgressMeter) logBytes(direction, name string, read, total int64) {
	p.fileIndexMutex.Lock()
	idx := p.fileIndex[name]
	p.fileIndexMutex.Unlock()
	line := fmt.Sprintf("%s %d/%d %d/%d %s\n", direction, idx, p.estimatedFiles, read, total, name)
	if err := p.logger.Write([]byte(line)); err != nil {
		p.logger.Shutdown()
	}
}

func (p *ProgressMeter) writer() {
	p.update()
	for {
		select {
		case <-p.finished:
			return
		case <-time.After(time.Millisecond * 200):
			p.update()
		}
	}
}

func (p *ProgressMeter) update() {
	if p.dryRun || (p.estimatedFiles == 0 && p.skippedFiles == 0) {
		return
	}

	// (%d of %d files, %d skipped) %f B / %f B, %f B skipped
	// skipped counts only show when > 0

	out := fmt.Sprintf("\rGit LFS: (%d of %d files", p.finishedFiles, p.estimatedFiles)
	if p.skippedFiles > 0 {
		out += fmt.Sprintf(", %d skipped", p.skippedFiles)
	}
	out += fmt.Sprintf(") %s / %s", formatBytes(p.currentBytes), formatBytes(p.estimatedBytes))
	if p.skippedBytes > 0 {
		out += fmt.Sprintf(", %s skipped", formatBytes(p.skippedBytes))
	}

	fmt.Fprintf(os.Stdout, pad(out))
}

func formatBytes(i int64) string {
	switch {
	case i > 1099511627776:
		return fmt.Sprintf("%#0.2f TB", float64(i)/1099511627776)
	case i > 1073741824:
		return fmt.Sprintf("%#0.2f GB", float64(i)/1073741824)
	case i > 1048576:
		return fmt.Sprintf("%#0.2f MB", float64(i)/1048576)
	case i > 1024:
		return fmt.Sprintf("%#0.2f KB", float64(i)/1024)
	}

	return fmt.Sprintf("%d B", i)
}

const defaultWidth = 80

// pad pads the given message to occupy the entire maximum width of the terminal
// LFS is attached to. In doing so, this safeguards subsequent prints of shorter
// messages from leaving stray characters from the previous message on the
// screen by writing over them with whitespace padding.
func pad(msg string) string {
	width := defaultWidth
	size, err := ts.GetSize()
	if err == nil {
		// If `ts.GetSize()` was successful, set the width to the number
		// of columns present in the terminal LFS is attached to.
		// Otherwise, fall-back to `defaultWidth`.
		width = size.Col()
	}

	// Pad the string with whitespace so that printing at the start of the
	// line removes all traces from the last print.removes all traces from
	// the last print.
	padding := strings.Repeat(" ", maxInt(0, width-len(msg)))

	return msg + padding
}

// maxInt returns the greater of two `int`s, "a", or "b". This function
// originally comes from `github.com/git-lfs/git-lfs/tools#MaxInt`, but would
// introduce an import cycle if depended on directly.
func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
