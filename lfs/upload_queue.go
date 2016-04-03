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
	size     int64
	object   *ObjectResource
}

// NewUploadable builds the Uploadable from the given information.
// "filename" can be empty if a raw object is pushed (see "object-id" flag in push command)/
func NewUploadable(oid, filename string) (*Uploadable, error) {
	localMediaPath, err := LocalMediaPath(oid)
	if err != nil {
		return nil, Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	if len(filename) > 0 {
		if err := ensureFile(filename, localMediaPath); err != nil {
			return nil, err
		}
	}

	fi, err := os.Stat(localMediaPath)
	if err != nil {
		return nil, Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	return &Uploadable{oid: oid, OidPath: localMediaPath, Filename: filename, size: fi.Size()}, nil
}

func (u *Uploadable) Check() (*ObjectResource, error) {
	return UploadCheck(u.OidPath)
}

func (u *Uploadable) Transfer(cb CopyCallback) error {
	wcb := func(total, read int64, current int) error {
		cb(total, read, current)
		return nil
	}

	return UploadObject(u.object, wcb)
}

func (u *Uploadable) Object() *ObjectResource {
	return u.object
}

func (u *Uploadable) Oid() string {
	return u.oid
}

func (u *Uploadable) Size() int64 {
	return u.size
}

func (u *Uploadable) Name() string {
	return u.Filename
}

func (u *Uploadable) SetObject(o *ObjectResource) {
	u.object = o
}

// NewUploadQueue builds an UploadQueue, allowing `workers` concurrent uploads.
func NewUploadQueue(files int, size int64, dryRun bool, metadata *TransferMetadata) *TransferQueue {
	q := newTransferQueue(files, size, dryRun)
	q.transferKind = "upload"
	q.metadata = metadata
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

	cleaned, err := PointerClean(file, file.Name(), stat.Size(), nil)
	if cleaned != nil {
		cleaned.Teardown()
	}

	if err != nil {
		return err
	}

	if expectedOid != cleaned.Oid {
		return fmt.Errorf("Trying to push %q with OID %s.\nNot found in %s.", smudgePath, expectedOid, filepath.Dir(cleanPath))
	}

	return nil
}
