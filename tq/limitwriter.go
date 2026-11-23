package tq

import "io"

// limitWriter writes at most limit bytes to its underlying writer and drops
// whatever a source sends beyond that, while still reporting every byte as
// written. Downloads use it so that a server returning more data than the
// object's declared size cannot fill the local disk with a temporary file; the
// caller keeps seeing the whole response, so the size and OID checks still run
// over every byte received and report the mismatch as before.
type limitWriter struct {
	writer io.Writer
	limit  int64
}

func newLimitWriter(writer io.Writer, limit int64) *limitWriter {
	if limit < 0 {
		limit = 0
	}
	return &limitWriter{writer: writer, limit: limit}
}

// Write reports len(p) as written even when it discards the bytes beyond the
// limit, so the copy that drives it reads the whole source.
func (w *limitWriter) Write(p []byte) (int, error) {
	total := len(p)
	if w.limit <= 0 {
		return total, nil
	}
	if int64(len(p)) > w.limit {
		p = p[:w.limit]
	}
	written, err := w.writer.Write(p)
	w.limit -= int64(written)
	if err != nil {
		return written, err
	}
	return total, nil
}
