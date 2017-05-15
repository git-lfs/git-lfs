package odb

import (
	"bytes"
	"io"
	"io/ioutil"
)

// ObjectDatabase enables the reading and writing of objects against a storage
// backend, the Storer.
type ObjectDatabase struct {
	// s is the storage backend which opens/creates/reads/writes.
	s Storer
}

// FromFilesystem constructs an *ObjectDatabase instance that is backed by a
// directory on the filesystem. Specifically, this should point to:
//
//  /absolute/repo/path/.git/objects
func FromFilesystem(root string) (*ObjectDatabase, error) {
	return &ObjectDatabase{s: NewFileStorer(root)}, nil
}

// encode encodes and saves an object to the storage backend and uses an
// in-memory buffer to calculate the object's encoded body.
func (d *ObjectDatabase) encode(object Object) (sha []byte, n int, err error) {
	return d.encodeBuffer(object, bytes.NewBuffer(nil))
}

// encodeBuffer encodes and saves an object to the storage backend by using the
// given buffer to calculate and store the object's encoded body.
func (d *ObjectDatabase) encodeBuffer(object Object, buf io.ReadWriter) (sha []byte, n int, err error) {
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

	return d.save(to.Sha(), tmp)
}

// save writes the given buffer to the location given by the storer "o.s" as
// identified by the sha []byte.
func (o *ObjectDatabase) save(sha []byte, buf io.Reader) ([]byte, int, error) {
	f, err := o.s.Create(sha)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	n, err := io.Copy(f, buf)
	return sha, int(n), err
}

// decode decodes an object given by the sha "sha []byte" into the given object
// "into", or returns an error if one was encountered.
func (o *ObjectDatabase) decode(sha []byte, into Object) error {
	f, err := o.s.Open(sha)
	if err != nil {
		return err
	}

	r, err := NewReadCloser(f)
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
