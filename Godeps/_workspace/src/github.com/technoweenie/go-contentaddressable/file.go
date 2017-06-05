package contentaddressable

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"os"
	"path/filepath"
)

var (
	AlreadyClosed = errors.New("Already closed.")
	HasData       = errors.New("Destination file already has data.")
)

// File handles the atomic writing of a content addressable file.  It writes to
// a temp file, and then renames to the final location after Accept().
type File struct {
	Oid          string
	filename     string
	tempFilename string
	file         *os.File
	tempFile     *os.File
	hasher       hash.Hash
}

// NewFile initializes a content addressable file for writing.  It opens both
// the given filename, and a temp filename in exclusive mode.  The *File OID
// is taken from the base name of the given filename.
func NewFile(filename string) (*File, error) {
	oid := filepath.Base(filename)
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	tempFilename := filename + "-temp"
	tempFile, err := os.OpenFile(tempFilename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		cleanupFile(tempFile)
		return nil, err
	}

	caw := &File{
		Oid:          oid,
		filename:     filename,
		tempFilename: tempFilename,
		file:         file,
		tempFile:     tempFile,
		hasher:       sha256.New(),
	}

	return caw, nil
}

// Write sends data to the temporary file.
func (w *File) Write(p []byte) (int, error) {
	if w.Closed() {
		return 0, AlreadyClosed
	}

	w.hasher.Write(p)
	return w.tempFile.Write(p)
}

// Accept verifies the written content SHA-256 signature matches the given OID.
// If it matches, the temp file is renamed to the original filename.  If not,
// an error is returned.
func (w *File) Accept() error {
	if w.Closed() {
		return AlreadyClosed
	}

	sig := hex.EncodeToString(w.hasher.Sum(nil))
	if sig != w.Oid {
		return fmt.Errorf("Content mismatch.  Expected OID %s, got %s", w.Oid, sig)
	}

	if err := cleanupFile(w.file); err != nil {
		return err
	}
	w.file = nil

	w.tempFile.Close()
	err := os.Rename(w.tempFilename, w.filename)
	w.Close()
	return err
}

// Close cleans up the internal file objects.
func (w *File) Close() error {
	if w.tempFile != nil {
		if err := cleanupFile(w.tempFile); err != nil {
			return err
		}
		w.tempFile = nil
	}

	if w.file != nil {
		if err := cleanupFile(w.file); err != nil {
			return err
		}
		w.file = nil
	}

	return nil
}

// Closed reports whether this file object has been closed.
func (w *File) Closed() bool {
	if w.tempFile == nil || w.file == nil {
		return true
	}
	return false
}

func cleanupFile(f *os.File) error {
	err := f.Close()
	if err := os.RemoveAll(f.Name()); err != nil {
		return err
	}

	return err
}
