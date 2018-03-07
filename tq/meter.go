package tq

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
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
	finishedFiles     int64 // int64s must come first for struct alignment
	transferringFiles int64
	estimatedBytes    int64
	lastBytes         int64
	currentBytes      int64
	sampleCount       uint64
	avgBytes          float64
	lastAvg           time.Time
	estimatedFiles    int32
	paused            uint32
	fileIndex         map[string]int64 // Maps a file name to its transfer number
	fileIndexMutex    *sync.Mutex
	updates           chan *tasklog.Update

	DryRun    bool
	Logger    *tools.SyncWriter
	Direction Direction
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
		fileIndex:      make(map[string]int64),
		fileIndexMutex: &sync.Mutex{},
		updates:        make(chan *tasklog.Update),
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

	defer m.update(false)
	atomic.AddInt32(&m.estimatedFiles, 1)
	atomic.AddInt64(&m.estimatedBytes, size)
}

// Skip tells the progress meter that a file of size `size` is being skipped
// because the transfer is unnecessary.
func (m *Meter) Skip(size int64) {
	if m == nil {
		return
	}

	defer m.update(false)
	atomic.AddInt64(&m.finishedFiles, 1)
	atomic.AddInt64(&m.currentBytes, size)
}

// StartTransfer tells the progress meter that a transferring file is being
// added to the TransferQueue.
func (m *Meter) StartTransfer(name string) {
	if m == nil {
		return
	}

	defer m.update(false)
	idx := atomic.AddInt64(&m.transferringFiles, 1)
	m.fileIndexMutex.Lock()
	m.fileIndex[name] = idx
	m.fileIndexMutex.Unlock()
}

// TransferBytes increments the number of bytes transferred
func (m *Meter) TransferBytes(direction, name string, read, total int64, current int) {
	if m == nil {
		return
	}

	defer m.update(false)

	now := time.Now()
	since := now.Sub(m.lastAvg)
	atomic.AddInt64(&m.currentBytes, int64(current))
	atomic.AddInt64(&m.lastBytes, int64(current))

	if since > time.Second {
		m.lastAvg = now

		bps := float64(m.lastBytes) / since.Seconds()

		m.avgBytes = (m.avgBytes*float64(m.sampleCount) + bps) / (float64(m.sampleCount) + 1.0)

		atomic.StoreInt64(&m.lastBytes, 0)
		atomic.AddUint64(&m.sampleCount, 1)
	}

	m.logBytes(direction, name, read, total)
}

// FinishTransfer increments the finished transfer count
func (m *Meter) FinishTransfer(name string) {
	if m == nil {
		return
	}

	defer m.update(false)
	atomic.AddInt64(&m.finishedFiles, 1)
	m.fileIndexMutex.Lock()
	delete(m.fileIndex, name)
	m.fileIndexMutex.Unlock()
}

// Flush sends the latest progress update, while leaving the meter active.
func (m *Meter) Flush() {
	if m == nil {
		return
	}

	m.update(true)
}

// Finish shuts down the Meter.
func (m *Meter) Finish() {
	if m == nil {
		return
	}

	m.update(false)
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

func (m *Meter) update(force bool) {
	if m.skipUpdate() {
		return
	}

	m.updates <- &tasklog.Update{
		S:     m.str(),
		At:    time.Now(),
		Force: force,
	}
}

func (m *Meter) skipUpdate() bool {
	return m.DryRun ||
		m.estimatedFiles == 0 ||
		atomic.LoadUint32(&m.paused) == 1
}

func (m *Meter) str() string {
	// (Uploading|Downloading) LFS objects: 100% (10/10) 100 MiB | 10 MiB/s

	direction := strings.Title(m.Direction.String()) + "ing"
	percentage := 100 * float64(m.finishedFiles) / float64(m.estimatedFiles)

	return fmt.Sprintf("%s LFS objects: %3.f%% (%d/%d), %s | %s",
		direction,
		percentage,
		m.finishedFiles, m.estimatedFiles,
		humanize.FormatBytes(clamp(m.currentBytes)),
		humanize.FormatByteRate(clampf(m.avgBytes), time.Second))
}

// clamp clamps the given "x" within the acceptable domain of the uint64 integer
// type, so as to prevent over- and underflow.
func clamp(x int64) uint64 {
	if x < 0 {
		return 0
	}
	if x > math.MaxInt64 {
		return math.MaxUint64
	}
	return uint64(x)
}

func clampf(x float64) uint64 {
	if x < 0 {
		return 0
	}
	if x > math.MaxUint64 {
		return math.MaxUint64
	}
	return uint64(x)
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
