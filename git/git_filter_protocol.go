// Package git contains various commands that shell out to git
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package git

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/github/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

const (
	MaxPacketLenght = 65516
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
	r *bufio.Reader
	w *bufio.Writer
}

func NewObjectScanner(r io.Reader, w io.Writer) *ObjectScanner {
	return &ObjectScanner{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

func (o *ObjectScanner) readPacket() ([]byte, error) {
	pktLenHex, err := ioutil.ReadAll(io.LimitReader(o.r, 4))
	if err != nil || len(pktLenHex) != 4 { // TODO check pktLenHex length
		return nil, err
	}
	pktLen, err := strconv.ParseInt(string(pktLenHex), 16, 0)
	if err != nil {
		return nil, err
	}
	if pktLen == 0 {
		return nil, nil
	} else if pktLen <= 4 {
		return nil, errors.New("Invalid packet length.")
	}
	return ioutil.ReadAll(io.LimitReader(o.r, pktLen-4))
}

func (o *ObjectScanner) readPacketText() (string, error) {
	data, err := o.readPacket()
	return strings.TrimSuffix(string(data), "\n"), err
}

func (o *ObjectScanner) readPacketList() ([]string, error) {
	var list []string
	for {
		data, err := o.readPacketText()
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

func (o *ObjectScanner) writePacket(data []byte) error {
	if len(data) > MaxPacketLenght {
		return errors.New("Packet length exceeds maximal length")
	}
	_, err := o.w.WriteString(fmt.Sprintf("%04x", len(data)+4))
	if err != nil {
		return err
	}
	_, err = o.w.Write(data)
	if err != nil {
		return err
	}
	err = o.w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (o *ObjectScanner) writeFlush() error {
	_, err := o.w.WriteString(fmt.Sprintf("%04x", 0))
	if err != nil {
		return err
	}
	err = o.w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (o *ObjectScanner) writePacketText(data string) error {
	//TODO: there is probably a more efficient way to do this. worth it?
	return o.writePacket([]byte(data + "\n"))
}

func (o *ObjectScanner) writePacketList(list []string) error {
	for _, i := range list {
		err := o.writePacketText(i)
		if err != nil {
			return err
		}
	}
	return o.writeFlush()
}

func (o *ObjectScanner) writeStatus(status string) error {
	return o.writePacketList([]string{"status=" + status})
}

func (o *ObjectScanner) Init() bool {
	tracerx.Printf("Initialize filter")
	reqVer := "version=2"

	initMsg, err := o.readPacketText()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error: reading filter initialization failed with %s\n", err)
		return false
	}
	if initMsg != "git-filter-client" {
		fmt.Fprintf(os.Stderr,
			"Error: invalid filter protocol welcome message: %s\n", initMsg)
		return false
	}

	supVers, err := o.readPacketList()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error: reading filter versions failed with %s\n", err)
		return false
	}
	if !isStringInSlice(supVers, reqVer) {
		fmt.Fprintf(os.Stderr,
			"Error: filter '%s' not supported (your Git supports: %s)\n",
			reqVer, supVers)
		return false
	}

	err = o.writePacketList([]string{"git-filter-server", reqVer})
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error: writing filter initialization failed with %s\n", err)
		return false
	}
	return true
}

func (o *ObjectScanner) NegotiateCapabilities() bool {
	reqCaps := []string{"capability=clean", "capability=smudge"}

	supCaps, err := o.readPacketList()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error: reading filter capabilities failed with %s\n", err)
		return false
	}
	for _, reqCap := range reqCaps {
		if !isStringInSlice(supCaps, reqCap) {
			fmt.Fprintf(os.Stderr,
				"Error: filter '%s' not supported (your Git supports: %s)\n",
				reqCap, supCaps)
			return false
		}
	}

	err = o.writePacketList(reqCaps)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error: writing filter capabilities failed with %s\n", err)
		return false
	}

	return true
}

func (o *ObjectScanner) ReadRequest() (map[string]string, []byte, error) {
	tracerx.Printf("Process filter command.")

	requestList, err := o.readPacketList()
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
		chunk, err := o.readPacket()
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
			o.writeFlush()
			break
		} else if chunkSize > MaxPacketLenght {
			chunkSize = MaxPacketLenght // TODO check packets with the exact size
		}
		err := o.writePacket(outputData[:chunkSize])
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
