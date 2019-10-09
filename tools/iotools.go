package tools

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"os"

	"github.com/git-lfs/git-lfs/errors"
)

const (
	// memoryBufferLimit is the number of bytes to buffer in memory before
	// spooling the contents of an `io.Reader` in `Spool()` to a temporary
	// file on disk.
	memoryBufferLimit = 1024
)

// CopyWithCallback copies reader to writer while performing a progress callback
func CopyWithCallback(writer io.Writer, reader io.Reader, totalSize int64, cb CopyCallback) (int64, error) {
	if success, _ := CloneFile(writer, reader); success {
		if cb != nil {
			cb(totalSize, totalSize, 0)
		}
		return totalSize, nil
	}
	if cb == nil {
		return io.Copy(writer, reader)
	}

	cbReader := &CallbackReader{
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
// contents of 'from' into a temprorary file created in the directory "dir".
// That buffer is held in memory until the file grows to larger than
// 'memoryBufferLimit`, then the remaining contents are spooled to disk.
//
// The temporary file is cleaned up after the copy is complete.
//
// The number of bytes written to "to", as well as any error encountered are
// returned.
func Spool(to io.Writer, from io.Reader, dir string) (n int64, err error) {
	// First, buffer up to `memoryBufferLimit` in memory.
	buf := make([]byte, memoryBufferLimit)
	if bn, err := from.Read(buf); err != nil && err != io.EOF {
		return int64(bn), err
	} else {
		buf = buf[:bn]
	}

	var spool io.Reader = bytes.NewReader(buf)
	if err != io.EOF {
		// If we weren't at the end of the stream, create a temporary
		// file, and spool the remaining contents there.
		tmp, err := ioutil.TempFile(dir, "")
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

		// The spooled contents will now be the concatenation of the
		// contents we stored in memory, then the remainder of the
		// contents on disk.
		spool = io.MultiReader(spool, tmp)
	}

	return io.Copy(to, spool)
}

// Split the input on the NUL character. Usable with bufio.Scanner.
func SplitOnNul(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i := 0; i < len(data); i++ {
		if data[i] == '\x00' {
			return i + 1, data[:i], nil
		}
	}
	return 0, nil, nil
}
