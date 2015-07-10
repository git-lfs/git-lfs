package lfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/github/git-lfs/vendor/_nuts/github.com/cheggaaa/pb"
)

type ProgressMeter struct {
	totalBytes   int64
	startedFiles int64
	totalFiles   int
	bar          *pb.ProgressBar
	logger       *progressLogger
	fileIndex    map[string]int64 // Maps a file name to its transfer number
}

type progressEvent int

const (
	transferStart = iota
	transferBytes
	transferFinish
)

func NewProgressMeter(files int, bytes int64) *ProgressMeter {
	bar := pb.New64(bytes)
	bar.SetUnits(pb.U_BYTES)
	bar.ShowBar = false
	bar.Prefix(fmt.Sprintf("(0 of %d files) ", files))
	bar.Start()

	logger, err := newProgressLogger()
	if err != nil {
		// TODO display an error
	}

	return &ProgressMeter{
		totalBytes: bytes,
		totalFiles: files,
		bar:        bar,
		logger:     logger,
		fileIndex:  make(map[string]int64),
	}
}

func (p *ProgressMeter) Log(event progressEvent, direction, name string, read, total int64, current int) {
	switch event {
	case transferStart:
		idx := atomic.AddInt64(&p.startedFiles, 1)
		p.fileIndex[name] = idx
	case transferBytes:
		p.bar.Add(current)
		p.logBytes(direction, name, read, total)
	case transferFinish:
		delete(p.fileIndex, name)
	}

	p.bar.Prefix(fmt.Sprintf("(%d of %d files) ", p.startedFiles, p.totalFiles))
}

func (p *ProgressMeter) Finish() {
	p.bar.Finish()
	p.logger.Close()
}

func (p *ProgressMeter) logBytes(direction, name string, read, total int64) {
	idx := p.fileIndex[name]
	line := fmt.Sprintf("%s %d/%d %d/%d %s\n", direction, idx, p.totalFiles, read, total, name)
	if err := p.logger.Write([]byte(line)); err != nil {
		p.logger.Shutdown()
	}
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
