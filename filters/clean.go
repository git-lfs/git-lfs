package gitmediafilters

import (
	".."
	"crypto/sha1"
	"encoding/hex"
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

	sha1Hash := sha1.New()
	writer := io.MultiWriter(sha1Hash, tmp)
	written, err := io.Copy(writer, reader)

	return &CleanedAsset{written, tmp, hex.EncodeToString(sha1Hash.Sum(nil)), ""}, err
}

func (a *CleanedAsset) Close() error {
	return os.Remove(a.File.Name())
}
