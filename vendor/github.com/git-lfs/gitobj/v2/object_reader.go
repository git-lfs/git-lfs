package gitobj

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

// ObjectReader provides an io.Reader implementation that can read Git object
// headers, as well as provide an uncompressed view into the object contents
// itself.
type ObjectReader struct {
	// header is the object header type
	header *struct {
		// typ is the ObjectType encoded in the header pointed at by
		// this reader.
		typ ObjectType
		// size is the number of uncompressed bytes following the header
		// that encodes the object.
		size int64
	}
	// r is the underling uncompressed reader.
	r *bufio.Reader

	// closeFn supplies an optional function that, when called, frees an
	// resources (open files, memory, etc) held by this instance of the
	// *ObjectReader.
	//
	// closeFn returns any error encountered when closing/freeing resources
	// held.
	//
	// It is allowed to be nil.
	closeFn func() error
}

// NewObjectReader takes a given io.Reader that yields zlib-compressed data, and
// returns an *ObjectReader wrapping it, or an error if one occurred during
// construction time.
func NewObjectReader(r io.Reader) (*ObjectReader, error) {
	return NewObjectReadCloser(ioutil.NopCloser(r))
}

// NewObjectReader takes a given io.Reader that yields uncompressed data and
// returns an *ObjectReader wrapping it, or an error if one occurred during
// construction time.
func NewUncompressedObjectReader(r io.Reader) (*ObjectReader, error) {
	return NewUncompressedObjectReadCloser(ioutil.NopCloser(r))
}

// NewObjectReadCloser takes a given io.Reader that yields zlib-compressed data, and
// returns an *ObjectReader wrapping it, or an error if one occurred during
// construction time.
//
// It also calls the Close() function given by the implementation "r" of the
// type io.Closer.
func NewObjectReadCloser(r io.ReadCloser) (*ObjectReader, error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &ObjectReader{
		r: bufio.NewReader(zr),
		closeFn: func() error {
			if err := zr.Close(); err != nil {
				return err
			}
			if err := r.Close(); err != nil {
				return err
			}
			return nil
		},
	}, nil
}

// NewUncompressObjectReadCloser takes a given io.Reader that yields
// uncompressed data, and returns an *ObjectReader wrapping it, or an error if
// one occurred during construction time.
//
// It also calls the Close() function given by the implementation "r" of the
// type io.Closer.
func NewUncompressedObjectReadCloser(r io.ReadCloser) (*ObjectReader, error) {
	return &ObjectReader{
		r:       bufio.NewReader(r),
		closeFn: r.Close,
	}, nil
}

// Header returns information about the Object's header, or an error if one
// occurred while reading the data.
//
// Header information is cached, so this function is safe to call at any point
// during the object read, and can be called more than once.
func (r *ObjectReader) Header() (typ ObjectType, size int64, err error) {
	if r.header != nil {
		return r.header.typ, r.header.size, nil
	}

	typs, err := r.r.ReadString(' ')
	if err != nil {
		return UnknownObjectType, 0, err
	}
	if len(typs) == 0 {
		return UnknownObjectType, 0, fmt.Errorf(
			"gitobj: object type must not be empty",
		)
	}
	typs = strings.TrimSuffix(typs, " ")

	sizeStr, err := r.r.ReadString('\x00')
	if err != nil {
		return UnknownObjectType, 0, err
	}
	sizeStr = strings.TrimSuffix(sizeStr, "\x00")

	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return UnknownObjectType, 0, err
	}

	r.header = &struct {
		typ  ObjectType
		size int64
	}{
		ObjectTypeFromString(typs),
		size,
	}

	return r.header.typ, r.header.size, nil
}

// Read reads uncompressed bytes into the buffer "p", and returns the number of
// uncompressed bytes read. Otherwise, it returns any error encountered along
// the way.
//
// This function is safe to call before reading the Header information, as any
// call to Read() will ensure that read has been called at least once.
func (r *ObjectReader) Read(p []byte) (n int, err error) {
	if _, _, err = r.Header(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

// Close frees any resources held by the ObjectReader and must be called before
// disposing of this instance.
//
// It returns any error encountered by the *ObjectReader during close.
func (r *ObjectReader) Close() error {
	if r.closeFn == nil {
		return nil
	}
	return r.closeFn()
}
