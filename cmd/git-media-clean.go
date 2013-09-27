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

	asset := &gitmedia.LargeAsset{
		MediaType: "application/vnd.github.large-asset",
		Size:      written,
		MD5:       hex.EncodeToString(md5Hash.Sum(nil)),
		SHA1:      sha,
	}

	output := ChooseWriter(asset, tmp)
	defer output.Close()

	output.Write([]byte(fmt.Sprintf("# %d\n", len(gitmedia.MediaWarning))))
	output.Write(gitmedia.MediaWarning)
	enc := json.NewEncoder(output)
	enc.Encode(asset)
}

func ChooseWriter(asset *gitmedia.LargeAsset, tmp *os.File) io.WriteCloser {
	mediafile := gitmedia.LocalMediaPath(asset.SHA1)
	if stat, _ := os.Stat(mediafile); stat == nil {
		return NewGitMediaCleanWriter(asset, tmp)
	} else {
		return NewGitMediaExistingWriter(asset, tmp)
	}
}

type GitMediaExistingWriter struct {
	tempfile *os.File
	writer   io.Writer
}

func NewGitMediaExistingWriter(asset *gitmedia.LargeAsset, tmp *os.File) io.WriteCloser {
	return &GitMediaExistingWriter{tmp, os.Stdout}
}

func (w *GitMediaExistingWriter) Write(buf []byte) (int, error) {
	return w.writer.Write(buf)
}

func (w *GitMediaExistingWriter) Close() error {
	return os.Remove(w.tempfile.Name())
}

type GitMediaCleanWriter struct {
	metafile *os.File
	*GitMediaExistingWriter
}

func NewGitMediaCleanWriter(asset *gitmedia.LargeAsset, tmp *os.File) io.WriteCloser {
	mediafile := gitmedia.LocalMediaPath(asset.SHA1)
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

	return &GitMediaCleanWriter{
		metafile:               file,
		GitMediaExistingWriter: &GitMediaExistingWriter{tmp, io.MultiWriter(os.Stdout, file)},
	}
}

func (w *GitMediaCleanWriter) Write(buf []byte) (int, error) {
	return w.writer.Write(buf)
}

func (w *GitMediaCleanWriter) Close() error {
	w.GitMediaExistingWriter.Close()
	return w.metafile.Close()
}
