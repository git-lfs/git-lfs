package odb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
)

// ObjectDatabase enables the reading and writing of objects against a storage
// backend.
type ObjectDatabase struct {
	// s is the storage backend which opens/creates/reads/writes.
	s storer

	// closed is a uint32 managed by sync/atomic's <X>Uint32 methods. It
	// yields a value of 0 if the *ObjectDatabase it is stored upon is open,
	// and a value of 1 if it is closed.
	closed uint32
	// objectScanner is the running instance of `*git.ObjectScanner` used to
	// scan packed objects not found in .git/objects/xx/... directly.
	objectScanner *git.ObjectScanner
}

// FromFilesystem constructs an *ObjectDatabase instance that is backed by a
// directory on the filesystem. Specifically, this should point to:
//
//  /absolute/repo/path/.git/objects
func FromFilesystem(root string) (*ObjectDatabase, error) {
	os, err := git.NewObjectScanner()
	if err != nil {
		return nil, err
	}

	return &ObjectDatabase{
		s:             newFileStorer(root),
		objectScanner: os,
	}, nil
}

// Close closes the *ObjectDatabase, freeing any open resources (namely: the
// `*git.ObjectScanner instance), and returning any errors encountered in
// closing them.
//
// If Close() has already been called, this function will return an error.
func (o *ObjectDatabase) Close() error {
	if !atomic.CompareAndSwapUint32(&o.closed, 0, 1) {
		return errors.New("git/odb: *ObjectDatabase already closed")
	}

	if err := o.objectScanner.Close(); err != nil {
		return err
	}
	return nil
}

// Blob returns a *Blob as identified by the SHA given, or an error if one was
// encountered.
func (o *ObjectDatabase) Blob(sha []byte) (*Blob, error) {
	var b Blob

	if err := o.decode(sha, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// Tree returns a *Tree as identified by the SHA given, or an error if one was
// encountered.
func (o *ObjectDatabase) Tree(sha []byte) (*Tree, error) {
	var t Tree
	if err := o.decode(sha, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Commit returns a *Commit as identified by the SHA given, or an error if one
// was encountered.
func (o *ObjectDatabase) Commit(sha []byte) (*Commit, error) {
	var c Commit

	if err := o.decode(sha, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// WriteBlob stores a *Blob on disk and returns the SHA it is uniquely
// identified by, or an error if one was encountered.
func (o *ObjectDatabase) WriteBlob(b *Blob) ([]byte, error) {
	buf, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	defer buf.Close()

	sha, _, err := o.encodeBuffer(b, buf)
	if err != nil {
		return nil, err
	}

	if err = b.Close(); err != nil {
		return nil, err
	}

	return sha, nil
}

// WriteTree stores a *Tree on disk and returns the SHA it is uniquely
// identified by, or an error if one was encountered.
func (o *ObjectDatabase) WriteTree(t *Tree) ([]byte, error) {
	sha, _, err := o.encode(t)
	if err != nil {
		return nil, err
	}
	return sha, nil
}

// WriteCommit stores a *Commit on disk and returns the SHA it is uniquely
// identified by, or an error if one was encountered.
func (o *ObjectDatabase) WriteCommit(c *Commit) ([]byte, error) {
	sha, _, err := o.encode(c)
	if err != nil {
		return nil, err
	}
	return sha, nil
}

// Root returns the filesystem root that this *ObjectDatabase works within, if
// backed by a fileStorer (constructed by FromFilesystem). If so, it returns
// the fully-qualified path on a disk and a value of true.
//
// Otherwise, it returns empty-string and a value of false.
func (o *ObjectDatabase) Root() (string, bool) {
	type rooter interface {
		Root() string
	}

	if root, ok := o.s.(rooter); ok {
		return root.Root(), true
	}
	return "", false
}

// encode encodes and saves an object to the storage backend and uses an
// in-memory buffer to calculate the object's encoded body.
func (d *ObjectDatabase) encode(object Object) (sha []byte, n int64, err error) {
	return d.encodeBuffer(object, bytes.NewBuffer(nil))
}

// encodeBuffer encodes and saves an object to the storage backend by using the
// given buffer to calculate and store the object's encoded body.
func (d *ObjectDatabase) encodeBuffer(object Object, buf io.ReadWriter) (sha []byte, n int64, err error) {
	cn, err := object.Encode(buf)
	if err != nil {
		return nil, 0, err
	}

	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, 0, err
	}
	defer tmp.Close()

	to := NewObjectWriter(tmp)
	if _, err = to.WriteHeader(object.Type(), int64(cn)); err != nil {
		return nil, 0, err
	}

	if seek, ok := buf.(io.Seeker); ok {
		if _, err = seek.Seek(0, io.SeekStart); err != nil {
			return nil, 0, err
		}
	}

	if _, err = io.Copy(to, buf); err != nil {
		return nil, 0, err
	}

	if err = to.Close(); err != nil {
		return nil, 0, err
	}

	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return nil, 0, err
	}
	return d.save(to.Sha(), tmp)
}

// save writes the given buffer to the location given by the storer "o.s" as
// identified by the sha []byte.
func (o *ObjectDatabase) save(sha []byte, buf io.Reader) ([]byte, int64, error) {
	n, err := o.s.Store(sha, buf)

	return sha, n, err
}

// open gives an `*ObjectReader` for the given loose object keyed by the given
// "sha" []byte, or an error.
func (o *ObjectDatabase) open(sha []byte) (*ObjectReader, error) {
	f, err := o.s.Open(sha)
	if err != nil {
		if !os.IsNotExist(err) {
			// If there was some other issue beyond not being able
			// to find the object, return that immediately and don't
			// try and fallback to the *git.ObjectScanner.
			return nil, err
		}

		// Otherwise, if the file simply couldn't be found, attempt to
		// load its contents from the *git.ObjectScanner by leveraging
		// `git-cat-file --batch`.
		if atomic.LoadUint32(&o.closed) == 1 {
			return nil, errors.New("git/odb: cannot use closed *git.ObjectScanner")
		}

		if !o.objectScanner.Scan(hex.EncodeToString(sha)) {
			return nil, o.objectScanner.Err()
		}

		return NewUncompressedObjectReader(io.MultiReader(
			// Git object header:
			strings.NewReader(fmt.Sprintf("%s %d\x00",
				o.objectScanner.Type(), o.objectScanner.Size(),
			)),

			// Git object (uncompressed) contents:
			o.objectScanner.Contents(),
		))
	}

	return NewObjectReadCloser(f)
}

// decode decodes an object given by the sha "sha []byte" into the given object
// "into", or returns an error if one was encountered.
//
// Ordinarily, it closes the object's underlying io.ReadCloser (if it implements
// the `io.Closer` interface), but skips this if the "into" Object is of type
// BlobObjectType. Blob's don't exhaust the buffer completely (they instead
// maintain a handle on the blob's contents via an io.LimitedReader) and
// therefore cannot be closed until signaled explicitly by git/odb.Blob.Close().
func (o *ObjectDatabase) decode(sha []byte, into Object) error {
	r, err := o.open(sha)
	if err != nil {
		return err
	}

	typ, size, err := r.Header()
	if err != nil {
		return err
	} else if typ != into.Type() {
		return &UnexpectedObjectType{Got: typ, Wanted: into.Type()}
	}

	if _, err = into.Decode(r, size); err != nil {
		return err
	}

	if into.Type() == BlobObjectType {
		return nil
	}
	return r.Close()
}
