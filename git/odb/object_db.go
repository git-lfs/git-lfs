package odb

import (
	"bytes"
	"io"
	"io/ioutil"
)

// ObjectDatabase enables the reading and writing of objects against a storage
// backend.
type ObjectDatabase struct {
	// s is the storage backend which opens/creates/reads/writes.
	s storer
}

// FromFilesystem constructs an *ObjectDatabase instance that is backed by a
// directory on the filesystem. Specifically, this should point to:
//
//  /absolute/repo/path/.git/objects
func FromFilesystem(root string) (*ObjectDatabase, error) {
	return &ObjectDatabase{s: newFileStorer(root)}, nil
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
		return nil, err
	}

	return NewObjectReadCloser(f)
}

// decode decodes an object given by the sha "sha []byte" into the given object
// "into", or returns an error if one was encountered.
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

	if err = r.Close(); err != nil {
		return err
	}
	return nil
}
