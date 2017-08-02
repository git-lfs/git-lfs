package git

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterProcessScannerInitializesWithCorrectSupportedValues(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	if err := pl.writePacketText("git-filter-client"); err != nil {
		t.Fatalf("expected... %v", err.Error())
	}

	require.Nil(t, pl.writePacketText("git-filter-client"))
	require.Nil(t, pl.writePacketList([]string{"version=2"}))

	fps := NewFilterProcessScanner(&from, &to)
	err := fps.Init()

	assert.Nil(t, err)

	out, err := newPktline(&to, nil).readPacketList()
	assert.Nil(t, err)
	assert.Equal(t, []string{"git-filter-server", "version=2"}, out)
}

func TestFilterProcessScannerRejectsUnrecognizedInitializationMessages(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	require.Nil(t, pl.writePacketText("git-filter-client-unknown"))
	require.Nil(t, pl.writeFlush())

	fps := NewFilterProcessScanner(&from, &to)
	err := fps.Init()

	require.NotNil(t, err)
	assert.Equal(t, "invalid filter-process pkt-line welcome message: git-filter-client-unknown", err.Error())
	assert.Empty(t, to.Bytes())
}

func TestFilterProcessScannerRejectsUnsupportedFilters(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	require.Nil(t, pl.writePacketText("git-filter-client"))
	// Write an unsupported version
	require.Nil(t, pl.writePacketList([]string{"version=0"}))

	fps := NewFilterProcessScanner(&from, &to)
	err := fps.Init()

	require.NotNil(t, err)
	assert.Equal(t, "filter 'version=2' not supported (your Git supports: [version=0])", err.Error())
	assert.Empty(t, to.Bytes())
}

func TestFilterProcessScannerNegotitatesSupportedCapabilities(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	require.Nil(t, pl.writePacketList([]string{
		"capability=clean", "capability=smudge", "capability=not-invented-yet",
	}))

	fps := NewFilterProcessScanner(&from, &to)
	err := fps.NegotiateCapabilities()

	assert.Nil(t, err)

	out, err := newPktline(&to, nil).readPacketList()
	assert.Nil(t, err)
	assert.Equal(t, []string{"capability=clean", "capability=smudge"}, out)
}

func TestFilterProcessScannerDoesNotNegotitatesUnsupportedCapabilities(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	// Write an unsupported capability
	require.Nil(t, pl.writePacketList([]string{
		"capability=unsupported",
	}))

	fps := NewFilterProcessScanner(&from, &to)
	err := fps.NegotiateCapabilities()

	require.NotNil(t, err)
	assert.Equal(t, "filter 'capability=clean' not supported (your Git supports: [capability=unsupported])", err.Error())
	assert.Empty(t, to.Bytes())
}

func TestFilterProcessScannerReadsRequestHeadersAndPayload(t *testing.T) {
	var from, to bytes.Buffer

	pl := newPktline(nil, &from)
	// Headers
	require.Nil(t, pl.writePacketList([]string{
		"foo=bar", "other=woot", "crazy='sq',\\$x=.bin",
	}))
	// Multi-line packet
	require.Nil(t, pl.writePacketText("first"))
	require.Nil(t, pl.writePacketText("second"))
	require.Nil(t, pl.writeFlush())

	req, err := readRequest(NewFilterProcessScanner(&from, &to))

	assert.Nil(t, err)
	assert.Equal(t, req.Header["foo"], "bar")
	assert.Equal(t, req.Header["other"], "woot")
	assert.Equal(t, req.Header["crazy"], "'sq',\\$x=.bin")

	payload, err := ioutil.ReadAll(req.Payload)
	assert.Nil(t, err)
	assert.Equal(t, []byte("first\nsecond\n"), payload)
}

func TestFilterProcessScannerRejectsInvalidHeaderPackets(t *testing.T) {
	from := bytes.NewBuffer([]byte{
		0x30, 0x30, 0x30, 0x34, // 0004 (invalid packet length)
	})

	req, err := readRequest(NewFilterProcessScanner(from, nil))

	require.NotNil(t, err)
	assert.Equal(t, "Invalid packet length.", err.Error())

	assert.Nil(t, req)
}

func TestFilterProcessScannerWritesLists(t *testing.T) {
	var to bytes.Buffer

	fps := NewFilterProcessScanner(nil, &to)
	err := fps.WriteList([]string{"hello", "goodbye"})

	assert.NoError(t, err)
	assert.Equal(t, "000ahello\n000cgoodbye\n0000", to.String())
}

// readRequest performs a single scan operation on the given
// `*FilterProcessScanner`, "s", and returns: an error if there was one, or a
// request if there was one.  If neither, it returns (nil, nil).
func readRequest(s *FilterProcessScanner) (*Request, error) {
	s.Scan()

	if err := s.Err(); err != nil {
		return nil, err
	}

	return s.Request(), nil
}
