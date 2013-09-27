package gitmediaclean

import (
	".."
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

type CleanedAsset struct {
	File *os.File
	*gitmedia.LargeAsset
}

func Clean(reader io.Reader) (*CleanedAsset, error) {
	tmp, err := gitmedia.TempFile()
	if err != nil {
		return nil, err
	}

	sha1Hash := sha1.New()
	md5Hash := md5.New()
	writer := io.MultiWriter(sha1Hash, md5Hash, tmp)

	written, _ := io.Copy(writer, os.Stdin)
	sha := hex.EncodeToString(sha1Hash.Sum(nil))

	asset := &gitmedia.LargeAsset{
		MediaType: "application/vnd.github.large-asset",
		Size:      written,
		MD5:       hex.EncodeToString(md5Hash.Sum(nil)),
		SHA1:      sha,
	}

	return &CleanedAsset{tmp, asset}, nil
}
