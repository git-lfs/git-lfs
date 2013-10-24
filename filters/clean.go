package gitmediafilters

import (
	".."
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strconv"
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
	oidHash.Write([]byte(strconv.FormatInt(written, 10)))

	return &CleanedAsset{written, tmp, hex.EncodeToString(oidHash.Sum(nil)), ""}, err
}

func (a *CleanedAsset) Close() error {
	return os.Remove(a.File.Name())
}
