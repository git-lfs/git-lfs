package pointer

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/github/git-media/gitmedia"
	"io"
	"os"
)

type CleanedAsset struct {
	File          *os.File
	mediafilepath string
	*Pointer
}

func Clean(reader io.Reader) (*CleanedAsset, error) {
	tmp, err := gitmedia.TempFile()
	if err != nil {
		return nil, err
	}

	oidHash := sha256.New()
	writer := io.MultiWriter(oidHash, tmp)
	written, err := io.Copy(writer, reader)

	pointer := NewPointer(hex.EncodeToString(oidHash.Sum(nil)), written)
	return &CleanedAsset{tmp, "", pointer}, err
}

func (a *CleanedAsset) Close() error {
	return os.Remove(a.File.Name())
}
