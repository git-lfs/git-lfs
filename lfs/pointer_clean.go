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

func PointerClean(reader io.Reader, size int64, cb CopyCallback) (*cleanedAsset, error) {
	tmp, err := TempFile("")
	if err != nil {
		return nil, err
	}

	defer tmp.Close()

	oidHash := sha256.New()
	writer := io.MultiWriter(oidHash, tmp)

	if size == 0 {
		cb = nil
	}

	by, ptr, err := DecodeFrom(reader)
	if err == nil && len(by) < 512 {
		return nil, &CleanedPointerError{ptr, by}
	}

	multi := io.MultiReader(bytes.NewReader(by), reader)
	written, err := CopyWithCallback(writer, multi, size, cb)

	pointer := NewPointer(hex.EncodeToString(oidHash.Sum(nil)), written)
	return &cleanedAsset{tmp.Name(), pointer}, err
}

func (a *cleanedAsset) Teardown() error {
	return os.Remove(a.Filename)
}
