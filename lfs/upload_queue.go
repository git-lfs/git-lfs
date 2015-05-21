package lfs

import (
	"fmt"
	"os"
	"path/filepath"
)

// Uploadable describes a file that can be uploaded.
type Uploadable struct {
	oid      string
	OidPath  string
	Filename string
	CB       CopyCallback
	size     int64
	object   *objectResource
}

// NewUploadable builds the Uploadable from the given information.
func NewUploadable(oid, filename string, index, totalFiles int) (*Uploadable, *WrappedError) {
	path, err := LocalMediaPath(oid)
	if err != nil {
		return nil, Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	if err := ensureFile(filename, path); err != nil {
		return nil, Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	fi, err := os.Stat(filename)
	if err != nil {
		return nil, Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	cb, file, cbErr := CopyCallbackFile("push", filename, index, totalFiles)
	if cbErr != nil {
		fmt.Fprintln(os.Stderr, cbErr.Error())
	}

	if file != nil {
		defer file.Close()
	}

	return &Uploadable{oid: oid, OidPath: path, Filename: filename, CB: cb, size: fi.Size()}, nil
}

func (u *Uploadable) Check() (*objectResource, *WrappedError) {
	return UploadCheck(u.OidPath)
}

func (u *Uploadable) Transfer(cb CopyCallback) *WrappedError {
	wcb := func(total, read int64, current int) error {
		cb(total, read, current)
		if u.CB != nil {
			return u.CB(total, read, current)
		}
		return nil
	}

	return UploadObject(u.object, wcb)
}

func (u *Uploadable) Object() *objectResource {
	return u.object
}

func (u *Uploadable) Oid() string {
	return u.oid
}

func (u *Uploadable) Size() int64 {
	return u.size
}

func (u *Uploadable) SetObject(o *objectResource) {
	u.object = o
}

// NewUploadQueue builds an UploadQueue, allowing `workers` concurrent uploads.
func NewUploadQueue(workers, files int) *TransferQueue {
	q := newTransferQueue(workers, files)
	q.transferKind = "upload"
	return q
}

// ensureFile makes sure that the cleanPath exists before pushing it.  If it
// does not exist, it attempts to clean it by reading the file at smudgePath.
func ensureFile(smudgePath, cleanPath string) error {
	if _, err := os.Stat(cleanPath); err == nil {
		return nil
	}

	expectedOid := filepath.Base(cleanPath)
	localPath := filepath.Join(LocalWorkingDir, smudgePath)
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	cleaned, err := PointerClean(file, stat.Size(), nil)
	if err != nil {
		return err
	}

	cleaned.Close()

	if expectedOid != cleaned.Oid {
		return fmt.Errorf("Expected %s to have an OID of %s, got %s", smudgePath, expectedOid, cleaned.Oid)
	}

	return nil
}
