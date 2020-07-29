package pack

import (
	"compress/zlib"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
)

// Packfile encapsulates the behavior of accessing an unpacked representation of
// all of the objects encoded in a single packfile.
type Packfile struct {
	// Version is the version of the packfile.
	Version uint32
	// Objects is the total number of objects in the packfile.
	Objects uint32
	// idx is the corresponding "pack-*.idx" file giving the positions of
	// objects in this packfile.
	idx *Index

	// hash is the hash algorithm used in this pack.
	hash hash.Hash

	// r is an io.ReaderAt that allows read access to the packfile itself.
	r io.ReaderAt
}

// Close closes the packfile if the underlying data stream is closeable. If so,
// it returns any error involved in closing.
func (p *Packfile) Close() error {
	var iErr error
	if p.idx != nil {
		iErr = p.idx.Close()
	}

	if close, ok := p.r.(io.Closer); ok {
		return close.Close()
	}
	return iErr
}

// Object returns a reference to an object packed in the receiving *Packfile. It
// does not attempt to unpack the packfile, rather, that is accomplished by
// calling Unpack() on the returned *Object.
//
// If there was an error loading or buffering the base, it will be returned
// without an object.
//
// If the object given by the SHA-1 name, "name", could not be found,
// (nil, errNotFound) will be returned.
//
// If the object was able to be loaded successfully, it will be returned without
// any error.
func (p *Packfile) Object(name []byte) (*Object, error) {
	// First, try and determine the offset of the last entry in the
	// delta-base chain by loading it from the corresponding pack index.
	entry, err := p.idx.Entry(name)
	if err != nil {
		if !IsNotFound(err) {
			// If the error was not an errNotFound, re-wrap it with
			// additional context.
			err = fmt.Errorf("gitobj/pack: could not load index: %s", err)
		}
		return nil, err
	}

	// If all goes well, then unpack the object at that given offset.
	r, err := p.find(int64(entry.PackOffset))
	if err != nil {
		return nil, err
	}

	return &Object{
		data: r,
		typ:  r.Type(),
	}, nil
}

// find finds and returns a Chain element corresponding to the offset of its
// last element as given by the "offset" argument.
//
// If find returns a ChainBase, it loads that data into memory, but does not
// zlib-flate it. Otherwise, if find returns a ChainDelta, it loads all of the
// leading elements in the chain recursively, but does not apply one delta to
// another.
func (p *Packfile) find(offset int64) (Chain, error) {
	// Read the first byte in the chain element.
	buf := make([]byte, 1)
	if _, err := p.r.ReadAt(buf, offset); err != nil {
		return nil, err
	}

	// Store the original offset; this will be compared to when loading
	// chain elements of type OBJ_OFS_DELTA.
	objectOffset := offset

	// Of the first byte, (0123 4567):
	//   - Bit 0 is the M.S.B., and indicates whether there is more data
	//     encoded in the length.
	//   - Bits 1-3 ((buf[0] >> 4) & 0x7) are the object type.
	//   - Bits 4-7 (buf[0] & 0xf) are the first 4 bits of the variable
	//     length size of the encoded delta or base.
	typ := PackedObjectType((buf[0] >> 4) & 0x7)
	size := uint64(buf[0] & 0xf)
	shift := uint(4)
	offset += 1

	for buf[0]&0x80 != 0 {
		// If there is more data to be read, read it.
		if _, err := p.r.ReadAt(buf, offset); err != nil {
			return nil, err
		}

		// And update the size, bitshift, and offset accordingly.
		size |= (uint64(buf[0]&0x7f) << shift)
		shift += 7
		offset += 1
	}

	switch typ {
	case TypeObjectOffsetDelta, TypeObjectReferenceDelta:
		// If the type of delta-base element is a delta, (either
		// OBJ_OFS_DELTA, or OBJ_REFS_DELTA), we must load the base,
		// which itself could be either of the two above, or a
		// OBJ_COMMIT, OBJ_BLOB, etc.
		//
		// Recursively load the base, and keep track of the updated
		// offset.
		base, offset, err := p.findBase(typ, offset, objectOffset)
		if err != nil {
			return nil, err
		}

		// Now load the delta to apply to the base, given at the offset
		// "offset" and for length "size".
		//
		// NB: The delta instructions are zlib compressed, so ensure
		// that we uncompress the instructions first.
		zr, err := zlib.NewReader(&OffsetReaderAt{
			o: offset,
			r: p.r,
		})
		if err != nil {
			return nil, err
		}

		delta, err := ioutil.ReadAll(zr)
		if err != nil {
			return nil, err
		}

		// Then compose the two and return it as a *ChainDelta.
		return &ChainDelta{
			base:  base,
			delta: delta,
		}, nil
	case TypeCommit, TypeTree, TypeBlob, TypeTag:
		// Otherwise, the object's contents are given to be the
		// following zlib-compressed data.
		//
		// The length of the compressed data itself is not known,
		// rather, "size" determines the length of the data after
		// inflation.
		return &ChainBase{
			offset: offset,
			size:   int64(size),
			typ:    typ,

			r: p.r,
		}, nil
	}
	// Otherwise, we received an invalid object type.
	return nil, errUnrecognizedObjectType
}

// findBase finds the base (an object, or another delta) for a given
// OBJ_OFS_DELTA or OBJ_REFS_DELTA at the given offset.
//
// It returns the preceding Chain, as well as an updated read offset into the
// underlying packfile data.
//
// If any of the above could not be completed successfully, findBase returns an
// error.
func (p *Packfile) findBase(typ PackedObjectType, offset, objOffset int64) (Chain, int64, error) {
	var baseOffset int64

	hashlen := p.hash.Size()

	// We assume that we have to read at least an object ID's worth (the
	// hash length in the case of a OBJ_REF_DELTA, or greater than the
	// length of the base offset encoded in an OBJ_OFS_DELTA).
	var sha [32]byte
	if _, err := p.r.ReadAt(sha[:hashlen], offset); err != nil {
		return nil, baseOffset, err
	}

	switch typ {
	case TypeObjectOffsetDelta:
		// If the object is of type OBJ_OFS_DELTA, read a
		// variable-length integer, and find the object at that
		// location.
		i := 0
		c := int64(sha[i])
		baseOffset = c & 0x7f

		for c&0x80 != 0 {
			i += 1
			c = int64(sha[i])

			baseOffset += 1
			baseOffset <<= 7
			baseOffset |= c & 0x7f
		}

		baseOffset = objOffset - baseOffset
		offset += int64(i) + 1
	case TypeObjectReferenceDelta:
		// If the delta is an OBJ_REFS_DELTA, find the location of its
		// base by reading the SHA-1 name and looking it up in the
		// corresponding pack index file.
		e, err := p.idx.Entry(sha[:hashlen])
		if err != nil {
			return nil, baseOffset, err
		}

		baseOffset = int64(e.PackOffset)
		offset += int64(hashlen)
	default:
		// If we did not receive an OBJ_OFS_DELTA, or OBJ_REF_DELTA, the
		// type given is not a delta-fied type. Return an error.
		return nil, baseOffset, fmt.Errorf(
			"gitobj/pack: type %s is not deltafied", typ)
	}

	// Once we have determined the base offset of the object's chain base,
	// read the delta-base chain beginning at that offset.
	r, err := p.find(baseOffset)
	return r, offset, err
}
