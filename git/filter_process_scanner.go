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
func isStringInSlice(s []string, what string) bool {
	for i := range s {
		if s[i] == what {
			return true
		}
	}
	return false
}

type FilterProcessScanner struct {
	pl *pktline

	req *Request
	err error
}

func NewFilterProcessScanner(r io.Reader, w io.Writer) *FilterProcessScanner {
	return &FilterProcessScanner{
		pl: newPktline(r, w),
	}
}

func (o *FilterProcessScanner) Init() error {
	tracerx.Printf("Initialize filter")
	reqVer := "version=2"

	initMsg, err := o.pl.readPacketText()
	if err != nil {
		return errors.Wrap(err, "reading filter initialization")
	}
	if initMsg != "git-filter-client" {
		return fmt.Errorf("invalid filter pkt-line welcome message: %s", initMsg)
	}

	supVers, err := o.pl.readPacketList()
	if err != nil {
		return errors.Wrap(err, "reading filter versions")
	}
	if !isStringInSlice(supVers, reqVer) {
		return fmt.Errorf("filter '%s' not supported (your Git supports: %s)", reqVer, supVers)
	}

	err = o.pl.writePacketList([]string{"git-filter-server", reqVer})
	if err != nil {
		return errors.Wrap(err, "writing filter initialization failed")
	}
	return nil
}

func (o *FilterProcessScanner) NegotiateCapabilities() error {
	reqCaps := []string{"capability=clean", "capability=smudge"}

	supCaps, err := o.pl.readPacketList()
	if err != nil {
		return fmt.Errorf("reading filter capabilities failed with %s", err)
	}
	for _, reqCap := range reqCaps {
		if !isStringInSlice(supCaps, reqCap) {
			return fmt.Errorf("filter '%s' not supported (your Git supports: %s)", reqCap, supCaps)
		}
	}

	err = o.pl.writePacketList(reqCaps)
	if err != nil {
		return fmt.Errorf("writing filter capabilities failed with %s", err)
	}

	return nil
}

type Request struct {
	Header  map[string]string
	Payload io.Reader
}

func (o *FilterProcessScanner) Scan() bool {
	o.req, o.err = nil, nil

	req, err := o.readRequest()
	if err != nil {
		o.err = err
		return false
	}

	o.req = req
	return true
}

func (o *FilterProcessScanner) Request() *Request { return o.req }
func (o *FilterProcessScanner) Err() error        { return o.err }

func (o *FilterProcessScanner) WriteResponse(outputData []byte) error {
	for {
		chunkSize := len(outputData)
		if chunkSize == 0 {
			o.pl.writeFlush()
			break
		} else if chunkSize > MaxPacketLength {
			chunkSize = MaxPacketLength
		}
		err := o.pl.writePacket(outputData[:chunkSize])
		if err != nil {
			if werr := o.WriteStatus("error"); werr != nil {
				return werr
			}

			return err
		}
		outputData = outputData[chunkSize:]
	}

	return o.WriteStatus("success")
}

func (o *FilterProcessScanner) readRequest() (*Request, error) {
	tracerx.Printf("Read filter protocol request.")

	requestList, err := o.pl.readPacketList()
	if err != nil {
		return nil, err
	}

	req := &Request{
		Header:  make(map[string]string),
		Payload: &packetReader{pl: o.pl},
	}

	for _, pair := range requestList {
		v := strings.Split(pair, "=")
		req.Header[v[0]] = v[1]
	}

	return req, nil
}

func (o *FilterProcessScanner) WriteStatus(status string) error {
	return o.pl.writePacketList([]string{"status=" + status})
}
