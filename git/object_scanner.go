package git

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/git-lfs/gitobj/v2"
	"github.com/git-lfs/gitobj/v2/errors"
)

// object represents a generic Git object of any type.
type object struct {
	// Contents reads Git's internal object representation.
	Contents io.Reader
	// Oid is the ID of the object.
	Oid string
	// Size is the size in bytes of the object.
	Size int64
	// Type is the type of the object being held.
	Type string
	// object is the gitobj object being handled.
	object gitobj.Object
}

// ObjectScanner is a scanner type that scans for Git objects reference-able in
// Git's object database by their unique OID.
type ObjectScanner struct {
	// object is the object that the ObjectScanner last scanned, or nil.
	object *object
	// err is the error (if any) that the ObjectScanner encountered during
	// its last scan, or nil.
	err error

	gitobj *gitobj.ObjectDatabase
}

// NewObjectScanner constructs a new instance of the `*ObjectScanner` type and
// returns it. It backs the ObjectScanner with an invocation of the `git
// cat-file --batch` command. If any errors were encountered while starting that
// command, they will be returned immediately.
//
// Otherwise, an `*ObjectScanner` is returned with no error.
func NewObjectScanner(gitEnv, osEnv Environment) (*ObjectScanner, error) {
	gitdir, err := GitCommonDir()
	if err != nil {
		return nil, err
	}

	gitobj, err := ObjectDatabase(osEnv, gitEnv, gitdir, "")
	if err != nil {
		return nil, err
	}

	return NewObjectScannerFrom(gitobj), nil
}

// NewObjectScannerFrom returns a new `*ObjectScanner` populated with data from
// the given `io.Reader`, "r". It supplies no close function, and discards any
// input given to the Scan() function.
func NewObjectScannerFrom(db *gitobj.ObjectDatabase) *ObjectScanner {
	return &ObjectScanner{gitobj: db}
}

// Scan scans for a particular object given by the "oid" parameter. Once the
// scan is complete, the Contents(), Sha1(), Size() and Type() functions may be
// called and will return data corresponding to the given OID.
//
// Scan() returns whether the scan was successful, or in other words, whether or
// not the scanner can continue to progress.
func (s *ObjectScanner) Scan(oid string) bool {
	if err := s.reset(); err != nil {
		s.err = err
		return false
	}

	obj, err := s.scan(oid)
	s.object = obj

	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return false
	}
	return true
}

// Close closes and frees any resources owned by the *ObjectScanner that it is
// called upon. If there were any errors in freeing that (those) resource(s), it
// it will be returned, otherwise nil.
func (s *ObjectScanner) Close() error {
	if s == nil {
		return nil
	}

	s.reset()
	s.gitobj.Close()

	return nil
}

// Contents returns an io.Reader which reads Git's representation of the object
// that was last scanned for.
func (s *ObjectScanner) Contents() io.Reader {
	return s.object.Contents
}

// Sha1 returns the SHA1 object ID of the object that was last scanned for.
func (s *ObjectScanner) Sha1() string {
	return s.object.Oid
}

// Size returns the size in bytes of the object that was last scanned for.
func (s *ObjectScanner) Size() int64 {
	return s.object.Size
}

// Type returns the type of the object that was last scanned for.
func (s *ObjectScanner) Type() string {
	return s.object.Type
}

// Err returns the error (if any) that was encountered during the last Scan()
// operation.
func (s *ObjectScanner) Err() error { return s.err }

func (s *ObjectScanner) reset() error {
	if s.object != nil {
		if c, ok := s.object.object.(interface {
			Close() error
		}); ok && c != nil {
			if err := c.Close(); err != nil {
				return err
			}
		}
	}

	s.object, s.err = nil, nil
	return nil
}

type missingErr struct {
	oid string
}

func (m *missingErr) Error() string {
	return fmt.Sprintf("missing object: %s", m.oid)
}

func IsMissingObject(err error) bool {
	_, ok := err.(*missingErr)
	return ok
}

func mustDecode(oid string) []byte {
	x, _ := hex.DecodeString(oid)
	return x
}

func (s *ObjectScanner) scan(oid string) (*object, error) {
	var (
		obj      gitobj.Object
		size     int64
		contents io.Reader
	)

	obj, err := s.gitobj.Object(mustDecode(oid))
	if err != nil {
		if errors.IsNoSuchObject(err) {
			return nil, &missingErr{oid: oid}
		}
		return nil, err
	}

	// Currently, we're only interested in the size and contents of blobs,
	// and gitobj only exposes the size easily for us for blobs anyway.
	if obj.Type() == gitobj.BlobObjectType {
		blob := obj.(*gitobj.Blob)
		size = blob.Size
		contents = blob.Contents
	}

	return &object{
		Contents: contents,
		Oid:      oid,
		Size:     size,
		Type:     obj.Type().String(),
		object:   obj,
	}, nil
}
