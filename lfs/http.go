package lfs

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

type transferStats struct {
	HeaderSize int
	BodySize   int
	Start      time.Time
	Stop       time.Time
}

type transfer struct {
	requestStats  *transferStats
	responseStats *transferStats
}

var (
	// TODO should use some locks
	transfers           = make(map[*http.Response]*transfer)
	transferBuckets     = make(map[string][]*http.Response)
	transfersLock       sync.Mutex
	transferBucketsLock sync.Mutex
)

func LogTransfer(key string, res *http.Response) {
	if Config.isLoggingStats {
		transferBucketsLock.Lock()
		transferBuckets[key] = append(transferBuckets[key], res)
		transferBucketsLock.Unlock()
	}
}

type HttpClient struct {
	*http.Client
}

func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
	traceHttpRequest(req)

	crc := countingRequest(req)
	if req.Body != nil {
		// Only set the body if we have a body, but create the countingRequest
		// anyway to make using zeroed stats easier.
		req.Body = crc
	}

	start := time.Now()
	res, err := c.Client.Do(req)
	if err != nil {
		return res, err
	}

	traceHttpResponse(res)

	cresp := countingResponse(res)
	res.Body = cresp

	if Config.isLoggingStats {
		reqHeaderSize := 0
		resHeaderSize := 0

		if dump, err := httputil.DumpRequest(req, false); err == nil {
			reqHeaderSize = len(dump)
		}

		if dump, err := httputil.DumpResponse(res, false); err == nil {
			resHeaderSize = len(dump)
		}

		reqstats := &transferStats{HeaderSize: reqHeaderSize, BodySize: crc.Count}

		// Response body size cannot be figured until it is read. Do not rely on a Content-Length
		// header because it may not exist or be -1 in the case of chunked responses.
		resstats := &transferStats{HeaderSize: resHeaderSize, Start: start}
		t := &transfer{requestStats: reqstats, responseStats: resstats}
		transfersLock.Lock()
		transfers[res] = t
		transfersLock.Unlock()
	}

	return res, err
}

func (c *Configuration) HttpClient() *HttpClient {
	if c.httpClient != nil {
		return c.httpClient
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	sslVerify, _ := c.GitConfig("http.sslverify")
	if sslVerify == "false" || Config.GetenvBool("GIT_SSL_NO_VERIFY", false) {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c.httpClient = &HttpClient{
		&http.Client{Transport: tr, CheckRedirect: checkRedirect},
	}

	return c.httpClient
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 3 {
		return errors.New("stopped after 3 redirects")
	}

	oldest := via[0]
	for key, _ := range oldest.Header {
		if key == "Authorization" {
			if req.URL.Scheme != oldest.URL.Scheme || req.URL.Host != oldest.URL.Host {
				continue
			}
		}
		req.Header.Set(key, oldest.Header.Get(key))
	}

	tracerx.Printf("api: redirect %s %s to %s", oldest.Method, oldest.URL, req.URL)

	return nil
}

var tracedTypes = []string{"json", "text", "xml", "html"}

func traceHttpRequest(req *http.Request) {
	tracerx.Printf("HTTP: %s %s", req.Method, req.URL.String())

	if Config.isTracingHttp == false {
		return
	}

	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(dump))
	for scanner.Scan() {
		fmt.Fprintf(os.Stderr, "> %s\n", scanner.Text())
	}
}

func traceHttpResponse(res *http.Response) {
	if res == nil {
		return
	}

	tracerx.Printf("HTTP: %d", res.StatusCode)

	if Config.isTracingHttp == false {
		return
	}

	dump, err := httputil.DumpResponse(res, false)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(dump))
	for scanner.Scan() {
		fmt.Fprintf(os.Stderr, "< %s\n", scanner.Text())
	}
}

func countingRequest(req *http.Request) *countingReadCloser {
	return &countingReadCloser{request: req, ReadCloser: req.Body}
}

func countingResponse(res *http.Response) *countingReadCloser {
	return &countingReadCloser{response: res, ReadCloser: res.Body}
}

type countingReadCloser struct {
	Count    int
	request  *http.Request
	response *http.Response
	io.ReadCloser
}

func (c *countingReadCloser) Read(b []byte) (int, error) {
	n, err := c.ReadCloser.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	}

	c.Count += n

	if Config.isTracingHttp {
		contentType := ""
		if c.response != nil { // Response, only print certain kinds of data
			contentType = strings.ToLower(strings.SplitN(c.response.Header.Get("Content-Type"), ";", 2)[0])
		} else {
			contentType = strings.ToLower(strings.SplitN(c.request.Header.Get("Content-Type"), ";", 2)[0])
		}

		for _, tracedType := range tracedTypes {
			if strings.Contains(contentType, tracedType) {
				fmt.Fprint(os.Stderr, string(b[0:n]))
			}
		}
	}

	if err == io.EOF && Config.isLoggingStats {
		// This transfer is done, we're checking it this way so we can also
		// catch transfers where the caller forgets to Close() the Body.
		if c.response != nil {
			transfersLock.Lock()
			if transfer, ok := transfers[c.response]; ok {
				transfer.responseStats.BodySize = c.Count
				transfer.responseStats.Stop = time.Now()
			}
			transfersLock.Unlock()
		}
	}
	return n, err
}

// LogHttpStats is intended to be called after all HTTP operations for the
// commmand have finished. It dumps k/v logs, one line per transfer into
// a log file with the current timestamp.
func LogHttpStats() {
	if !Config.isLoggingStats {
		return
	}

	file, err := statsLogFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error logging http stats: %s\n", err)
		return
	}

	fmt.Fprintf(file, "concurrent=%d batch=%v time=%d version=%s\n", Config.ConcurrentTransfers(), Config.BatchTransfer(), time.Now().Unix(), Version)

	for key, responses := range transferBuckets {
		for _, response := range responses {
			stats := transfers[response]
			fmt.Fprintf(file, "key=%s reqheader=%d reqbody=%d resheader=%d resbody=%d restime=%d status=%d url=%s\n",
				key,
				stats.requestStats.HeaderSize,
				stats.requestStats.BodySize,
				stats.responseStats.HeaderSize,
				stats.responseStats.BodySize,
				stats.responseStats.Stop.Sub(stats.responseStats.Start).Nanoseconds(),
				response.StatusCode,
				response.Request.URL)
		}
	}

	fmt.Fprintf(os.Stderr, "HTTP Stats logged to file %s\n", file.Name())
}

func statsLogFile() (*os.File, error) {
	logBase := filepath.Join(LocalLogDir, "http")
	if err := os.MkdirAll(logBase, 0755); err != nil {
		return nil, err
	}

	logFile := fmt.Sprintf("http-%d.log", time.Now().Unix())
	return os.Create(filepath.Join(logBase, logFile))
}
