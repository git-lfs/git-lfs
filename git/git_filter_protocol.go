// Package git contains various commands that shell out to git
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package git

import (
	"fmt"
	"io"
	"strings"

	"github.com/github/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

const (
	MaxPacketLength = 65516
)

// Private function copied from "github.com/xeipuuv/gojsonschema/utils.go"
// TODO: Is there a way to reuse this?
func isStringInSlice(s []string, what string) bool {
	for i := range s {
		if s[i] == what {
			return true
		}
	}
	return false
}

type ObjectScanner struct {
	p *protocol

	req *Request
	err error
}

func NewObjectScanner(r io.Reader, w io.Writer) *ObjectScanner {
	return &ObjectScanner{
		p: newProtocolRW(r, w),
	}
}

func (o *ObjectScanner) Init() error {
	tracerx.Printf("Initialize filter")
	reqVer := "version=2"

	initMsg, err := o.p.readPacketText()
	if err != nil {
		return errors.Wrap(err, "reading filter initialization")
	}
	if initMsg != "git-filter-client" {
		return fmt.Errorf("invalid filter protocol welcome message: %s", initMsg)
	}

	supVers, err := o.p.readPacketList()
	if err != nil {
		return errors.Wrap(err, "reading filter versions")
	}
	if !isStringInSlice(supVers, reqVer) {
		return fmt.Errorf("filter '%s' not supported (your Git supports: %s)", reqVer, supVers)
	}

	err = o.p.writePacketList([]string{"git-filter-server", reqVer})
	if err != nil {
		return errors.Wrap(err, "writing filter initialization failed")
	}
	return nil
}

func (o *ObjectScanner) NegotiateCapabilities() error {
	reqCaps := []string{"capability=clean", "capability=smudge"}

	supCaps, err := o.p.readPacketList()
	if err != nil {
		return fmt.Errorf("reading filter capabilities failed with %s", err)
	}
	for _, reqCap := range reqCaps {
		if !isStringInSlice(supCaps, reqCap) {
			return fmt.Errorf("filter '%s' not supported (your Git supports: %s)", reqCap, supCaps)
		}
	}

	err = o.p.writePacketList(reqCaps)
	if err != nil {
		return fmt.Errorf("writing filter capabilities failed with %s", err)
	}

	return nil
}

type Request struct {
	Header  map[string]string
	Payload []byte
}

func (o *ObjectScanner) Scan() bool {
	o.req, o.err = nil, nil

	req, err := o.readRequest()
	if err != nil {
		o.err = err
		return false
	}

	o.req = req
	return true
}

func (o *ObjectScanner) Request() *Request { return o.req }
func (o *ObjectScanner) Err() error        { return o.err }

func (o *ObjectScanner) WriteResponse(outputData []byte) error {
	for {
		chunkSize := len(outputData)
		if chunkSize == 0 {
			o.p.writeFlush()
			break
		} else if chunkSize > MaxPacketLength {
			chunkSize = MaxPacketLength // TODO check packets with the exact size
		}
		err := o.p.writePacket(outputData[:chunkSize])
		if err != nil {
			if werr := o.writeStatus("error"); werr != nil {
				return werr
			}

			return err
		}
		outputData = outputData[chunkSize:]
	}

	return o.writeStatus("success")
}

func (o *ObjectScanner) readRequest() (*Request, error) {
	tracerx.Printf("Process filter command.")

	requestList, err := o.p.readPacketList()
	if err != nil {
		return nil, err
	}

	req := &Request{
		Header: make(map[string]string),
	}

	for _, pair := range requestList {
		v := strings.Split(pair, "=")
		req.Header[v[0]] = v[1]
	}

	for {
		chunk, err := o.p.readPacket()
		if err != nil {
			if werr := o.writeStatus("error"); werr != nil {
				return nil, werr
			}

			return nil, err
		}
		if len(chunk) == 0 {
			break
		}
		req.Payload = append(req.Payload, chunk...) // probably more efficient way?!
	}

	if err := o.writeStatus("success"); err != nil {
		return nil, err
	}

	return req, nil
}

func (o *ObjectScanner) writeStatus(status string) error {
	return o.p.writePacketList([]string{"status=" + status})
}
