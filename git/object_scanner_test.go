package git

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectScannerInitializesWithCorrectSupportedValues(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	if err := pl.writePacketText("git-filter-client"); err != nil {
		t.Fatalf("expected... %v", err.Error())
	}

	require.Nil(t, pl.writePacketText("git-filter-client"))
	require.Nil(t, pl.writePacketList([]string{"version=2"}))

	os := NewObjectScanner(&from, &to)
	err := os.Init()

	assert.Nil(t, err)

	out, err := newPktline(&to, nil).readPacketList()
	assert.Nil(t, err)
	assert.Equal(t, []string{"git-filter-server", "version=2"}, out)
}

func TestObjectScannerRejectsUnrecognizedInitializationMessages(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	require.Nil(t, pl.writePacketText("git-filter-client-unknown"))

	os := NewObjectScanner(&from, &to)
	err := os.Init()

	require.NotNil(t, err)
	assert.Equal(t, "invalid filter pkt-line welcome message: git-filter-client-unknown", err.Error())
	assert.Empty(t, to.Bytes())
}

func TestObjectScannerRejectsUnsupportedFilters(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	require.Nil(t, pl.writePacketText("git-filter-client"))
	// Write an unsupported version
	require.Nil(t, pl.writePacketList([]string{"version=0"}))

	os := NewObjectScanner(&from, &to)
	err := os.Init()

	require.NotNil(t, err)
	assert.Equal(t, "filter 'version=2' not supported (your Git supports: [version=0])", err.Error())
	assert.Empty(t, to.Bytes())
}

func TestObjectScannerNegotitatesSupportedCapabilities(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	require.Nil(t, pl.writePacketList([]string{
		"capability=clean", "capability=smudge", "capability=not-invented-yet",
	}))

	os := NewObjectScanner(&from, &to)
	err := os.NegotiateCapabilities()

	assert.Nil(t, err)

	out, err := newPktline(&to, nil).readPacketList()
	assert.Nil(t, err)
	assert.Equal(t, []string{"capability=clean", "capability=smudge"}, out)
}

func TestObjectScannerDoesNotNegotitatesUnsupportedCapabilities(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	// Write an unsupported capability
	require.Nil(t, pl.writePacketList([]string{
		"capability=unsupported",
	}))

	os := NewObjectScanner(&from, &to)
	err := os.NegotiateCapabilities()

	require.NotNil(t, err)
	assert.Equal(t, "filter 'capability=clean' not supported (your Git supports: [capability=unsupported])", err.Error())
	assert.Empty(t, to.Bytes())
}

func TestObjectScannerReadsRequestHeadersAndPayload(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	// Headers
	require.Nil(t, pl.writePacketList([]string{
		"foo=bar", "other=woot",
	}))
	// Multi-line packet
	require.Nil(t, pl.writePacketText("first"))
	require.Nil(t, pl.writePacketText("second"))
	_, err := from.Write([]byte{0x30, 0x30, 0x30, 0x30}) // flush packet
	assert.Nil(t, err)

	req, err := readRequest(NewObjectScanner(&from, &to))

	assert.Nil(t, err)
	assert.Equal(t, req.Header["foo"], "bar")
	assert.Equal(t, req.Header["other"], "woot")

	payload, err := ioutil.ReadAll(req.Payload)
	assert.Nil(t, err)
	assert.Equal(t, []byte("first\nsecond\n"), payload)
}

func TestObjectScannerRejectsInvalidHeaderPackets(t *testing.T) {
	var from bytes.Buffer

	pl := newPktline(nil, &from)
	// (Invalid) headers
	require.Nil(t, pl.writePacket([]byte{}))

	req, err := readRequest(NewObjectScanner(&from, nil))

	require.NotNil(t, err)
	assert.Equal(t, "Invalid packet length.", err.Error())

	assert.Nil(t, req)
}

func TestObjectScannerWritesResponsesInOneChunk(t *testing.T) {
	var buf bytes.Buffer

	err := NewObjectScanner(nil, &buf).WriteResponse([]byte(
		"hello world",
	))

	assert.Nil(t, err)

	pl := newPktline(&buf, nil)

	payload, err := pl.readPacket()
	assert.Nil(t, err)
	assert.Equal(t, []byte("hello world"), payload)

	// read terminating packet
	_, err = pl.readPacket()
	assert.Nil(t, err)

	status, err := pl.readPacketList()
	assert.Nil(t, err)
	assert.Equal(t, []string{"status=success"}, status)
}

func TestObjectScannerWritesEmptyResponses(t *testing.T) {
	var buf bytes.Buffer

	err := NewObjectScanner(nil, &buf).WriteResponse([]byte{})

	assert.Nil(t, err)

	pl := newPktline(&buf, nil)

	payload, err := pl.readPacket()
	assert.Nil(t, err)
	assert.Empty(t, payload)

	status, err := pl.readPacketList()
	assert.Nil(t, err)
	assert.Equal(t, []string{"status=success"}, status)
}

func TestObjectScannerWritesResponsesInMultipleChunks(t *testing.T) {
	payload := make([]byte, MaxPacketLength*2)
	for i := 0; i < 2; i++ {
		for j := 0; j < MaxPacketLength; j++ {
			payload[(i*MaxPacketLength)+j] = byte(i)
		}
	}

	var buf bytes.Buffer

	err := NewObjectScanner(nil, &buf).WriteResponse(payload)
	assert.Nil(t, err)

	pl := newPktline(&buf, nil)

	for i := 0; i < 2; i++ {
		pkt, err := pl.readPacket()
		assert.Nil(t, err)

		part := make([]byte, MaxPacketLength)
		for j := 0; j < len(part); j++ {
			part[j] = byte(i)
		}

		assert.Equal(t, part, pkt)
	}

	// read empty packet after flushing
	_, err = pl.readPacket()

	status, err := pl.readPacketList()
	assert.Nil(t, err)
	assert.Equal(t, []string{"status=success"}, status)
}

// readRequest preforms a single scan operation on the given `*ObjectScanner`,
// "s", and returns: an error if there was one, or a request if there was one.
// If neither, it returns (nil, nil).
func readRequest(s *ObjectScanner) (*Request, error) {
	s.Scan()

	if err := s.Err(); err != nil {
		return nil, err
	}

	return s.Request(), nil
}
