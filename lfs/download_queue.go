package lfs

import (
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/tq"
)

func NewDownloadable(p *WrappedPointer) *tq.Transfer {
	path, _ := LocalMediaPath(p.Oid)

	return &tq.Transfer{
		Oid:  p.Oid,
		Size: p.Size,
		Name: p.Name,
		Path: path,
	}
}

// NewDownloadCheckQueue builds a checking queue, checks that objects are there but doesn't download
func NewDownloadCheckQueue(cfg *config.Configuration, options ...tq.Option) *tq.TransferQueue {
	allOptions := make([]tq.Option, len(options), len(options)+1)
	allOptions = append(allOptions, options...)
	allOptions = append(allOptions, tq.DryRun(true))
	return NewDownloadQueue(cfg, allOptions...)
}

// NewDownloadQueue builds a DownloadQueue, allowing concurrent downloads.
func NewDownloadQueue(cfg *config.Configuration, options ...tq.Option) *tq.TransferQueue {
	return tq.NewTransferQueue(tq.Download, TransferManifest(cfg), options...)
}
