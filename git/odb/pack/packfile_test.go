package pack

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/git-lfs/git-lfs/errors"

	"github.com/stretchr/testify/assert"
)

func TestPackObjectReturnsObjectWithSingleBaseAtLowOffset(t *testing.T) {
	const original = "Hello, world!\n"
	compressed, _ := compress(original)

	p := &Packfile{
		idx: IndexWith(map[string]uint32{
			"cccccccccccccccccccccccccccccccccccccccc": 32,
		}),
		r: bytes.NewReader(append([]byte{
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,

			// (0001 1000) (msb=0, type=commit, size=14)
			0x1e}, compressed...),
		),
	}

	o, err := p.Object(DecodeHex(t, "cccccccccccccccccccccccccccccccccccccccc"))
	assert.NoError(t, err)

	assert.Equal(t, TypeCommit, o.Type())

	unpacked, err := o.Unpack()
	assert.Equal(t, []byte(original), unpacked)
	assert.NoError(t, err)
}

func TestPackObjectReturnsObjectWithSingleBaseAtHighOffset(t *testing.T) {
	original := strings.Repeat("four", 64)
	compressed, _ := compress(original)

	p := &Packfile{
		idx: IndexWith(map[string]uint32{
			"cccccccccccccccccccccccccccccccccccccccc": 32,
		}),
		r: bytes.NewReader(append([]byte{
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,

			// (1001 0000) (msb=1, type=commit, size=0)
			0x90,
			// (1000 0000) (msb=0, size=1 -> size=256)
			0x10},

			compressed...,
		)),
	}

	o, err := p.Object(DecodeHex(t, "cccccccccccccccccccccccccccccccccccccccc"))
	assert.NoError(t, err)

	assert.Equal(t, TypeCommit, o.Type())

	unpacked, err := o.Unpack()
	assert.Equal(t, []byte(original), unpacked)
	assert.NoError(t, err)
}

func TestPackObjectReturnsObjectWithDeltaBaseOffset(t *testing.T) {
	const original = "Hello"
	compressed, _ := compress(original)

	delta, err := compress(string([]byte{
		0x05, // Source size: 5.
		0x0e, // Destination size: 14.

		0x91, // (1000 0001) (instruction=copy, bitmask=0001)
		0x00, // (0000 0000) (offset=0)
		0x05, // (0000 0101) (size=5)

		0x09, // (0000 0111) (instruction=add, size=7)

		// Contents: ...
		',', ' ', 'w', 'o', 'r', 'l', 'd', '!', '\n',
	}))

	p := &Packfile{
		idx: IndexWith(map[string]uint32{
			"cccccccccccccccccccccccccccccccccccccccc": uint32(32 + 1 + len(compressed)),
		}),
		r: bytes.NewReader(append(append([]byte{
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,

			0x35, // (0011 0101) (msb=0, type=blob, size=5)
		}, compressed...), append([]byte{
			0x6e, // (0110 1010) (msb=0, type=obj_ofs_delta, size=10)
			0x12, // (0001 0001) (ofs_delta=-17, len(compressed))
		}, delta...)...)),
	}

	o, err := p.Object(DecodeHex(t, "cccccccccccccccccccccccccccccccccccccccc"))
	assert.NoError(t, err)

	assert.Equal(t, TypeBlob, o.Type())

	unpacked, err := o.Unpack()
	assert.Equal(t, []byte(original+", world!\n"), unpacked)
	assert.NoError(t, err)
}

func TestPackfileObjectReturnsObjectWithDeltaBaseReference(t *testing.T) {
	const original = "Hello!\n"
	compressed, _ := compress(original)

	delta, _ := compress(string([]byte{
		0x07, // Source size: 7.
		0x0e, // Destination size: 14.

		0x91, // (1001 0001) (copy, smask=0001, omask=0001)
		0x00, // (0000 0000) (offset=0)
		0x05, // (0000 0101) (size=5)

		0x7,                               // (0000 0111) (add, length=6)
		',', ' ', 'w', 'o', 'r', 'l', 'd', // (data ...)

		0x91, // (1001 0001) (copy, smask=0001, omask=0001)
		0x05, // (0000 0101) (offset=5)
		0x02, // (0000 0010) (size=2)
	}))

	p := &Packfile{
		idx: IndexWith(map[string]uint32{
			"cccccccccccccccccccccccccccccccccccccccc": 32,
			"dddddddddddddddddddddddddddddddddddddddd": 52,
		}),
		r: bytes.NewReader(append(append([]byte{
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,

			0x37, // (0011 0101) (msb=0, type=blob, size=7)
		}, compressed...), append([]byte{
			0x7f, // (0111 1111) (msb=0, type=obj_ref_delta, size=15)

			// SHA-1 "cccccccccccccccccccccccccccccccccccccccc",
			// original blob contents is "Hello!\n"
			0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc,
			0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc,
		}, delta...)...)),
	}

	o, err := p.Object(DecodeHex(t, "dddddddddddddddddddddddddddddddddddddddd"))
	assert.NoError(t, err)

	assert.Equal(t, TypeBlob, o.Type())

	unpacked, err := o.Unpack()
	assert.Equal(t, []byte("Hello, world!\n"), unpacked)
	assert.NoError(t, err)
}

func TestPackfileClosesReadClosers(t *testing.T) {
	r := new(ReaderAtCloser)
	p := &Packfile{
		r: r,
	}

	assert.NoError(t, p.Close())
	assert.EqualValues(t, 1, r.N)
}

func TestPackfileClosePropogatesCloseErrors(t *testing.T) {
	e := errors.New("git/odb/pack: testing")
	p := &Packfile{
		r: &ReaderAtCloser{E: e},
	}

	assert.Equal(t, e, p.Close())
}

type ReaderAtCloser struct {
	E error
	N uint64
}

func (r *ReaderAtCloser) ReadAt(p []byte, at int64) (int, error) {
	return 0, nil
}

func (r *ReaderAtCloser) Close() error {
	atomic.AddUint64(&r.N, 1)
	return r.E
}

func IndexWith(offsets map[string]uint32) *Index {
	header := []byte{
		0xff, 0x74, 0x4f, 0x63,
		0x00, 0x00, 0x00, 0x02,
	}

	ns := make([][]byte, 0, len(offsets))
	for name, _ := range offsets {
		x, _ := hex.DecodeString(name)
		ns = append(ns, x)
	}
	sort.Slice(ns, func(i, j int) bool {
		return bytes.Compare(ns[i], ns[j]) < 0
	})

	fanout := make([]uint32, 256)
	for i := 0; i < len(fanout); i++ {
		var n uint32

		for _, name := range ns {
			if name[0] <= byte(i) {
				n++
			}
		}

		fanout[i] = n
	}

	crcs := make([]byte, 4*len(offsets))
	for i, _ := range ns {
		binary.BigEndian.PutUint32(crcs[i*4:], 0)
	}

	offs := make([]byte, 4*len(offsets))
	for i, name := range ns {
		binary.BigEndian.PutUint32(offs[i*4:], offsets[hex.EncodeToString(name)])
	}

	buf := make([]byte, 0)
	buf = append(buf, header...)
	for _, f := range fanout {
		x := make([]byte, 4)
		binary.BigEndian.PutUint32(x, f)

		buf = append(buf, x...)
	}
	for _, n := range ns {
		buf = append(buf, n...)
	}
	buf = append(buf, crcs...)
	buf = append(buf, offs...)

	return &Index{
		fanout: fanout,
		r:      bytes.NewReader(buf),

		version: new(V2),
	}
}

func DecodeHex(t *testing.T, str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		t.Fatalf("git/odb/pack: unexpected hex.DecodeString error: %s", err)
	}

	return b
}
