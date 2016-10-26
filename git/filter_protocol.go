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

type protocol struct {
	r *bufio.Reader
	w *bufio.Writer
}

func newProtocolRW(r io.Reader, w io.Writer) *protocol {
	return &protocol{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

func (p *protocol) readPacket() ([]byte, error) {
	pktLenHex, err := ioutil.ReadAll(io.LimitReader(p.r, 4))
	if err != nil || len(pktLenHex) != 4 { // TODO check pktLenHex length
		return nil, err
	}

	pktLen, err := strconv.ParseInt(string(pktLenHex), 16, 0)
	if err != nil {
		return nil, err
	}

	if pktLen == 0 {
		return nil, nil
	}
	if pktLen <= 4 {
		return nil, errors.New("Invalid packet length.")
	}

	return ioutil.ReadAll(io.LimitReader(p.r, pktLen-4))
}

func (p *protocol) readPacketText() (string, error) {
	data, err := p.readPacket()
	return strings.TrimSuffix(string(data), "\n"), err
}

func (p *protocol) readPacketList() ([]string, error) {
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

func (p *protocol) writePacket(data []byte) error {
	if len(data) > MaxPacketLength {
		return errors.New("Packet length exceeds maximal length")
	}

	if _, err := p.w.WriteString(fmt.Sprintf("%04x", len(data)+4)); err != nil {
		return err
	}

	if _, err := p.w.Write(data); err != nil {
		return err
	}

	if err := p.w.Flush(); err != nil {
		return err
	}

	return nil
}

func (p *protocol) writeFlush() error {
	if _, err := p.w.WriteString(fmt.Sprintf("%04x", 0)); err != nil {
		return err
	}

	if err := p.w.Flush(); err != nil {
		return err
	}

	return nil
}

func (p *protocol) writePacketText(data string) error {
	//TODO: there is probably a more efficient way to do this. worth it?
	return p.writePacket([]byte(data + "\n"))
}

func (p *protocol) writePacketList(list []string) error {
	for _, i := range list {
		if err := p.writePacketText(i); err != nil {
			return err
		}
	}

	return p.writeFlush()
}
