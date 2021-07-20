package pktline

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

type Pktline struct {
	r *bufio.Reader
	w *bufio.Writer
}

func NewPktline(r io.Reader, w io.Writer) *Pktline {
	return &Pktline{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

// ReadPacket reads a single packet entirely and returns the data encoded within
// it. Errors can occur in several cases, as described below.
//
// 1) If no data was present in the reader, and no more data could be read (the
//    pipe was closed, etc) than an io.EOF will be returned.
// 2) If there was some data to be read, but the pipe or reader was closed
//    before an entire packet (or header) could be ingested, an
//    io.ErrShortBuffer error will be returned.
// 3) If there was a valid header, but no body associated with the packet, an
//    "Invalid packet length." error will be returned.
// 4) If the data in the header could not be parsed as a hexadecimal length in
//    the Git Pktline format, the parse error will be returned.
//
// If none of the above cases fit the state of the data on the wire, the packet
// is returned along with a nil error.
func (p *Pktline) ReadPacket() ([]byte, error) {
	slice, _, err := p.ReadPacketWithLength()
	return slice, err
}

// ReadPacketWithLength is exactly like ReadPacket, but on success, it also
// returns the packet length header value.  This is useful to distinguish
// between flush and delim packets, which will return 0 and 1 respectively.  For
// data packets, the length will be four more than the number of bytes in the
// slice.
func (p *Pktline) ReadPacketWithLength() ([]byte, int, error) {
	var pktLenHex [4]byte
	if n, err := io.ReadFull(p.r, pktLenHex[:]); err != nil {
		return nil, 0, err
	} else if n != 4 {
		return nil, 0, io.ErrShortBuffer
	}

	pktLen, err := strconv.ParseInt(string(pktLenHex[:]), 16, 0)
	if err != nil {
		return nil, 0, err
	}

	// 0: flush packet, 1: delim packet
	if pktLen == 0 || pktLen == 1 {
		return nil, int(pktLen), nil
	}
	if pktLen < 4 {
		return nil, int(pktLen), errors.New("Invalid packet length.")
	}

	payload, err := ioutil.ReadAll(io.LimitReader(p.r, pktLen-4))
	return payload, int(pktLen), err
}

// ReadPacketText follows identical semantics to the `readPacket()` function,
// but additionally removes the trailing `\n` LF from the end of the packet, if
// present.
func (p *Pktline) ReadPacketText() (string, error) {
	data, err := p.ReadPacket()
	return strings.TrimSuffix(string(data), "\n"), err
}

// ReadPacketTextWithLength follows identical semantics to the
// `ReadPacketWithLength()` function, but additionally removes the trailing `\n`
// LF from the end of the packet, if present.  The length field is not modified.
func (p *Pktline) ReadPacketTextWithLength() (string, int, error) {
	data, pktLen, err := p.ReadPacketWithLength()
	return strings.TrimSuffix(string(data), "\n"), pktLen, err
}

// ReadPacketList reads as many packets as possible using the `readPacketText`
// function before encountering a flush packet. It returns a slice of all the
// packets it read, or an error if one was encountered.
func (p *Pktline) ReadPacketList() ([]string, error) {
	var list []string
	for {
		data, pktLen, err := p.ReadPacketTextWithLength()
		if err != nil {
			return nil, err
		}

		if pktLen == 0 {
			break
		}

		list = append(list, data)
	}

	return list, nil
}

// WritePacket writes the given data in "data" to the underlying data stream
// using Git's `pkt-line` format.
//
// If the data was longer than MaxPacketLength, an error will be returned. If
// there was any error encountered while writing any component of the packet
// (hdr, payload), it will be returned.
//
// NB: writePacket does _not_ flush the underlying buffered writer. See instead:
// `writeFlush()`.
func (p *Pktline) WritePacket(data []byte) error {
	if len(data) > MaxPacketLength {
		return errors.New("Packet length exceeds maximal length")
	}

	if _, err := p.w.WriteString(fmt.Sprintf("%04x", len(data)+4)); err != nil {
		return err
	}

	if _, err := p.w.Write(data); err != nil {
		return err
	}

	return nil
}

// WriteDelim writes the separating "delim" packet and then flushes the
// underlying buffered writer.
//
// If any error was encountered along the way, it will be returned immediately
func (p *Pktline) WriteDelim() error {
	if _, err := p.w.WriteString(fmt.Sprintf("%04x", 1)); err != nil {
		return err
	}

	if err := p.w.Flush(); err != nil {
		return err
	}

	return nil
}

// WriteFlush writes the terminating "flush" packet and then flushes the
// underlying buffered writer.
//
// If any error was encountered along the way, it will be returned immediately
func (p *Pktline) WriteFlush() error {
	if _, err := p.w.WriteString(fmt.Sprintf("%04x", 0)); err != nil {
		return err
	}

	if err := p.w.Flush(); err != nil {
		return err
	}

	return nil
}

// WritePacketText follows the same semantics as `writePacket`, but appends a
// trailing "\n" LF character to the end of the data.
func (p *Pktline) WritePacketText(data string) error {
	return p.WritePacket([]byte(data + "\n"))
}

// WritePacketList writes a slice of strings using the semantics of
// and then writes a terminating flush sequence afterwords.
//
// If any error was encountered, it will be returned immediately.
func (p *Pktline) WritePacketList(list []string) error {
	for _, i := range list {
		if err := p.WritePacketText(i); err != nil {
			return err
		}
	}

	return p.WriteFlush()
}
