package tq

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

// Meter provides a progress bar type output for the TransferQueue. It
// is given an estimated file count and size up front and tracks the number of
// files and bytes transferred as well as the number of files and bytes that
// get skipped because the transfer is unnecessary.
type Meter struct {
	DryRun bool
	Logger *tools.SyncWriter

	finishedFiles   int64 // int64s must come first for struct alignment
	skippedFiles    int64
	estimatedBytes  int64
	currentBytes    int64
	skippedBytes    int64
	estimatedFiles  int32
	paused          uint32
	fileIndex       map[string]int64 // Maps a file name to its transfer number
	oidCurrentBytes map[string]int64 // Mapos OID to its current transferred bytes
	fileIndexMutex  *sync.Mutex
	updates         chan *tasklog.Update
}

type env interface {
	Get(key string) (val string, ok bool)
}

func (m *Meter) LoggerFromEnv(os env) *tools.SyncWriter {
	name, _ := os.Get("GIT_LFS_PROGRESS")
	if len(name) < 1 {
		return nil
	}
	return m.LoggerToFile(name)
}

func (m *Meter) LoggerToFile(name string) *tools.SyncWriter {
	printErr := func(err string) {
		fmt.Fprintf(os.Stderr, "Error creating progress logger: %s\n", err)
	}

	if !filepath.IsAbs(name) {
		printErr("GIT_LFS_PROGRESS must be an absolute path")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		printErr(err.Error())
		return nil
	}

	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		printErr(err.Error())
		return nil
	}

	return tools.NewSyncWriter(file)
}

// NewMeter creates a new Meter.
func NewMeter() *Meter {
	m := &Meter{
		oidCurrentBytes: make(map[string]int64),
		fileIndex:       make(map[string]int64),
		fileIndexMutex:  &sync.Mutex{},
		updates:         make(chan *tasklog.Update),
	}

	return m
}

// Start begins sending status updates to the optional log file, and stdout.
func (m *Meter) Start() {
	if m == nil {
		return
	}
	atomic.StoreUint32(&m.paused, 0)
}

// Pause stops sending status updates temporarily, until Start() is called again.
func (m *Meter) Pause() {
	if m == nil {
		return
	}
	atomic.StoreUint32(&m.paused, 1)
}

// Add tells the progress meter that a single file of the given size will
// possibly be transferred. If a file doesn't need to be transferred for some
// reason, be sure to call Skip(int64) with the same size.
func (m *Meter) Add(size int64) {
	if m == nil {
		return
	}

	defer m.update()
	atomic.AddInt32(&m.estimatedFiles, 1)
	atomic.AddInt64(&m.estimatedBytes, size)
}

// Skip tells the progress meter that a file of size `size` is being skipped
// because the transfer is unnecessary.
func (m *Meter) Skip(size int64) {
	if m == nil {
		return
	}

	defer m.update()
	atomic.AddInt64(&m.skippedFiles, 1)
	atomic.AddInt64(&m.skippedBytes, size)
	// Reduce bytes and files so progress easier to parse
	atomic.AddInt32(&m.estimatedFiles, -1)
	atomic.AddInt64(&m.estimatedBytes, -size)
}

// StartTransfer tells the progress meter that a transferring file is being
// added to the TransferQueue.
func (m *Meter) StartTransfer(name, oid string) {
	if m == nil {
		return
	}

	defer m.update()
	m.fileIndexMutex.Lock()
	m.oidCurrentBytes[oid] = int64(0)
	m.fileIndex[name] = int64(len(m.oidCurrentBytes))
	m.fileIndexMutex.Unlock()
}

// TransferBytes increments the number of bytes transferred
func (m *Meter) TransferBytes(direction, name, oid string, read, total int64, current int) {
	if m == nil {
		return
	}

	defer m.update()
	m.fileIndexMutex.Lock()
	m.oidCurrentBytes[oid] = read
	m.fileIndexMutex.Unlock()
	m.logBytes(direction, name, read, total)
}

// FinishTransfer increments the finished transfer count
func (m *Meter) FinishTransfer(name string) {
	if m == nil {
		return
	}

	defer m.update()
	atomic.AddInt64(&m.finishedFiles, 1)
	m.fileIndexMutex.Lock()
	delete(m.fileIndex, name)
	m.fileIndexMutex.Unlock()
}

// Finish shuts down the Meter.
func (m *Meter) Finish() {
	if m == nil {
		return
	}

	m.update()
	close(m.updates)
}

func (m *Meter) Updates() <-chan *tasklog.Update {
	if m == nil {
		return nil
	}
	return m.updates
}

func (m *Meter) Throttled() bool {
	return true
}

func (m *Meter) update() {
	if m.skipUpdate() {
		return
	}

	m.updates <- &tasklog.Update{
		S:  m.str(),
		At: time.Now(),
	}
}

func (m *Meter) skipUpdate() bool {
	return m.DryRun ||
		(m.estimatedFiles == 0 && m.skippedFiles == 0) ||
		atomic.LoadUint32(&m.paused) == 1
}

func (m *Meter) str() string {
	// (%d of %d files, %d skipped) %f B / %f B, %f B skipped
	// skipped counts only show when > 0

	current := int64(0)
	m.fileIndexMutex.Lock()
	for _, read := range m.oidCurrentBytes {
		current += read
	}
	m.fileIndexMutex.Unlock()

	out := fmt.Sprintf("\rGit LFS: (%d of %d files",
		m.finishedFiles,
		m.estimatedFiles)
	if m.skippedFiles > 0 {
		out += fmt.Sprintf(", %d skipped", m.skippedFiles)
	}
	out += fmt.Sprintf(") %s / %s",
		humanize.FormatBytes(uint64(current)),
		humanize.FormatBytes(uint64(m.estimatedBytes)))
	if m.skippedBytes > 0 {
		out += fmt.Sprintf(", %s skipped",
			humanize.FormatBytes(uint64(m.skippedBytes)))
	}

	return out
}

func (m *Meter) logBytes(direction, name string, read, total int64) {
	m.fileIndexMutex.Lock()
	idx := m.fileIndex[name]
	logger := m.Logger
	m.fileIndexMutex.Unlock()
	if logger == nil {
		return
	}

	line := fmt.Sprintf("%s %d/%d %d/%d %s\n", direction, idx, m.estimatedFiles, read, total, name)
	if err := m.Logger.Write([]byte(line)); err != nil {
		m.fileIndexMutex.Lock()
		m.Logger = nil
		m.fileIndexMutex.Unlock()
	}
}
