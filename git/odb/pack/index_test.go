package pack

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/git-lfs/git-lfs/errors"

	"github.com/stretchr/testify/assert"
)

var (
	idx *Index
)

func TestIndexEntrySearch(t *testing.T) {
	e, err := idx.Entry([]byte{
		0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,
		0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,
	})

	assert.NoError(t, err)
	assert.EqualValues(t, 6, e.PackOffset)
}

func TestIndexEntrySearchClampLeft(t *testing.T) {
	e, err := idx.Entry([]byte{
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	})

	assert.NoError(t, err)
	assert.EqualValues(t, 0, e.PackOffset)
}

func TestIndexEntrySearchClampRight(t *testing.T) {
	e, err := idx.Entry([]byte{
		0xff, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04,
		0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04,
	})

	assert.NoError(t, err)
	assert.EqualValues(t, 0x4ff, e.PackOffset)
}

func TestIndexSearchOutOfBounds(t *testing.T) {
	e, err := idx.Entry([]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	})

	assert.True(t, IsNotFound(err), "expected err to be 'not found'")
	assert.Nil(t, e)
}

func TestIndexEntryNotFound(t *testing.T) {
	e, err := idx.Entry([]byte{
		0x1, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6,
		0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x6,
	})

	assert.True(t, IsNotFound(err), "expected err to be 'not found'")
	assert.Nil(t, e)
}

func TestIndexCount(t *testing.T) {
	fanout := make([]uint32, 256)
	for i := 0; i < len(fanout); i++ {
		fanout[i] = uint32(i)
	}

	idx := &Index{fanout: fanout}

	assert.EqualValues(t, 255, idx.Count())
}

func TestIndexIsNotFound(t *testing.T) {
	assert.True(t, IsNotFound(errNotFound),
		"expected 'errNotFound' to satisfy 'IsNotFound()'")
}

func TestIndexIsNotFoundForOtherErrors(t *testing.T) {
	assert.False(t, IsNotFound(errors.New("git/odb/pack: misc")),
		"expected 'err' not to satisfy 'IsNotFound()'")
}

// init generates some fixture data and then constructs an *Index instance using
// it.
func init() {
	// eps is the number of SHA1 names generated under each 0x<t> slot.
	const eps = 5

	hdr := []byte{
		0xff, 0x74, 0x4f, 0x63, // Index file v2+ magic header
		0x00, 0x00, 0x00, 0x02, // 4-byte version indicator
	}

	// Create a fanout table using uint32s (later marshalled using
	// binary.BigEndian).
	//
	// Since we have an even distribution of SHA1s in the generated index,
	// each entry will increase by the number of entries per slot (see: eps
	// above).
	fanout := make([]uint32, indexFanoutEntries)
	for i := 0; i < len(fanout); i++ {
		// Begin the index at (i+1), since the fanout table mandates
		// objects less than the value at index "i".
		fanout[i] = uint32((i + 1) * eps)
	}

	offs := make([]uint32, 0, 256*eps)
	crcs := make([]uint32, 0, 256*eps)

	names := make([][]byte, 0, 256*eps)
	for i := 0; i < 256; i++ {
		// For each name, generate a unique SHA using the prefix "i",
		// and then suffix "j".
		//
		// In other words, when i=1, we will generate:
		//   []byte{0x1 0x0 0x0 0x0 ...}
		//   []byte{0x1 0x1 0x1 0x1 ...}
		//   []byte{0x1 0x2 0x2 0x2 ...}
		//
		// and etc.
		for j := 0; j < eps; j++ {
			var sha [20]byte

			sha[0] = byte(i)
			for r := 1; r < len(sha); r++ {
				sha[r] = byte(j)
			}

			cpy := make([]byte, len(sha))
			copy(cpy, sha[:])

			names = append(names, cpy)
			offs = append(offs, uint32((i*eps)+j))
			crcs = append(crcs, 0)
		}
	}

	// Create a buffer to hold the index contents:
	buf := bytes.NewBuffer(hdr)

	// Write each value in the fanout table using a 32bit network byte-order
	// integer.
	for _, f := range fanout {
		binary.Write(buf, binary.BigEndian, f)
	}
	// Write each SHA1 name to the table next.
	for _, name := range names {
		buf.Write(name)
	}
	// Then write each of the CRC values in network byte-order as a 32bit
	// unsigned integer.
	for _, crc := range crcs {
		binary.Write(buf, binary.BigEndian, crc)
	}
	// Do the same with the offsets.
	for _, off := range offs {
		binary.Write(buf, binary.BigEndian, off)
	}

	idx = &Index{
		fanout: fanout,
		// version is unimportant here, use V2 since it's more common in
		// the wild.
		version: new(V2),

		// *bytes.Buffer does not implement io.ReaderAt, but
		// *bytes.Reader does.
		//
		// Call (*bytes.Buffer).Bytes() to get the data, and then
		// construct a new *bytes.Reader with it to implement
		// io.ReaderAt.
		r: bytes.NewReader(buf.Bytes()),
	}
}
