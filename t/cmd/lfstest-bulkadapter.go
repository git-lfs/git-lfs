//go:build testtools
// +build testtools

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/tools"
)

var cfg = config.New()

// Global debug log file
var debugLog *os.File

func initDebugLog() {
	var err error
	debugLog, err = os.OpenFile("/tmp/lfstest-bulkadapter-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fall back to stderr if can't create log file
		debugLog = os.Stderr
	}
}

func writeToDebugLog(msg string) {
	if debugLog == nil {
		initDebugLog()
	}
	if !strings.HasSuffix(msg, "\n") {
		msg = msg + "\n"
	}
	debugLog.WriteString(fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), msg))
	debugLog.Sync()
}

// This test bulk adapter demonstrates the bulk transfer protocol
// It processes multiple files at once in bulks as defined by the protocol
func main() {
	initDebugLog()
	defer func() {
		if debugLog != nil && debugLog != os.Stderr {
			debugLog.Close()
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	errWriter := bufio.NewWriter(os.Stderr)
	apiClient, err := lfsapi.NewClient(cfg)
	if err != nil {
		writeToDebugLog("Error creating api client: " + err.Error())
		os.Exit(1)
	}

	// Track current bulk state
	var currentBulk *bulkState

	// Add timeout for reading from stdin
	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()

	for scanner.Scan() {
		// Reset timeout on each message
		if !timeout.Stop() {
			<-timeout.C
		}
		timeout.Reset(30 * time.Second)

		line := scanner.Text()

		// Log every incoming message for debugging
		writeToDebugLog(fmt.Sprintf("RECEIVED: %s", line))

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			writeToDebugLog(fmt.Sprintf("JSON PARSE ERROR: %v for input: %v", err, line))
			continue
		}

		writeToDebugLog(fmt.Sprintf("PARSED EVENT: %s", req.Event))

		switch req.Event {
		case "init":
			writeToDebugLog(fmt.Sprintf("Initialised test bulk adapter for %s", req.Operation))
			resp := &initResponse{}
			sendResponse(resp, writer, errWriter)

		case "bulk-header":
			writeToDebugLog(fmt.Sprintf("Starting bulk %s with %d items", req.Oid, req.Size))
			currentBulk = &bulkState{
				ID:            req.Oid,
				ItemCount:     int(req.Size),
				Items:         make([]*transferItem, 0, req.Size),
				TotalSize:     0,
				ProcessedSize: 0,
			}

		case "upload", "download":
			if currentBulk == nil {
				writeToDebugLog("Received transfer request outside of bulk context")
				continue
			}
			writeToDebugLog(fmt.Sprintf("Adding %s request for %s to bulk %s", req.Event, req.Oid, currentBulk.ID))

			item := &transferItem{
				Event:  req.Event,
				Oid:    req.Oid,
				Size:   req.Size,
				Path:   req.Path,
				Action: req.Action,
			}
			currentBulk.Items = append(currentBulk.Items, item)
			currentBulk.TotalSize += req.Size

		case "bulk-footer":
			if currentBulk == nil {
				writeToDebugLog("Received bulk footer without header")
				continue
			}
			writeToDebugLog(fmt.Sprintf("Processing bulk %s with %d items, total size %d",
				currentBulk.ID, len(currentBulk.Items), currentBulk.TotalSize))

			processBulk(apiClient, currentBulk, writer, errWriter)
			currentBulk = nil

		case "terminate":
			writeToDebugLog("Terminating test bulk adapter gracefully.")
			return
		default:
			writeToDebugLog(fmt.Sprintf("Unknown event: %s", req.Event))
		}
	}

	// Check if we exited due to timeout
	select {
	case <-timeout.C:
		writeToDebugLog("Adapter timed out waiting for input")
	default:
		// Handle scanner error or EOF
		if err := scanner.Err(); err != nil {
			writeToDebugLog(fmt.Sprintf("Scanner error: %v", err))
		} else {
			writeToDebugLog("Input stream closed, terminating adapter.")
		}
	}
}

type bulkState struct {
	ID            string
	ItemCount     int
	Items         []*transferItem
	TotalSize     int64
	ProcessedSize int64
}

type transferItem struct {
	Event  string
	Oid    string
	Size   int64
	Path   string
	Action *action
}

func processBulk(apiClient *lfsapi.Client, bulk *bulkState, writer, errWriter *bufio.Writer) {
	writeToDebugLog(fmt.Sprintf("Starting bulk processing for %s", bulk.ID))

	// Send initial progress
	sendBulkProgress(bulk.ID, 0, 0, writer, errWriter)

	// Process each item in the bulk
	for i, item := range bulk.Items {
		writeToDebugLog(fmt.Sprintf("Processing item %d/%d: %s (%s)",
			i+1, len(bulk.Items), item.Oid, item.Event))

		var err error
		var tempPath string

		if item.Event == "download" {
			tempPath, err = performBulkDownload(apiClient, item, writer, errWriter)
		} else if item.Event == "upload" {
			err = performBulkUpload(apiClient, item, writer, errWriter)
		}

		// Send item completion
		if err != nil {
			sendItemComplete(item.Oid, "", &transferError{Code: 1, Message: err.Error()}, writer, errWriter)
		} else {
			sendItemComplete(item.Oid, tempPath, nil, writer, errWriter)
		}

		// Update bulk progress
		bulk.ProcessedSize += item.Size
		progressIncrement := int(item.Size)
		sendBulkProgress(bulk.ID, bulk.ProcessedSize, progressIncrement, writer, errWriter)
	}

	// Send bulk completion
	sendBulkComplete(bulk.ID, nil, writer, errWriter)
	writeToDebugLog(fmt.Sprintf("Bulk %s completed successfully", bulk.ID))
}

func performBulkDownload(apiClient *lfsapi.Client, item *transferItem, writer, errWriter *bufio.Writer) (string, error) {
	writeToDebugLog(fmt.Sprintf("Downloading %s", item.Oid))

	// Create temp file
	dlFile, err := os.CreateTemp("", "lfs-bulk-download-*.tmp")
	if err != nil {
		return "", fmt.Errorf("Error creating temp file: %v", err)
	}
	defer dlFile.Close()

	dlfilename := dlFile.Name()

	// Make download request
	req, err := http.NewRequest("GET", item.Action.Href, nil)
	if err != nil {
		return "", fmt.Errorf("Error creating request: %v", err)
	}

	for k := range item.Action.Header {
		req.Header.Set(k, item.Action.Header[k])
	}

	res, err := apiClient.DoAPIRequestWithAuth("origin", req)
	if err != nil {
		return "", fmt.Errorf("Error downloading: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return "", fmt.Errorf("Invalid status: %d", res.StatusCode)
	}

	// Copy with progress tracking
	written, err := tools.CopyWithCallback(dlFile, res.Body, res.ContentLength, func(totalSize int64, readSoFar int64, readSinceLast int) error {
		// Progress for individual items could be sent here if needed
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("Error writing download: %v", err)
	}

	if written != item.Size {
		return "", fmt.Errorf("Downloaded size mismatch: expected %d, got %d", item.Size, written)
	}

	return dlfilename, nil
}

func performBulkUpload(apiClient *lfsapi.Client, item *transferItem, writer, errWriter *bufio.Writer) error {
	writeToDebugLog(fmt.Sprintf("Uploading %s from %s", item.Oid, item.Path))

	req, err := http.NewRequest("PUT", item.Action.Href, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	for k := range item.Action.Header {
		req.Header.Set(k, item.Action.Header[k])
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	if req.Header.Get("Transfer-Encoding") == "chunked" {
		req.TransferEncoding = []string{"chunked"}
	} else {
		req.Header.Set("Content-Length", strconv.FormatInt(item.Size, 10))
	}

	req.ContentLength = item.Size

	f, err := os.OpenFile(item.Path, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("Cannot read data from %q: %v", item.Path, err)
	}
	defer f.Close()

	// Progress callback for individual upload (simplified for bulk)
	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		return nil
	}
	req.Body = tools.NewBodyWithCallback(f, item.Size, cb)

	res, err := apiClient.DoAPIRequestWithAuth("origin", req)
	if err != nil {
		return fmt.Errorf("Error uploading: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return fmt.Errorf("Invalid status: %d", res.StatusCode)
	}

	io.Copy(io.Discard, res.Body)
	return nil
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
	writeToDebugLog(fmt.Sprintf("Sent message %v", string(b)))
	return nil
}

func sendBulkProgress(bulkId string, bytesSoFar int64, bytesSinceLast int, writer, errWriter *bufio.Writer) {
	resp := &progressResponse{"progress", bulkId, bytesSoFar, bytesSinceLast}
	err := sendResponse(resp, writer, errWriter)
	if err != nil {
		writeToDebugLog(fmt.Sprintf("Unable to send bulk progress: %v", err))
	}
}

func sendItemComplete(oid string, path string, transferErr *transferError, writer, errWriter *bufio.Writer) {
	resp := &transferResponse{"complete", oid, path, transferErr}
	err := sendResponse(resp, writer, errWriter)
	if err != nil {
		writeToDebugLog(fmt.Sprintf("Unable to send item completion: %v", err))
	}
}

func sendBulkComplete(bulkId string, transferErr *transferError, writer, errWriter *bufio.Writer) {
	resp := &bulkCompleteResponse{"bulk-complete", bulkId, transferErr}
	err := sendResponse(resp, writer, errWriter)
	if err != nil {
		writeToDebugLog(fmt.Sprintf("Unable to send bulk completion: %v", err))
	}
}

// Struct definitions
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
	Remote              string  `json:"remote"`
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
	Path  string         `json:"path,omitempty"`
	Error *transferError `json:"error,omitempty"`
}

type progressResponse struct {
	Event          string `json:"event"`
	Oid            string `json:"oid"`
	BytesSoFar     int64  `json:"bytesSoFar"`
	BytesSinceLast int    `json:"bytesSinceLast"`
}

type bulkCompleteResponse struct {
	Event string         `json:"event"`
	Oid   string         `json:"oid"`
	Error *transferError `json:"error,omitempty"`
}
