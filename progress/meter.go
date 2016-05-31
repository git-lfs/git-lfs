package progress

import (
	"fmt"
	"os"
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
	estimatedFiles    int
	startTime         time.Time
	finished          chan interface{}
	logger            *progressLogger
	fileIndex         map[string]int64 // Maps a file name to its transfer number
	fileIndexMutex    *sync.Mutex
	dryRun            bool
}

// NewProgressMeter creates a new ProgressMeter for the number and size of
// files given.
func NewProgressMeter(estFiles int, estBytes int64, dryRun bool, logPath string) *ProgressMeter {
	logger, err := newProgressLogger(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating progress logger: %s\n", err)
	}

	return &ProgressMeter{
		logger:         logger,
		startTime:      time.Now(),
		fileIndex:      make(map[string]int64),
		fileIndexMutex: &sync.Mutex{},
		finished:       make(chan interface{}),
		estimatedFiles: estFiles,
		estimatedBytes: estBytes,
		dryRun:         dryRun,
	}
}

func (p *ProgressMeter) Start() {
	if atomic.SwapInt32(&p.started, 1) == 0 {
		go p.writer()
	}
}

// Add tells the progress meter that a transferring file is being added to the
// TransferQueue.
func (p *ProgressMeter) Add(name string) {
	idx := atomic.AddInt64(&p.transferringFiles, 1)
	p.fileIndexMutex.Lock()
	p.fileIndex[name] = idx
	p.fileIndexMutex.Unlock()
}

// Skip tells the progress meter that a file of size `size` is being skipped
// because the transfer is unnecessary.
func (p *ProgressMeter) Skip(size int64) {
	atomic.AddInt64(&p.skippedFiles, 1)
	atomic.AddInt64(&p.skippedBytes, size)
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
	if p.dryRun || p.estimatedFiles == 0 {
		return
	}

	width := 80 // default to 80 chars wide if ts.GetSize() fails
	size, err := ts.GetSize()
	if err == nil {
		width = size.Col()
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

	padlen := width - len(out)
	if 0 < padlen {
		out += strings.Repeat(" ", padlen)
	}

	fmt.Fprintf(os.Stdout, out)
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
