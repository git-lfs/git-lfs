package tools

import "io"

// closeFn is the type of func Close() in the io.Closer interface.
type closeFn func() error

// syncFn is the type of func Sync() in the *os.File implementation.
type syncFn func() error

// SyncWriter provides a wrapper around an io.Writer that synchronizes all
// write after they occur, if the underlying writer supports synchronization.
type SyncWriter struct {
	w io.Writer

	closeFn closeFn
	syncFn  syncFn
}

// NewSyncWriter returns a new instance of the *SyncWriter that sends all writes
// to the given io.Writer.
func NewSyncWriter(w io.Writer) *SyncWriter {
	sw := &SyncWriter{
		w: w,
	}

	if sync, ok := w.(interface {
		Sync() error
	}); ok {
		sw.syncFn = sync.Sync
	} else {
		sw.syncFn = func() error { return nil }
	}

	if close, ok := w.(io.Closer); ok {
		sw.closeFn = close.Close
	} else {
		sw.closeFn = func() error { return nil }
	}

	return sw
}

// Write will write to the file and perform a Sync() if writing succeeds.
func (w *SyncWriter) Write(b []byte) error {
	if _, err := w.w.Write(b); err != nil {
		return err
	}
	return w.syncFn()
}

// Close will call Close() on the underlying file
func (w *SyncWriter) Close() error {
	return w.closeFn()
}
