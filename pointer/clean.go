package pointer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"github.com/github/git-lfs/lfs"
	"io"
	"os"
)

type cleanedAsset struct {
	File          *os.File
	mediafilepath string
	*Pointer
}

type CleanedPointerError struct {
	Bytes []byte
}

func (e *CleanedPointerError) Error() string {
	return "Cannot clean a Git LFS pointer.  Skipping."
}

func Clean(reader io.Reader, size int64, cb lfs.CopyCallback) (*cleanedAsset, error) {
	tmp, err := lfs.TempFile("")
	if err != nil {
		return nil, err
	}

	oidHash := sha256.New()
	writer := io.MultiWriter(oidHash, tmp)

	if size == 0 {
		cb = nil
	}

	by, _, err := DecodeFrom(reader)
	if err == nil && len(by) < 512 {
		return nil, &CleanedPointerError{by}
	}

	multi := io.MultiReader(bytes.NewReader(by), reader)
	written, err := lfs.CopyWithCallback(writer, multi, size, cb)

	pointer := NewPointer(hex.EncodeToString(oidHash.Sum(nil)), written)
	return &cleanedAsset{tmp, "", pointer}, err
}

func (a *cleanedAsset) Close() error {
	return a.File.Close()
}

func (a *cleanedAsset) Teardown() error {
	return os.Remove(a.File.Name())
}
