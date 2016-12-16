package lfs

import (
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/tq"
)

// NewUploadQueue builds an UploadQueue, allowing `workers` concurrent uploads.
func NewUploadQueue(cfg *config.Configuration, options ...tq.Option) *tq.TransferQueue {
	return tq.NewTransferQueue(tq.Upload, TransferManifest(cfg), options...)
}
