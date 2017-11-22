// +build testtools

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/tools"
)

var cfg = config.New()

// This test custom adapter just acts as a bridge for uploads/downloads
// in order to demonstrate & test the custom transfer adapter protocols
// All we actually do is relay the requests back to the normal storage URLs
// of our test server for simplicity, but this proves the principle
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	errWriter := bufio.NewWriter(os.Stderr)
	apiClient, err := lfsapi.NewClient(cfg)
	if err != nil {
		writeToStderr("Error creating api client: "+err.Error(), errWriter)
		os.Exit(1)
	}

	for scanner.Scan() {
		line := scanner.Text()
		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			writeToStderr(fmt.Sprintf("Unable to parse request: %v\n", line), errWriter)
			continue
		}

		switch req.Event {
		case "init":
			writeToStderr(fmt.Sprintf("Initialised test custom adapter for %s\n", req.Operation), errWriter)
			resp := &initResponse{}
			sendResponse(resp, writer, errWriter)
		case "download":
			writeToStderr(fmt.Sprintf("Received download request for %s\n", req.Oid), errWriter)
			performDownload(apiClient, req.Oid, req.Size, req.Action, writer, errWriter)
		case "upload":
			writeToStderr(fmt.Sprintf("Received upload request for %s\n", req.Oid), errWriter)
			performUpload(apiClient, req.Oid, req.Size, req.Action, req.Path, writer, errWriter)
		case "terminate":
			writeToStderr("Terminating test custom adapter gracefully.\n", errWriter)
			break
		}
	}

}

func writeToStderr(msg string, errWriter *bufio.Writer) {
	if !strings.HasSuffix(msg, "\n") {
		msg = msg + "\n"
	}
	errWriter.WriteString(msg)
	errWriter.Flush()
}

func sendResponse(r interface{}, writer, errWriter *bufio.Writer) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	// Line oriented JSON
	b = append(b, '\n')
	_, err = writer.Write(b)
	if err != nil {
		return err
	}
	writer.Flush()
	writeToStderr(fmt.Sprintf("Sent message %v", string(b)), errWriter)
	return nil
}

func sendTransferError(oid string, code int, message string, writer, errWriter *bufio.Writer) {
	resp := &transferResponse{"complete", oid, "", &transferError{code, message}}
	err := sendResponse(resp, writer, errWriter)
	if err != nil {
		writeToStderr(fmt.Sprintf("Unable to send transfer error: %v\n", err), errWriter)
	}
}

func sendProgress(oid string, bytesSoFar int64, bytesSinceLast int, writer, errWriter *bufio.Writer) {
	resp := &progressResponse{"progress", oid, bytesSoFar, bytesSinceLast}
	err := sendResponse(resp, writer, errWriter)
	if err != nil {
		writeToStderr(fmt.Sprintf("Unable to send progress update: %v\n", err), errWriter)
	}
}

func performDownload(apiClient *lfsapi.Client, oid string, size int64, a *action, writer, errWriter *bufio.Writer) {
	// We just use the URLs we're given, so we're just a proxy for the direct method
	// but this is enough to test intermediate custom adapters
	req, err := http.NewRequest("GET", a.Href, nil)
	if err != nil {
		sendTransferError(oid, 2, err.Error(), writer, errWriter)
		return
	}

	for k := range a.Header {
		req.Header.Set(k, a.Header[k])
	}

	res, err := apiClient.DoWithAuth("origin", req)
	if err != nil {
		sendTransferError(oid, res.StatusCode, err.Error(), writer, errWriter)
		return
	}
	defer res.Body.Close()

	dlFile, err := ioutil.TempFile("", "lfscustomdl")
	if err != nil {
		sendTransferError(oid, 3, err.Error(), writer, errWriter)
		return
	}
	defer dlFile.Close()
	dlfilename := dlFile.Name()
	// Turn callback into progress messages
	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		sendProgress(oid, readSoFar, readSinceLast, writer, errWriter)
		return nil
	}
	_, err = tools.CopyWithCallback(dlFile, res.Body, res.ContentLength, cb)
	if err != nil {
		sendTransferError(oid, 4, fmt.Sprintf("cannot write data to tempfile %q: %v", dlfilename, err), writer, errWriter)
		os.Remove(dlfilename)
		return
	}
	if err := dlFile.Close(); err != nil {
		sendTransferError(oid, 5, fmt.Sprintf("can't close tempfile %q: %v", dlfilename, err), writer, errWriter)
		os.Remove(dlfilename)
		return
	}

	// completed
	complete := &transferResponse{"complete", oid, dlfilename, nil}
	err = sendResponse(complete, writer, errWriter)
	if err != nil {
		writeToStderr(fmt.Sprintf("Unable to send completion message: %v\n", err), errWriter)
	}
}

func performUpload(apiClient *lfsapi.Client, oid string, size int64, a *action, fromPath string, writer, errWriter *bufio.Writer) {
	// We just use the URLs we're given, so we're just a proxy for the direct method
	// but this is enough to test intermediate custom adapters
	req, err := http.NewRequest("PUT", a.Href, nil)
	if err != nil {
		sendTransferError(oid, 2, err.Error(), writer, errWriter)
		return
	}

	for k := range a.Header {
		req.Header.Set(k, a.Header[k])
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	if req.Header.Get("Transfer-Encoding") == "chunked" {
		req.TransferEncoding = []string{"chunked"}
	} else {
		req.Header.Set("Content-Length", strconv.FormatInt(size, 10))
	}

	req.ContentLength = size

	f, err := os.OpenFile(fromPath, os.O_RDONLY, 0644)
	if err != nil {
		sendTransferError(oid, 3, fmt.Sprintf("Cannot read data from %q: %v", fromPath, err), writer, errWriter)
		return
	}
	defer f.Close()

	// Turn callback into progress messages
	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		sendProgress(oid, readSoFar, readSinceLast, writer, errWriter)
		return nil
	}
	req.Body = tools.NewBodyWithCallback(f, size, cb)

	res, err := apiClient.DoWithAuth("origin", req)
	if err != nil {
		sendTransferError(oid, res.StatusCode, fmt.Sprintf("Error uploading data for %s: %v", oid, err), writer, errWriter)
		return
	}

	if res.StatusCode > 299 {
		msg := fmt.Sprintf("Invalid status for %s %s: %d",
			req.Method, strings.SplitN(req.URL.String(), "?", 2)[0], res.StatusCode)
		sendTransferError(oid, res.StatusCode, msg, writer, errWriter)
		return
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	// completed
	complete := &transferResponse{"complete", oid, "", nil}
	err = sendResponse(complete, writer, errWriter)
	if err != nil {
		writeToStderr(fmt.Sprintf("Unable to send completion message: %v\n", err), errWriter)
	}

}

// Structs reimplemented so closer to a real external implementation
type header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type action struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header,omitempty"`
	ExpiresAt time.Time         `json:"expires_at,omitempty"`
}
type transferError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Combined request struct which can accept anything
type request struct {
	Event               string  `json:"event"`
	Operation           string  `json:"operation"`
	Concurrent          bool    `json:"concurrent"`
	ConcurrentTransfers int     `json:"concurrenttransfers"`
	Oid                 string  `json:"oid"`
	Size                int64   `json:"size"`
	Path                string  `json:"path"`
	Action              *action `json:"action"`
}

type initResponse struct {
	Error *transferError `json:"error,omitempty"`
}
type transferResponse struct {
	Event string         `json:"event"`
	Oid   string         `json:"oid"`
	Path  string         `json:"path,omitempty"` // always blank for upload
	Error *transferError `json:"error,omitempty"`
}
type progressResponse struct {
	Event          string `json:"event"`
	Oid            string `json:"oid"`
	BytesSoFar     int64  `json:"bytesSoFar"`
	BytesSinceLast int    `json:"bytesSinceLast"`
}
