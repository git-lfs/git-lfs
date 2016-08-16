// Package httputil provides additional helper functions for http services
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package httputil

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

	"github.com/github/git-lfs/config"
	"github.com/rubyist/tracerx"
)

type httpTransferStats struct {
	HeaderSize int
	BodySize   int
	Start      time.Time
	Stop       time.Time
}

type httpTransfer struct {
	requestStats  *httpTransferStats
	responseStats *httpTransferStats
}

var (
	// TODO should use some locks
	httpTransfers           = make(map[*http.Response]*httpTransfer)
	httpTransferBuckets     = make(map[string][]*http.Response)
	httpTransfersLock       sync.Mutex
	httpTransferBucketsLock sync.Mutex
	httpClients             map[string]*HttpClient
	httpClientsMutex        sync.Mutex
	UserAgent               string
)

func LogTransfer(cfg *config.Configuration, key string, res *http.Response) {
	if cfg.IsLoggingStats {
		httpTransferBucketsLock.Lock()
		httpTransferBuckets[key] = append(httpTransferBuckets[key], res)
		httpTransferBucketsLock.Unlock()
	}
}

type HttpClient struct {
	Config *config.Configuration
	*http.Client
}

func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
	traceHttpRequest(c.Config, req)

	crc := countingRequest(c.Config, req)
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

	traceHttpResponse(c.Config, res)

	cresp := countingResponse(c.Config, res)
	res.Body = cresp

	if c.Config.IsLoggingStats {
		reqHeaderSize := 0
		resHeaderSize := 0

		if dump, err := httputil.DumpRequest(req, false); err == nil {
			reqHeaderSize = len(dump)
		}

		if dump, err := httputil.DumpResponse(res, false); err == nil {
			resHeaderSize = len(dump)
		}

		reqstats := &httpTransferStats{HeaderSize: reqHeaderSize, BodySize: crc.Count}

		// Response body size cannot be figured until it is read. Do not rely on a Content-Length
		// header because it may not exist or be -1 in the case of chunked responses.
		resstats := &httpTransferStats{HeaderSize: resHeaderSize, Start: start}
		t := &httpTransfer{requestStats: reqstats, responseStats: resstats}
		httpTransfersLock.Lock()
		httpTransfers[res] = t
		httpTransfersLock.Unlock()
	}

	return res, err
}

// NewHttpClient returns a new HttpClient for the given host (which may be "host:port")
func NewHttpClient(c *config.Configuration, host string) *HttpClient {
	httpClientsMutex.Lock()
	defer httpClientsMutex.Unlock()

	if httpClients == nil {
		httpClients = make(map[string]*HttpClient)
	}
	if client, ok := httpClients[host]; ok {
		return client
	}

	dialtime := c.Git.Int("lfs.dialtimeout", 30)
	keepalivetime := c.Git.Int("lfs.keepalive", 1800) // 30 minutes
	tlstime := c.Git.Int("lfs.tlstimeout", 30)

	tr := &http.Transport{
		Proxy: ProxyFromGitConfigOrEnvironment(c),
		Dial: (&net.Dialer{
			Timeout:   time.Duration(dialtime) * time.Second,
			KeepAlive: time.Duration(keepalivetime) * time.Second,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(tlstime) * time.Second,
		MaxIdleConnsPerHost: c.ConcurrentTransfers(),
	}

	tr.TLSClientConfig = &tls.Config{}
	if isCertVerificationDisabledForHost(c, host) {
		tr.TLSClientConfig.InsecureSkipVerify = true
	} else {
		tr.TLSClientConfig.RootCAs = getRootCAsForHost(c, host)
	}

	client := &HttpClient{
		Config: c,
		Client: &http.Client{Transport: tr, CheckRedirect: CheckRedirect},
	}
	httpClients[host] = client

	return client
}

func CheckRedirect(req *http.Request, via []*http.Request) error {
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

	oldestUrl := strings.SplitN(oldest.URL.String(), "?", 2)[0]
	newUrl := strings.SplitN(req.URL.String(), "?", 2)[0]
	tracerx.Printf("api: redirect %s %s to %s", oldest.Method, oldestUrl, newUrl)

	return nil
}

var tracedTypes = []string{"json", "text", "xml", "html"}

func traceHttpRequest(cfg *config.Configuration, req *http.Request) {
	tracerx.Printf("HTTP: %s", TraceHttpReq(req))

	if cfg.IsTracingHttp == false {
		return
	}

	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		return
	}

	traceHttpDump(cfg, ">", dump)
}

func traceHttpResponse(cfg *config.Configuration, res *http.Response) {
	if res == nil {
		return
	}

	tracerx.Printf("HTTP: %d", res.StatusCode)

	if cfg.IsTracingHttp == false {
		return
	}

	dump, err := httputil.DumpResponse(res, false)
	if err != nil {
		return
	}

	if isTraceableContent(res.Header) {
		fmt.Fprintf(os.Stderr, "\n\n")
	} else {
		fmt.Fprintf(os.Stderr, "\n")
	}

	traceHttpDump(cfg, "<", dump)
}

func traceHttpDump(cfg *config.Configuration, direction string, dump []byte) {
	scanner := bufio.NewScanner(bytes.NewBuffer(dump))

	for scanner.Scan() {
		line := scanner.Text()
		if !cfg.IsDebuggingHttp && strings.HasPrefix(strings.ToLower(line), "authorization: basic") {
			fmt.Fprintf(os.Stderr, "%s Authorization: Basic * * * * *\n", direction)
		} else {
			fmt.Fprintf(os.Stderr, "%s %s\n", direction, line)
		}
	}
}

func isTraceableContent(h http.Header) bool {
	ctype := strings.ToLower(strings.SplitN(h.Get("Content-Type"), ";", 2)[0])
	for _, tracedType := range tracedTypes {
		if strings.Contains(ctype, tracedType) {
			return true
		}
	}
	return false
}

func countingRequest(cfg *config.Configuration, req *http.Request) *CountingReadCloser {
	return &CountingReadCloser{
		request:         req,
		cfg:             cfg,
		ReadCloser:      req.Body,
		isTraceableType: isTraceableContent(req.Header),
		useGitTrace:     false,
	}
}

func countingResponse(cfg *config.Configuration, res *http.Response) *CountingReadCloser {
	return &CountingReadCloser{
		response:        res,
		cfg:             cfg,
		ReadCloser:      res.Body,
		isTraceableType: isTraceableContent(res.Header),
		useGitTrace:     true,
	}
}

type CountingReadCloser struct {
	Count           int
	request         *http.Request
	response        *http.Response
	cfg             *config.Configuration
	isTraceableType bool
	useGitTrace     bool
	io.ReadCloser
}

func (c *CountingReadCloser) Read(b []byte) (int, error) {
	n, err := c.ReadCloser.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	}

	c.Count += n

	if c.isTraceableType {
		chunk := string(b[0:n])
		if c.useGitTrace {
			tracerx.Printf("HTTP: %s", chunk)
		}

		if c.cfg.IsTracingHttp {
			fmt.Fprint(os.Stderr, chunk)
		}
	}

	if err == io.EOF && c.cfg.IsLoggingStats {
		// This httpTransfer is done, we're checking it this way so we can also
		// catch httpTransfers where the caller forgets to Close() the Body.
		if c.response != nil {
			httpTransfersLock.Lock()
			if httpTransfer, ok := httpTransfers[c.response]; ok {
				httpTransfer.responseStats.BodySize = c.Count
				httpTransfer.responseStats.Stop = time.Now()
			}
			httpTransfersLock.Unlock()
		}
	}
	return n, err
}

// LogHttpStats is intended to be called after all HTTP operations for the
// commmand have finished. It dumps k/v logs, one line per httpTransfer into
// a log file with the current timestamp.
func LogHttpStats(cfg *config.Configuration) {
	if !cfg.IsLoggingStats {
		return
	}

	file, err := statsLogFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error logging http stats: %s\n", err)
		return
	}

	fmt.Fprintf(file, "concurrent=%d batch=%v time=%d version=%s\n", cfg.ConcurrentTransfers(), cfg.BatchTransfer(), time.Now().Unix(), config.Version)

	for key, responses := range httpTransferBuckets {
		for _, response := range responses {
			stats := httpTransfers[response]
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
	logBase := filepath.Join(config.LocalLogDir, "http")
	if err := os.MkdirAll(logBase, 0755); err != nil {
		return nil, err
	}

	logFile := fmt.Sprintf("http-%d.log", time.Now().Unix())
	return os.Create(filepath.Join(logBase, logFile))
}

func TraceHttpReq(req *http.Request) string {
	return fmt.Sprintf("%s %s", req.Method, strings.SplitN(req.URL.String(), "?", 2)[0])
}

func init() {
	UserAgent = config.VersionDesc
}
