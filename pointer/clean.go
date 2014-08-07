package pointer

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/github/git-media/gitmedia"
	"io"
	"os"
)

type cleanedAsset struct {
	File          *os.File
	mediafilepath string
	*Pointer
}

func Clean(reader io.Reader, size int64, cb gitmedia.CopyCallback) (*cleanedAsset, error) {
	tmp, err := gitmedia.TempFile("")
	if err != nil {
		return nil, err
	}

	oidHash := sha256.New()
	writer := io.MultiWriter(oidHash, tmp)

	if size == 0 {
		cb = nil
	}

	written, err := gitmedia.CopyWithCallback(writer, reader, size, cb)

	pointer := NewPointer(hex.EncodeToString(oidHash.Sum(nil)), written)
	return &cleanedAsset{tmp, "", pointer}, err
}

func (a *cleanedAsset) Close() error {
	return os.Remove(a.File.Name())
}
