package lfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/git-lfs/git-lfs/tools"
)

type cleanedAsset struct {
	Filename string
	*Pointer
}

func PointerClean(reader io.Reader, fileName string, fileSize int64, cb progress.CopyCallback) (*cleanedAsset, error) {
	extensions, err := config.Config.SortedExtensions()
	if err != nil {
		return nil, err
	}

	var oid string
	var size int64
	var tmp *os.File
	var exts []*PointerExtension
	if len(extensions) > 0 {
		request := &pipeRequest{"clean", reader, fileName, extensions}

		var response pipeResponse
		if response, err = pipeExtensions(request); err != nil {
			return nil, err
		}

		oid = response.results[len(response.results)-1].oidOut
		tmp = response.file
		var stat os.FileInfo
		if stat, err = os.Stat(tmp.Name()); err != nil {
			return nil, err
		}
		size = stat.Size()

		for _, result := range response.results {
			if result.oidIn != result.oidOut {
				ext := NewPointerExtension(result.name, len(exts), result.oidIn)
				exts = append(exts, ext)
			}
		}
	} else {
		oid, size, tmp, err = copyToTemp(reader, fileSize, cb)
		if err != nil {
			return nil, err
		}
	}

	pointer := NewPointer(oid, size, exts)
	return &cleanedAsset{tmp.Name(), pointer}, err
}

func copyToTemp(reader io.Reader, fileSize int64, cb progress.CopyCallback) (oid string, size int64, tmp *os.File, err error) {
	tmp, err = TempFile("")
	if err != nil {
		return
	}

	defer tmp.Close()

	oidHash := sha256.New()
	writer := io.MultiWriter(oidHash, tmp)

	if fileSize == 0 {
		cb = nil
	}

	ptr, buf, err := DecodeFrom(reader)

	by := make([]byte, blobSizeCutoff)
	n, rerr := buf.Read(by)
	by = by[:n]

	if rerr != nil || (err == nil && len(by) < 512) {
		err = errors.NewCleanPointerError(ptr, by)
		return
	}

	var from io.Reader = bytes.NewReader(by)
	if int64(len(by)) < fileSize {
		// If there is still more data to be read from the file, tack on
		// the original reader and continue the read from there.
		from = io.MultiReader(from, reader)
	}

	size, err = tools.CopyWithCallback(writer, from, fileSize, cb)

	if err != nil {
		return
	}

	oid = hex.EncodeToString(oidHash.Sum(nil))
	return
}

func (a *cleanedAsset) Teardown() error {
	return os.Remove(a.Filename)
}
