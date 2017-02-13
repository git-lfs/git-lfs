package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/progress"
)

// CopyWithCallback copies reader to writer while performing a progress callback
func CopyWithCallback(writer io.Writer, reader io.Reader, totalSize int64, cb progress.CopyCallback) (int64, error) {
	if success, _ := CloneFile(writer, reader); success {
		if cb != nil {
			cb(totalSize, totalSize, 0)
		}
		return totalSize, nil
	}
	if cb == nil {
		return io.Copy(writer, reader)
	}

	cbReader := &progress.CallbackReader{
		C:         cb,
		TotalSize: totalSize,
		Reader:    reader,
	}
	return io.Copy(writer, cbReader)
}

// Get a new Hash instance of the type used to hash LFS content
func NewLfsContentHash() hash.Hash {
	return sha256.New()
}

// HashingReader wraps a reader and calculates the hash of the data as it is read
type HashingReader struct {
	reader io.Reader
	hasher hash.Hash
}

func NewHashingReader(r io.Reader) *HashingReader {
	return &HashingReader{r, NewLfsContentHash()}
}

func NewHashingReaderPreloadHash(r io.Reader, hash hash.Hash) *HashingReader {
	return &HashingReader{r, hash}
}

func (r *HashingReader) Hash() string {
	return hex.EncodeToString(r.hasher.Sum(nil))
}

func (r *HashingReader) Read(b []byte) (int, error) {
	w, err := r.reader.Read(b)
	if err == nil || err == io.EOF {
		_, e := r.hasher.Write(b[0:w])
		if e != nil && err == nil {
			return w, e
		}
	}

	return w, err
}

// RetriableReader wraps a error response of reader as RetriableError()
type RetriableReader struct {
	reader io.Reader
}

func NewRetriableReader(r io.Reader) io.Reader {
	return &RetriableReader{r}
}

func (r *RetriableReader) Read(b []byte) (int, error) {
	n, err := r.reader.Read(b)

	// EOF is a successful response as it is used to signal a graceful end
	// of input c.f. https://git.io/v6riQ
	//
	// Otherwise, if the error is non-nil and already retriable (in the
	// case that the underlying reader `r.reader` is itself a
	// `*RetriableReader`, return the error wholesale:
	if err == nil || err == io.EOF || errors.IsRetriableError(err) {
		return n, err
	}

	return n, errors.NewRetriableError(err)
}

// Spool spools the contents from 'from' to 'to' by buffering the entire
// contents of 'from' into a temprorary file, first.
//
// A name of "lfs-<time>" is used, where "<time>" is the number of seconds since
// the Unix epoch.
//
// The temporary file is cleaned up after the copy is complete.
//
// The number of bytes written to "to", as well as any error encountered are
// returned.
func Spool(to io.Writer, from io.Reader) (n int64, err error) {
	return SpoolName(to, from, fmt.Sprintf("lfs-%d", time.Now().Unix()))
}

// Spool spools the contents from 'from' to 'to' by buffering the entire
// contents of 'from' into a temprorary file named "name", first.
//
// The temporary file is cleaned up after the copy is complete.
//
// The number of bytes written to "to", as well as any error encountered are
// returned.
func SpoolName(to io.Writer, from io.Reader, name string) (n int64, err error) {
	tmp, err := ioutil.TempFile("", name)
	if err != nil {
		return 0, errors.Wrap(err, "spool tmp")
	}
	defer os.Remove(tmp.Name())

	if n, err = io.Copy(tmp, from); err != nil {
		return n, errors.Wrap(err, "unable to spool")
	}

	if _, err = tmp.Seek(0, io.SeekStart); err != nil {
		return 0, errors.Wrap(err, "unable to seek")
	}

	return io.Copy(to, tmp)
}
