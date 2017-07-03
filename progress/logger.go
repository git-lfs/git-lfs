package progress

import "os"

// progressLogger provides a wrapper around an os.File that can either
// write to the file or ignore all writes completely.
type progressLogger struct {
	writeData bool
	log       *os.File
}

// Write will write to the file and perform a Sync() if writing succeeds.
func (l *progressLogger) Write(b []byte) error {
	if l.writeData {
		return nil
	}
	if _, err := l.log.Write(b); err != nil {
		return err
	}
	return l.log.Sync()
}

// Close will call Close() on the underlying file
func (l *progressLogger) Close() error {
	if l.log != nil {
		return l.log.Close()
	}
	return nil
}

// Shutdown will cause the logger to ignore any further writes. It should
// be used when writing causes an error.
func (l *progressLogger) Shutdown() {
	l.writeData = false
}
