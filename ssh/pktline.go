package ssh

import (
	"io"

	"github.com/git-lfs/pktline"
	"github.com/rubyist/tracerx"
)

func pktlineReader(p Pktline) io.Reader {
	if pl, ok := p.(*pktline.Pktline); ok {
		return pktline.NewPktlineReaderFromPktline(pl, 65536)
	}
	tp := p.(*TraceablePktline)
	return pktline.NewPktlineReaderFromPktline(tp.pl, 65536)
}

type Pktline interface {
	ReadPacketList() ([]string, error)
	ReadPacketTextWithLength() (string, int, error)
	WritePacket([]byte) error
	WritePacketText(string) error
	WriteDelim() error
	WriteFlush() error
}

type TraceablePktline struct {
	id int
	pl *pktline.Pktline
}

func (tp *TraceablePktline) ReadPacketList() ([]string, error) {
	var list []string
	for {
		data, pktLen, err := tp.pl.ReadPacketTextWithLength()
		if err != nil {
			return nil, err
		}
		if pktLen <= 1 {
			tracerx.Printf("packet %02x < %04x", tp.id, pktLen)
		} else {
			tracerx.Printf("packet %02x < %s", tp.id, data)
		}

		if pktLen == 0 {
			break
		}

		list = append(list, data)
	}

	return list, nil
}

func (tp *TraceablePktline) ReadPacketTextWithLength() (string, int, error) {
	s, pktLen, err := tp.pl.ReadPacketTextWithLength()
	if err != nil {
		return "", 0, err
	}

	if pktLen <= 1 {
		tracerx.Printf("packet %02x < %04x", tp.id, pktLen)
	} else {
		tracerx.Printf("packet %02x < %s", tp.id, s)
	}
	return s, pktLen, nil
}

func (tp *TraceablePktline) WritePacket(b []byte) error {
	// Don't trace because this is probably binary data.
	return tp.pl.WritePacket(b)
}

func (tp *TraceablePktline) WritePacketText(s string) error {
	tracerx.Printf("packet %02x > %s", tp.id, s)
	return tp.pl.WritePacketText(s)
}

func (tp *TraceablePktline) WriteDelim() error {
	tracerx.Printf("packet %02x > 0001", tp.id)
	return tp.pl.WriteDelim()
}

func (tp *TraceablePktline) WriteFlush() error {
	tracerx.Printf("packet %02x > 0000", tp.id)
	return tp.pl.WriteFlush()
}
