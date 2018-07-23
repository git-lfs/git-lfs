// +build testtools

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/tools"
)

var backupDir string

// This test custom adapter just copies the files to a folder.
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	errWriter := bufio.NewWriter(os.Stderr)
	backupDir = os.Getenv("TEST_STANDALONE_BACKUP_PATH")
	if backupDir == "" {
		writeToStderr("TEST_STANDALONE_BACKUP_PATH backup dir not set", errWriter)
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
			performDownload(req.Oid, req.Size, writer, errWriter)
		case "upload":
			writeToStderr(fmt.Sprintf("Received upload request for %s\n", req.Oid), errWriter)
			performUpload(req.Oid, req.Size, req.Path, writer, errWriter)
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

func performCopy(oid, src, dst string, size int64, writer, errWriter *bufio.Writer) error {
	writeToStderr(fmt.Sprintf("Copying %s to %s\n", src, dst), errWriter)
	srcFile, err := os.OpenFile(src, os.O_RDONLY, 0644)
	if err != nil {
		sendTransferError(oid, 10, err.Error(), writer, errWriter)
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		sendTransferError(oid, 11, err.Error(), writer, errWriter)
		return err
	}
	defer dstFile.Close()

	// Turn callback into progress messages
	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		sendProgress(oid, readSoFar, readSinceLast, writer, errWriter)
		return nil
	}
	_, err = tools.CopyWithCallback(dstFile, srcFile, size, cb)
	if err != nil {
		sendTransferError(oid, 4, fmt.Sprintf("cannot write data to dst %q: %v", dst, err), writer, errWriter)
		os.Remove(dst)
		return err
	}
	if err := dstFile.Close(); err != nil {
		sendTransferError(oid, 5, fmt.Sprintf("can't close dst %q: %v", dst, err), writer, errWriter)
		os.Remove(dst)
		return err
	}
	return nil
}

func performDownload(oid string, size int64, writer, errWriter *bufio.Writer) {
	dlFile, err := ioutil.TempFile("", "lfscustomdl")
	if err != nil {
		sendTransferError(oid, 1, err.Error(), writer, errWriter)
		return
	}
	if err = dlFile.Close(); err != nil {
		sendTransferError(oid, 2, err.Error(), writer, errWriter)
		return
	}
	dlfilename := dlFile.Name()
	backupPath := filepath.Join(backupDir, oid)
	if err = performCopy(oid, backupPath, dlfilename, size, writer, errWriter); err != nil {
		return
	}

	// completed
	complete := &transferResponse{"complete", oid, dlfilename, nil}
	if err := sendResponse(complete, writer, errWriter); err != nil {
		writeToStderr(fmt.Sprintf("Unable to send completion message: %v\n", err), errWriter)
	}
}

func performUpload(oid string, size int64, fromPath string, writer, errWriter *bufio.Writer) {
	backupPath := filepath.Join(backupDir, oid)
	if err := performCopy(oid, fromPath, backupPath, size, writer, errWriter); err != nil {
		return
	}

	// completed
	complete := &transferResponse{"complete", oid, "", nil}
	if err := sendResponse(complete, writer, errWriter); err != nil {
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
