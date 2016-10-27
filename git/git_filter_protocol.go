// Package git contains various commands that shell out to git
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package git

import (
	"fmt"
	"io"
	"strings"

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
		return fmt.Errorf("reading filter initialization failed with %s", err)
	}
	if initMsg != "git-filter-client" {
		return fmt.Errorf("invalid filter protocol welcome message: %s", initMsg)
	}

	supVers, err := o.p.readPacketList()
	if err != nil {
		return fmt.Errorf("reading filter versions failed with %s", err)
	}
	if !isStringInSlice(supVers, reqVer) {
		return fmt.Errorf("filter '%s' not supported (your Git supports: %s)", reqVer, supVers)
	}

	err = o.p.writePacketList([]string{"git-filter-server", reqVer})
	if err != nil {
		return fmt.Errorf("writing filter initialization failed with %s", err)
	}
	return nil
}

func (o *ObjectScanner) NegotiateCapabilities() error {
	reqCaps := []string{"capability=clean", "capability=smudge"}

	supCaps, err := o.p.readPacketList()
	if err != nil {
		return fmt.Errorf("Error: reading filter capabilities failed with %s", err)
	}
	for _, reqCap := range reqCaps {
		if !isStringInSlice(supCaps, reqCap) {
			return fmt.Errorf("Error: filter '%s' not supported (your Git supports: %s)", reqCap, supCaps)
		}
	}

	err = o.p.writePacketList(reqCaps)
	if err != nil {
		return fmt.Errorf("Error: writing filter capabilities failed with %s", err)
	}

	return nil
}

func (o *ObjectScanner) ReadRequest() (map[string]string, []byte, error) {
	tracerx.Printf("Process filter command.")

	requestList, err := o.p.readPacketList()
	if err != nil {
		return nil, nil, err
	}

	requestMap := make(map[string]string)
	for _, pair := range requestList {
		v := strings.Split(pair, "=")
		requestMap[v[0]] = v[1]
	}

	var data []byte
	for {
		chunk, err := o.p.readPacket()
		if err != nil {
			// TODO: should we check the err of this call, to?!
			o.writeStatus("error")
			return nil, nil, err
		}
		if len(chunk) == 0 {
			break
		}
		data = append(data, chunk...) // probably more efficient way?!
	}
	o.writeStatus("success")
	return requestMap, data, nil
}

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
			// TODO: should we check the err of this call, to?!
			o.writeStatus("error")
			return err
		}
		outputData = outputData[chunkSize:]
	}
	o.writeStatus("success")
	return nil
}

func (o *ObjectScanner) writeStatus(status string) error {
	return o.p.writePacketList([]string{"status=" + status})
}
