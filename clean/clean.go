package gitmediaclean

import (
	".."
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

type CleanedAsset struct {
	File *os.File
	Sha  string
}

func Clean(reader io.Reader) (*CleanedAsset, error) {
	tmp, err := gitmedia.TempFile()
	if err != nil {
		return nil, err
	}

	sha1Hash := sha1.New()
	writer := io.MultiWriter(sha1Hash, tmp)
	io.Copy(writer, os.Stdin)

	return &CleanedAsset{tmp, hex.EncodeToString(sha1Hash.Sum(nil))}, nil
}
