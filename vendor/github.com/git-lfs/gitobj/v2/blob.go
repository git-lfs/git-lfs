package gitobj

import (
	"bytes"
	"fmt"
	"hash"
	"io"
	"os"
)

// Blob represents a Git object of type "blob".
type Blob struct {
	// Size is the total uncompressed size of the blob's contents.
	Size int64
	// Contents is a reader that yields the uncompressed blob contents. It
	// may only be read once. It may or may not implement io.ReadSeeker.
	Contents io.Reader

	// closeFn is a function that is called to free any resources held by
	// the Blob.  In particular, this will close a file, if the Blob is
	// being read from a file on disk.
	closeFn func() error
}

// NewBlobFromBytes returns a new *Blob that yields the data given.
func NewBlobFromBytes(contents []byte) *Blob {
	return &Blob{
		Contents: bytes.NewReader(contents),
		Size:     int64(len(contents)),
	}
}

// NewBlobFromFile returns a new *Blob that contains the contents of the file
// at location "path" on disk. NewBlobFromFile does not read the file ahead of
// time, and instead defers this task until encoding the blob to the object
// database.
//
// If the file cannot be opened or stat(1)-ed, an error will be returned.
//
// When the blob receives a function call Close(), the file will also be closed,
// and any error encountered in doing so will be returned from Close().
func NewBlobFromFile(path string) (*Blob, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("gitobj: could not open: %s: %s", path,
			err)
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("gitobj: could not stat %s: %s", path,
			err)
	}

	return &Blob{
		Contents: f,
		Size:     stat.Size(),

		closeFn: func() error {
			if err := f.Close(); err != nil {
				return fmt.Errorf(
					"gitobj: could not close %s: %s",
					path, err)
			}
			return nil
		},
	}, nil
}

// Type implements Object.ObjectType by returning the correct object type for
// Blobs, BlobObjectType.
func (b *Blob) Type() ObjectType { return BlobObjectType }

// Decode implements Object.Decode and decodes the uncompressed blob contents
// being read. It returns the number of bytes that it consumed off of the
// stream, which is always zero.
//
// If any errors are encountered while reading the blob, they will be returned.
func (b *Blob) Decode(hash hash.Hash, r io.Reader, size int64) (n int, err error) {
	b.Size = size
	b.Contents = io.LimitReader(r, size)

	b.closeFn = func() error {
		if closer, ok := r.(io.Closer); ok {
			return closer.Close()
		}
		return nil
	}

	return 0, nil
}

// Encode encodes the blob's contents to the given io.Writer, "w". If there was
// any error copying the blob's contents, that error will be returned.
//
// Otherwise, the number of bytes written will be returned.
func (b *Blob) Encode(to io.Writer) (n int, err error) {
	nn, err := io.Copy(to, b.Contents)

	return int(nn), err
}

// Closes closes any resources held by the open Blob, or returns nil if there
// were no errors.
func (b *Blob) Close() error {
	if b.closeFn == nil {
		return nil
	}
	return b.closeFn()
}

// Equal returns whether the receiving and given blobs are equal, or in other
// words, whether they are represented by the same SHA-1 when saved to the
// object database.
func (b *Blob) Equal(other *Blob) bool {
	if (b == nil) != (other == nil) {
		return false
	}

	if b != nil {
		return b.Contents == other.Contents &&
			b.Size == other.Size
	}
	return true
}
