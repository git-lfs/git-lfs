package gitmediafilters

import (
	".."
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

type CleanedAsset struct {
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
	io.Copy(writer, reader)

	return &CleanedAsset{tmp, hex.EncodeToString(sha1Hash.Sum(nil)), ""}, nil
}

func (a *CleanedAsset) Writer(writer io.Writer) io.WriteCloser {
	if stat := a.Stat(); stat == nil {
		return NewWriter(a, writer)
	}

	return NewExistingWriter(a, writer)
}

func (a *CleanedAsset) Path() string {
	if len(a.mediafilepath) == 0 {
		a.mediafilepath = gitmedia.LocalMediaPath(a.Sha)
	}
	return a.mediafilepath
}

func (a *CleanedAsset) Stat() os.FileInfo {
	stat, _ := os.Stat(a.Path())
	return stat
}
