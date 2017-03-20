package git

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writePackets
func writePacket(t *testing.T, w io.Writer, datas ...[]byte) {
	pl := newPktline(nil, w)

	for _, data := range datas {
		require.Nil(t, pl.writePacket(data))
	}

	require.Nil(t, pl.writeFlush())
}

func TestPktlineReaderReadsSinglePacketsInOneCall(t *testing.T) {
	var buf bytes.Buffer

	writePacket(t, &buf, []byte("asdf"))

	pr := &pktlineReader{pl: newPktline(&buf, nil)}

	data, err := ioutil.ReadAll(pr)

	assert.Nil(t, err)
	assert.Equal(t, []byte("asdf"), data)
}

func TestPktlineReaderReadsManyPacketsInOneCall(t *testing.T) {
	var buf bytes.Buffer

	writePacket(t, &buf, []byte("first\n"), []byte("second"))

	pr := &pktlineReader{pl: newPktline(&buf, nil)}

	data, err := ioutil.ReadAll(pr)

	assert.Nil(t, err)
	assert.Equal(t, []byte("first\nsecond"), data)
}

func TestPktlineReaderReadsSinglePacketsInMultipleCallsWithUnevenBuffering(t *testing.T) {
	var buf bytes.Buffer

	writePacket(t, &buf, []byte("asdf"))

	pr := &pktlineReader{pl: newPktline(&buf, nil)}

	var p1 [3]byte
	var p2 [1]byte

	n1, e1 := pr.Read(p1[:])
	assert.Equal(t, 3, n1)
	assert.Equal(t, []byte("asd"), p1[:])
	assert.Nil(t, e1)

	n2, e2 := pr.Read(p2[:])
	assert.Equal(t, 1, n2)
	assert.Equal(t, []byte("f"), p2[:])
	assert.Equal(t, io.EOF, e2)
}

func TestPktlineReaderReadsManyPacketsInMultipleCallsWithUnevenBuffering(t *testing.T) {
	var buf bytes.Buffer

	writePacket(t, &buf, []byte("first"), []byte("second"))

	pr := &pktlineReader{pl: newPktline(&buf, nil)}

	var p1 [4]byte
	var p2 [7]byte
	var p3 []byte

	n1, e1 := pr.Read(p1[:])
	assert.Equal(t, 4, n1)
	assert.Equal(t, []byte("firs"), p1[:])
	assert.Nil(t, e1)

	n2, e2 := pr.Read(p2[:])
	assert.Equal(t, 7, n2)
	assert.Equal(t, []byte("tsecond"), p2[:])
	assert.Equal(t, io.EOF, e2)

	n3, e3 := pr.Read(p3[:])
	assert.Equal(t, 0, n3)
	assert.Empty(t, p3)
	assert.Equal(t, io.EOF, e3)
}

func TestPktlineReaderReadsSinglePacketsInMultipleCallsWithEvenBuffering(t *testing.T) {
	var buf bytes.Buffer

	writePacket(t, &buf, []byte("firstother"))

	pr := &pktlineReader{pl: newPktline(&buf, nil)}

	var p1 [5]byte
	var p2 [5]byte

	n1, e1 := pr.Read(p1[:])
	assert.Equal(t, 5, n1)
	assert.Equal(t, []byte("first"), p1[:])
	assert.Nil(t, e1)

	n2, e2 := pr.Read(p2[:])
	assert.Equal(t, 5, n2)
	assert.Equal(t, []byte("other"), p2[:])
	assert.Equal(t, io.EOF, e2)
}

func TestPktlineReaderReadsManyPacketsInMultipleCallsWithEvenBuffering(t *testing.T) {
	var buf bytes.Buffer

	writePacket(t, &buf, []byte("first"), []byte("other"))

	pr := &pktlineReader{pl: newPktline(&buf, nil)}

	var p1 [5]byte
	var p2 [5]byte
	var p3 []byte

	n1, e1 := pr.Read(p1[:])
	assert.Equal(t, 5, n1)
	assert.Equal(t, []byte("first"), p1[:])
	assert.Nil(t, e1)

	n2, e2 := pr.Read(p2[:])
	assert.Equal(t, 5, n2)
	assert.Equal(t, []byte("other"), p2[:])
	assert.Equal(t, io.EOF, e2)

	n3, e3 := pr.Read(p3)
	assert.Equal(t, 0, n3)
	assert.Equal(t, io.EOF, e3)
}
