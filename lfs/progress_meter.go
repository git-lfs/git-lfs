package lfs

import (
	"fmt"

	"github.com/github/git-lfs/vendor/_nuts/github.com/cheggaaa/pb"
)

type ProgressTracker interface {
	Log(event progressEvent, name string, read, total int64, current int)
	Finish()
}

type ProgressMeter struct {
	totalBytes   int64
	totalFiles   int
	startedFiles int
	bar          *pb.ProgressBar
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

	return &ProgressMeter{
		totalBytes: bytes,
		totalFiles: files,
		bar:        bar,
	}
}

func (p *ProgressMeter) Log(event progressEvent, name string, read, total int64, current int) {
	switch event {
	case transferStart:
		p.startedFiles++
	case transferBytes:
		p.bar.Add(current)
	}

	p.bar.Prefix(fmt.Sprintf("(%d of %d files) ", p.startedFiles, p.totalFiles))
}

func (p *ProgressMeter) Finish() {
	p.bar.Finish()
}
