package gitobj

import (
	"compress/zlib"
	"fmt"
	"hash"
	"io"
	"sync/atomic"
)

// ObjectWriter provides an implementation of io.Writer that compresses and
// writes data given to it, and keeps track of the SHA1 hash of the data as it
// is written.
type ObjectWriter struct {
	// members managed via sync/atomic must be aligned at the top of this
	// structure (see: https://github.com/git-lfs/git-lfs/pull/2880).

	// wroteHeader is a uint32 managed by the sync/atomic package. It is 1
	// if the header was written, and 0 otherwise.
	wroteHeader uint32

	// w is the underling writer that this ObjectWriter is writing to.
	w io.Writer
	// sum is the in-progress hash calculation.
	sum hash.Hash

	// closeFn supplies an optional function that, when called, frees an
	// resources (open files, memory, etc) held by this instance of the
	// *ObjectWriter.
	//
	// closeFn returns any error encountered when closing/freeing resources
	// held.
	//
	// It is allowed to be nil.
	closeFn func() error
}

// nopCloser provides a no-op implementation of the io.WriteCloser interface by
// taking an io.Writer and wrapping it with a Close() method that returns nil.
type nopCloser struct {
	// Writer is an embedded io.Writer that receives the Write() method
	// call.
	io.Writer
}

// Close implements the io.Closer interface by returning nil.
func (n *nopCloser) Close() error {
	return nil
}

// NewObjectWriter returns a new *ObjectWriter instance that drains incoming
// writes into the io.Writer given, "w".  "hash" is a hash instance from the
// ObjectDatabase'e Hash method.
func NewObjectWriter(w io.Writer, hash hash.Hash) *ObjectWriter {
	return NewObjectWriteCloser(&nopCloser{w}, hash)
}

// NewObjectWriter returns a new *ObjectWriter instance that drains incoming
// writes into the io.Writer given, "w".  "sum" is a hash instance from the
// ObjectDatabase'e Hash method.
//
// Upon closing, it calls the given Close() function of the io.WriteCloser.
func NewObjectWriteCloser(w io.WriteCloser, sum hash.Hash) *ObjectWriter {
	zw := zlib.NewWriter(w)
	sum.Reset()

	return &ObjectWriter{
		w:   io.MultiWriter(zw, sum),
		sum: sum,

		closeFn: func() error {
			if err := zw.Close(); err != nil {
				return err
			}
			if err := w.Close(); err != nil {
				return err
			}
			return nil
		},
	}
}

// WriteHeader writes object header information and returns the number of
// uncompressed bytes written, or any error that was encountered along the way.
//
// WriteHeader MUST be called only once, or a panic() will occur.
func (w *ObjectWriter) WriteHeader(typ ObjectType, len int64) (n int, err error) {
	if !atomic.CompareAndSwapUint32(&w.wroteHeader, 0, 1) {
		panic("gitobj: cannot write headers more than once")
	}
	return fmt.Fprintf(w, "%s %d\x00", typ, len)
}

// Write writes the given buffer "p" of uncompressed bytes into the underlying
// data-stream, returning the number of uncompressed bytes written, along with
// any error encountered along the way.
//
// A call to WriteHeaders MUST occur before calling Write, or a panic() will
// occur.
func (w *ObjectWriter) Write(p []byte) (n int, err error) {
	if atomic.LoadUint32(&w.wroteHeader) != 1 {
		panic("gitobj: cannot write data without header")
	}
	return w.w.Write(p)
}

// Sha returns the in-progress SHA1 of the compressed object contents.
func (w *ObjectWriter) Sha() []byte {
	return w.sum.Sum(nil)
}

// Close closes the ObjectWriter and frees any resources held by it, including
// flushing the zlib-compressed content to the underling writer. It must be
// called before discarding of the Writer instance.
//
// If any error occurred while calling close, it will be returned immediately,
// otherwise nil.
func (w *ObjectWriter) Close() error {
	if w.closeFn == nil {
		return nil
	}
	return w.closeFn()
}
