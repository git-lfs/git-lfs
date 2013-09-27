package main

import (
	".."
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
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

	meta := &gitmedia.LargeAsset{
		MediaType: "application/vnd.github.large-asset",
		Size:      written,
		MD5:       hex.EncodeToString(md5Hash.Sum(nil)),
		SHA1:      sha,
	}

	mediafile := gitmedia.LocalMediaPath(sha)
	metafile := mediafile + ".json"

	if err := os.Rename(tmp.Name(), mediafile); err != nil {
		fmt.Printf("Unable to move %s to %s\n", tmp.Name(), mediafile)
		panic(err)
	}

	file, err := os.Create(metafile)
	if err != nil {
		fmt.Printf("Unable to create meta data file: %s\n", metafile)
		panic(err)
	}

	defer file.Close()
	defer func() { os.Remove(tmp.Name()) }()

	writer = io.MultiWriter(os.Stdout, file)
	writer.Write([]byte(fmt.Sprintf("# %d\n", len(gitmedia.MediaWarning))))
	writer.Write(gitmedia.MediaWarning)
	enc := json.NewEncoder(writer)
	enc.Encode(meta)
}
