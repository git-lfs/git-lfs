// Package git contains various commands that shell out to git
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package git

import (
	"fmt"
	"io"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

// FilterProcessScanner provides a scanner-like interface capable of
// initializing the filter process with the Git parent, and scanning for
// requests across the protocol.
//
// Reading a request (and errors) is as follows:
//
//     s := NewFilterProcessScanner(os.Stdin, os.Stderr)
//     for s.Scan() {
//             req := s.Request()
//     	       // ...
//     }
//
//     if err := s.Err(); err != nil {
//             // ...
//     }
type FilterProcessScanner struct {
	// pl is the *pktline instance used to read and write packets back and
	// forth between Git.
	pl *pktline

	// req is a temporary variable used to hold the value accessible by the
	// `Request()` function. It is cleared at the beginning of each `Scan()`
	// invocation, and written to at the end of each `Scan()` invocation.
	req *Request
	// err is a temporary variable used to hold the value accessible by the
	// `Request()` function. It is cleared at the beginning of each `Scan()`
	// invocation, and written to at the end of each `Scan()` invocation.
	err error
}

// NewFilterProcessScanner constructs a new instance of the
// `*FilterProcessScanner` type which reads packets from the `io.Reader` "r",
// and writes packets to the `io.Writer`, "w".
//
// Both reader and writers SHOULD NOT be `*git.PacketReader` or
// `*git.PacketWriter`s, they will be transparently treated as such. In other
// words, it is safe (and recommended) to pass `os.Stdin` and `os.Stdout`
// directly.
func NewFilterProcessScanner(r io.Reader, w io.Writer) *FilterProcessScanner {
	return &FilterProcessScanner{
		pl: newPktline(r, w),
	}
}

// Init initializes the filter and ACKs back and forth between the Git LFS
// subprocess and the Git parent process that each is a git-filter-server and
// client respectively.
//
// If either side wrote an invalid sequence of data, or did not meet
// expectations, an error will be returned. If the filter type is not supported,
// an error will be returned. If the pkt-line welcome message was invalid, an
// error will be returned.
//
// If there was an error reading or writing any of the packets below, an error
// will be returned.
func (o *FilterProcessScanner) Init() error {
	tracerx.Printf("Initialize filter-process")
	reqVer := "version=2"

	initMsg, err := o.pl.readPacketText()
	if err != nil {
		return errors.Wrap(err, "reading filter-process initialization")
	}
	if initMsg != "git-filter-client" {
		return fmt.Errorf("invalid filter-process pkt-line welcome message: %s", initMsg)
	}

	supVers, err := o.pl.readPacketList()
	if err != nil {
		return errors.Wrap(err, "reading filter-process versions")
	}
	if !isStringInSlice(supVers, reqVer) {
		return fmt.Errorf("filter '%s' not supported (your Git supports: %s)", reqVer, supVers)
	}

	err = o.pl.writePacketList([]string{"git-filter-server", reqVer})
	if err != nil {
		return errors.Wrap(err, "writing filter-process initialization failed")
	}
	return nil
}

// NegotiateCapabilities executes the process of negotiating capabilities
// between the filter client and server. If we don't support any of the
// capabilities given to LFS by Git, an error will be returned. If there was an
// error reading or writing capabilities between the two, an error will be
// returned.
func (o *FilterProcessScanner) NegotiateCapabilities() ([]string, error) {
	reqCaps := []string{"capability=clean", "capability=smudge"}

	supCaps, err := o.pl.readPacketList()
	if err != nil {
		return nil, fmt.Errorf("reading filter-process capabilities failed with %s", err)
	}

	for _, sup := range supCaps {
		if sup == "capability=delay" {
			reqCaps = append(reqCaps, "capability=delay")
			break
		}
	}

	for _, reqCap := range reqCaps {
		if !isStringInSlice(supCaps, reqCap) {
			return nil, fmt.Errorf("filter '%s' not supported (your Git supports: %s)", reqCap, supCaps)
		}
	}

	err = o.pl.writePacketList(reqCaps)
	if err != nil {
		return nil, fmt.Errorf("writing filter-process capabilities failed with %s", err)
	}

	return supCaps, nil
}

// Request represents a single command sent to LFS from the parent Git process.
type Request struct {
	// Header maps header strings to values, and is encoded as the first
	// part of the packet.
	Header map[string]string
	// Payload represents the body of the packet, and contains the contents
	// of the file in the index.
	Payload io.Reader
}

// Scan scans for the next request, or error and returns whether or not the scan
// was successful, indicating the presence of a valid request. If the Scan
// failed, there was either an error reading the next request (and the results
// of calling `Err()` should be inspected), or the pipe was closed and no more
// requests are present.
//
// Closing the pipe is Git's way to communicate that no more files will be
// filtered. Git expects that the LFS process exits after this event.
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

// Request returns the request read from a call to Scan(). It is available only
// after a call to `Scan()` has completed, and is re-initialized to nil at the
// beginning of the subsequent `Scan()` call.
func (o *FilterProcessScanner) Request() *Request { return o.req }

// Err returns any error encountered from the last call to Scan(). It is available only
// after a call to `Scan()` has completed, and is re-initialized to nil at the
// beginning of the subsequent `Scan()` call.
func (o *FilterProcessScanner) Err() error { return o.err }

// readRequest reads the headers of a request and yields an `io.Reader` which
// will read the body of the request. Since the body is _not_ offset, one
// request should be read in its entirety before consuming the next request.
func (o *FilterProcessScanner) readRequest() (*Request, error) {
	requestList, err := o.pl.readPacketList()
	if err != nil {
		return nil, err
	}

	req := &Request{
		Header:  make(map[string]string),
		Payload: &pktlineReader{pl: o.pl},
	}

	for _, pair := range requestList {
		v := strings.SplitN(pair, "=", 2)
		req.Header[v[0]] = v[1]
	}

	return req, nil
}

// WriteList writes a list of strings to the underlying pktline data stream in
// pktline format.
func (o *FilterProcessScanner) WriteList(list []string) error {
	return o.pl.writePacketList(list)
}

func (o *FilterProcessScanner) WriteStatus(status FilterProcessStatus) error {
	return o.pl.writePacketList([]string{"status=" + status.String()})
}

// isStringInSlice returns whether a given string "what" is contained in a
// slice, "s".
//
// isStringInSlice is copied from "github.com/xeipuuv/gojsonschema/utils.go"
func isStringInSlice(s []string, what string) bool {
	for i := range s {
		if s[i] == what {
			return true
		}
	}
	return false
}
