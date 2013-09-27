package main

import (
	".."
	"../clean"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func main() {
	tmp, err := gitmedia.TempFile()
	if err != nil {
		fmt.Println("Error trying to create temp file")
		panic(err)
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

	output := ChooseWriter(asset, tmp)
	defer output.Close()

	enc := gitmedia.NewEncoder(output)
	enc.Encode(asset)
}

func ChooseWriter(asset *gitmedia.LargeAsset, tmp *os.File) io.WriteCloser {
	mediafile := gitmedia.LocalMediaPath(asset.SHA1)
	if stat, _ := os.Stat(mediafile); stat == nil {
		return gitmediaclean.NewWriter(asset, tmp)
	} else {
		return gitmediaclean.NewExistingWriter(asset, tmp)
	}
}
