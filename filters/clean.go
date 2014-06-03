package filters

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/github/git-media/gitmedia"
	"io"
	"os"
)

type CleanedAsset struct {
	Size          int64
	File          *os.File
	Sha           string
	mediafilepath string
}

func Clean(reader io.Reader) (*CleanedAsset, error) {
	tmp, err := gitmedia.TempFile()
	if err != nil {
		return nil, err
	}

	oidHash := sha256.New()
	writer := io.MultiWriter(oidHash, tmp)
	written, err := io.Copy(writer, reader)

	return &CleanedAsset{written, tmp, hex.EncodeToString(oidHash.Sum(nil)), ""}, err
}

func (a *CleanedAsset) Close() error {
	return os.Remove(a.File.Name())
}
