package pack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetOpenOpensAPackedObject(t *testing.T) {
	const sha = "decafdecafdecafdecafdecafdecafdecafdecaf"
	const data = "Hello, world!\n"
	compressed, _ := compress(data)

	set := NewSetPacks(&Packfile{
		idx: IndexWith(map[string]uint32{
			sha: 0,
		}),
		r: bytes.NewReader(append([]byte{0x3e}, compressed...)),
	})

	o, err := set.Object(DecodeHex(t, sha))

	assert.NoError(t, err)
	assert.Equal(t, TypeBlob, o.Type())

	unpacked, err := o.Unpack()
	assert.NoError(t, err)
	assert.Equal(t, []byte(data), unpacked)
}

func TestSetOpenOpensPackedObjectsInPackOrder(t *testing.T) {
	p1 := &Packfile{
		Objects: 1,

		idx: IndexWith(map[string]uint32{
			"aa00000000000000000000000000000000000000": 1,
		}),
		r: bytes.NewReader(nil),
	}
	p2 := &Packfile{
		Objects: 2,

		idx: IndexWith(map[string]uint32{
			"aa11111111111111111111111111111111111111": 1,
			"aa22222222222222222222222222222222222222": 2,
		}),
		r: bytes.NewReader(nil),
	}
	p3 := &Packfile{
		Objects: 3,

		idx: IndexWith(map[string]uint32{
			"aa33333333333333333333333333333333333333": 3,
			"aa44444444444444444444444444444444444444": 4,
			"aa55555555555555555555555555555555555555": 5,
		}),
		r: bytes.NewReader(nil),
	}

	set := NewSetPacks(p1, p2, p3)

	var visited []*Packfile

	set.each(
		DecodeHex(t, "aa55555555555555555555555555555555555555"),
		func(p *Packfile) (*Object, error) {
			visited = append(visited, p)
			return nil, errNotFound
		},
	)

	require.Len(t, visited, 3)
	assert.EqualValues(t, visited[0].Objects, 3)
	assert.EqualValues(t, visited[1].Objects, 2)
	assert.EqualValues(t, visited[2].Objects, 1)
}
