package git

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	// MaxPacketLength is the maximum total (header+payload) length
	// encode-able within one packet using Git's pkt-line protocol.
	MaxPacketLength = 65516
)

type pktline struct {
	r *bufio.Reader
	w *bufio.Writer
}

func newPktline(r io.Reader, w io.Writer) *pktline {
	return &pktline{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

// readPacket reads a single packet entirely and returns the data encoded within
// it. Errors can occur in several cases, as described below.
//
// 1) If no data was present in the reader, and no more data could be read (the
//    pipe was closed, etc) than an io.EOF will be returned.
// 2) If there was some data to be read, but the pipe or reader was closed
//    before an entire packet (or header) could be ingested, an
//    io.ErrShortBuffer error will be returned.
// 3) If there was a valid header, but no body associated with the packet, an
//    "invalid packet length" error will be returned.
// 4) If the data in the header could not be parsed as a hexadecimal length in
//    the Git pktline format, the parse error will be returned.
//
// If none of the above cases fit the state of the data on the wire, the packet
// is returned along with a nil error.
func (p *pktline) readPacket() ([]byte, error) {
	var pktLenHex [4]byte
	if n, err := io.ReadFull(p.r, pktLenHex[:]); err != nil {
		return nil, err
	} else if n != 4 {
		return nil, io.ErrShortBuffer
	}

	pktLen, err := strconv.ParseInt(string(pktLenHex[:]), 16, 0)
	if err != nil {
		return nil, err
	}

	// pktLen==0: flush packet
	if pktLen == 0 {
		return nil, nil
	}
	if pktLen <= 4 {
		return nil, errors.New("invalid packet length")
	}

	payload, err := ioutil.ReadAll(io.LimitReader(p.r, pktLen-4))
	return payload, err
}

// readPacketText follows identical semantics to the `readPacket()` function,
// but additionally removes the trailing `\n` LF from the end of the packet, if
// present.
func (p *pktline) readPacketText() (string, error) {
	data, err := p.readPacket()
	return strings.TrimSuffix(string(data), "\n"), err
}

// readPacketList reads as many packets as possible using the `readPacketText`
// function before encountering a flush packet. It returns a slice of all the
// packets it read, or an error if one was encountered.
func (p *pktline) readPacketList() ([]string, error) {
	var list []string
	for {
		data, err := p.readPacketText()
		if err != nil {
			return nil, err
		}

		if len(data) == 0 {
			break
		}

		list = append(list, data)
	}

	return list, nil
}

// writePacket writes the given data in "data" to the underlying data stream
// using Git's `pkt-line` format.
//
// If the data was longer than MaxPacketLength, an error will be returned. If
// there was any error encountered while writing any component of the packet
// (hdr, payload), it will be returned.
//
// NB: writePacket does _not_ flush the underlying buffered writer. See instead:
// `writeFlush()`.
func (p *pktline) writePacket(data []byte) error {
	if len(data) > MaxPacketLength {
		return errors.New("packet length exceeds maximal length")
	}

	if _, err := p.w.WriteString(fmt.Sprintf("%04x", len(data)+4)); err != nil {
		return err
	}

	if _, err := p.w.Write(data); err != nil {
		return err
	}

	return nil
}

// writeFlush writes the terminating "flush" packet and then flushes the
// underlying buffered writer.
//
// If any error was encountered along the way, it will be returned immediately
func (p *pktline) writeFlush() error {
	if _, err := p.w.WriteString(fmt.Sprintf("%04x", 0)); err != nil {
		return err
	}

	if err := p.w.Flush(); err != nil {
		return err
	}

	return nil
}

// writePacketText follows the same semantics as `writePacket`, but appends a
// trailing "\n" LF character to the end of the data.
func (p *pktline) writePacketText(data string) error {
	return p.writePacket([]byte(data + "\n"))
}

// writePacketList writes a slice of strings using the semantics of
// and then writes a terminating flush sequence afterwords.
//
// If any error was encountered, it will be returned immediately.
func (p *pktline) writePacketList(list []string) error {
	for _, i := range list {
		if err := p.writePacketText(i); err != nil {
			return err
		}
	}

	return p.writeFlush()
}
