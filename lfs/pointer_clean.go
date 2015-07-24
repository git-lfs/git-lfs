package lfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

type cleanedAsset struct {
	Filename string
	*Pointer
}

type CleanedPointerError struct {
	Pointer *Pointer
	Bytes   []byte
}

func (e *CleanedPointerError) Error() string {
	return "Cannot clean a Git LFS pointer.  Skipping."
}

func PointerClean(reader io.Reader, fileName string, fileSize int64, cb CopyCallback) (*cleanedAsset, error) {
	extensions, err := SortExtensions(Config.Extensions())
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

func copyToTemp(reader io.Reader, fileSize int64, cb CopyCallback) (oid string, size int64, tmp *os.File, err error) {
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

	by, ptr, err := DecodeFrom(reader)
	if err == nil && len(by) < 512 {
		err = &CleanedPointerError{ptr, by}
		return
	}

	multi := io.MultiReader(bytes.NewReader(by), reader)
	size, err = CopyWithCallback(writer, multi, fileSize, cb)

	if err != nil {
		return
	}

	oid = hex.EncodeToString(oidHash.Sum(nil))
	return
}

func (a *cleanedAsset) Teardown() error {
	return os.Remove(a.Filename)
}
