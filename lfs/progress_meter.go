package lfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/github/git-lfs/vendor/_nuts/github.com/olekukonko/ts"
)

type ProgressMeter struct {
	transferringFiles int64
	finishedFiles     int64
	totalFiles        int64
	skippedFiles      int64
	totalBytes        int64
	currentBytes      int64
	startTime         time.Time
	finished          chan interface{}
	logger            *progressLogger
	fileIndex         map[string]int64 // Maps a file name to its transfer number
	show              bool
}

type progressEvent int

const (
	transferStart = iota
	transferBytes
	transferFinish
)

func NewProgressMeter() *ProgressMeter {
	logger, err := newProgressLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating progress logger: %s\n", err)
	}

	pm := &ProgressMeter{
		logger:    logger,
		startTime: time.Now(),
		fileIndex: make(map[string]int64),
		finished:  make(chan interface{}),
		show:      true,
	}

	go pm.writer()

	return pm
}

func (p *ProgressMeter) Add(name string, size int64) {
	atomic.AddInt64(&p.totalBytes, size)
	idx := atomic.AddInt64(&p.transferringFiles, 1)
	p.fileIndex[name] = idx
}

func (p *ProgressMeter) Skip() {
	atomic.AddInt64(&p.skippedFiles, 1)
}

func (p *ProgressMeter) Log(event progressEvent, direction, name string, read, total int64, current int) {
	switch event {
	case transferBytes:
		atomic.AddInt64(&p.currentBytes, int64(current))
		p.logBytes(direction, name, read, total)
	case transferFinish:
		atomic.AddInt64(&p.finishedFiles, 1)
		delete(p.fileIndex, name)
	}
}

func (p *ProgressMeter) Finish() {
	close(p.finished)
	p.update()
	p.logger.Close()
	if p.show {
		fmt.Fprintf(os.Stdout, "\n")
	}
}

func (p *ProgressMeter) Suppress() {
	p.show = false
}

func (p *ProgressMeter) logBytes(direction, name string, read, total int64) {
	idx := p.fileIndex[name]
	line := fmt.Sprintf("%s %d/%d %d/%d %s\n", direction, idx, p.totalFiles, read, total, name)
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
	if !p.show {
		return
	}

	width := 80 // default to 80 chars wide if ts.GetSize() fails
	size, err := ts.GetSize()
	if err == nil {
		width = size.Col()
	}

	out := fmt.Sprintf("\r(%d of %d files), %s/%s",
		p.finishedFiles,
		p.transferringFiles,
		formatBytes(p.currentBytes),
		formatBytes(p.totalBytes))

	if skipped := atomic.LoadInt64(&p.skippedFiles); skipped > 0 {
		out += fmt.Sprintf(", Skipped: %d", skipped)
	}

	padding := strings.Repeat(" ", width-len(out))
	fmt.Fprintf(os.Stdout, out+padding)
}

// progressLogger provides a wrapper around an os.File that can either
// write to the file or ignore all writes completely.
type progressLogger struct {
	writeData bool
	log       *os.File
}

// Write will write to the file and perform a Sync() if writing succeeds.
func (l *progressLogger) Write(b []byte) error {
	if l.writeData {
		if _, err := l.log.Write(b); err != nil {
			return err
		}
		return l.log.Sync()
	}
	return nil
}

// Close will call Close() on the underlying file
func (l *progressLogger) Close() error {
	if l.log != nil {
		return l.log.Close()
	}
	return nil
}

// Shutdown will cause the logger to ignore any further writes. It should
// be used when writing causes an error.
func (l *progressLogger) Shutdown() {
	l.writeData = false
}

// newProgressLogger creates a progressLogger based on the presence of
// the GIT_LFS_PROGRESS environment variable. If it is present and a log file
// is able to be created, the logger will write to the file. If it is absent,
// or there is an err creating the file, the logger will ignore all writes.
func newProgressLogger() (*progressLogger, error) {
	logPath := Config.Getenv("GIT_LFS_PROGRESS")

	if len(logPath) == 0 {
		return &progressLogger{}, nil
	}
	if !filepath.IsAbs(logPath) {
		return &progressLogger{}, fmt.Errorf("GIT_LFS_PROGRESS must be an absolute path")
	}

	cbDir := filepath.Dir(logPath)
	if err := os.MkdirAll(cbDir, 0755); err != nil {
		return &progressLogger{}, err
	}

	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return &progressLogger{}, err
	}

	return &progressLogger{true, file}, nil
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
