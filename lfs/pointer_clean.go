package lfs

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

type cleanedAsset struct {
	File          *os.File
	mediafilepath string
	*Pointer
}

func PointerClean(reader io.Reader, size int64, cb CopyCallback) (*cleanedAsset, error) {
	tmp, err := TempFile("")
	if err != nil {
		return nil, err
	}

	oidHash := sha256.New()
	writer := io.MultiWriter(oidHash, tmp)

	if size == 0 {
		cb = nil
	}

	written, err := CopyWithCallback(writer, reader, size, cb)

	pointer := NewPointer(hex.EncodeToString(oidHash.Sum(nil)), written)
	return &cleanedAsset{tmp, "", pointer}, err
}

func (a *cleanedAsset) Close() error {
	return a.File.Close()
}

func (a *cleanedAsset) Teardown() error {
	return os.Remove(a.File.Name())
}
