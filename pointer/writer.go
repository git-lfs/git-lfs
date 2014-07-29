package pointer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
)

// consistentFileWriter ensures that the file matching the given SHA-256
// signature is written to the given path.
type consistentFileWriter struct {
	Path    string
	Sha256  string
	writer  io.Writer
	tmpFile *os.File
	hasher  hash.Hash
}

func newFile(path, sha256Sig string) (*consistentFileWriter, error) {
	_, err := os.Stat(path)
	if err == nil {
		return nil, fmt.Errorf("File exists: %s", path)
	}

	tmpFile, err := ioutil.TempFile("", sha256Sig)
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	w := io.MultiWriter(tmpFile, h)
	return &consistentFileWriter{path, sha256Sig, w, tmpFile, h}, nil
}

func (w *consistentFileWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func (w *consistentFileWriter) Close() error {
	defer func() {
		w.tmpFile.Close()
		os.RemoveAll(w.tmpFile.Name())
	}()

	writtenSha := hex.EncodeToString(w.hasher.Sum(nil))
	if writtenSha != w.Sha256 {
		return fmt.Errorf("Unexpected SHA-256 trying to write %s.  Expected %s, got %s.", w.Path, w.Sha256, writtenSha)
	}

	offset, err := w.tmpFile.Seek(0, 0)
	if err != nil {
		return err
	}

	if offset != 0 {
		return fmt.Errorf("Expected offset 0, go %d", offset)
	}

	file, err := os.OpenFile(w.Path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0660)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, w.tmpFile)
	if err != nil {
		os.RemoveAll(w.Path)
	}
	return err
}
