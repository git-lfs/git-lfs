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
	"net/url"
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

// HttpClient returns a new HttpClient for the given host (which may be "host:port")
func (c *Configuration) HttpClient(host string) *HttpClient {
	c.httpClientsMutex.Lock()
	defer c.httpClientsMutex.Unlock()

	if c.httpClients == nil {
		c.httpClients = make(map[string]*HttpClient)
	}
	if client, ok := c.httpClients[host]; ok {
		return client
	}

	dialtime := c.GitConfigInt("lfs.dialtimeout", 30)
	keepalivetime := c.GitConfigInt("lfs.keepalive", 1800) // 30 minutes
	tlstime := c.GitConfigInt("lfs.tlstimeout", 30)

	tr := &http.Transport{
		Proxy: proxyFromGitConfigOrEnvironment,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(dialtime) * time.Second,
			KeepAlive: time.Duration(keepalivetime) * time.Second,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(tlstime) * time.Second,
		MaxIdleConnsPerHost: c.ConcurrentTransfers(),
	}

	tr.TLSClientConfig = &tls.Config{}
	if isCertVerificationDisabledForHost(host) {
		tr.TLSClientConfig.InsecureSkipVerify = true
	} else {
		tr.TLSClientConfig.RootCAs = getRootCAsForHost(host)
	}

	client := &HttpClient{
		&http.Client{Transport: tr, CheckRedirect: checkRedirect},
	}
	c.httpClients[host] = client

	return client
}

func proxyFromGitConfigOrEnvironment(req *http.Request) (*url.URL, error) {
	proxyURL, err := http.ProxyFromEnvironment(req)

	if proxyURL != nil {
		return proxyURL, err
	}

	if proxy, ok := Config.GitConfig("http.proxy"); ok {
		proxyURL, err := url.Parse(proxy)
		return proxyURL, err
	}

	return proxyURL, err
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

	oldestUrl := strings.SplitN(oldest.URL.String(), "?", 2)[0]
	newUrl := strings.SplitN(req.URL.String(), "?", 2)[0]
	tracerx.Printf("api: redirect %s %s to %s", oldest.Method, oldestUrl, newUrl)

	return nil
}

var tracedTypes = []string{"json", "text", "xml", "html"}

func traceHttpRequest(req *http.Request) {
	tracerx.Printf("HTTP: %s", traceHttpReq(req))

	if Config.isTracingHttp == false {
		return
	}

	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		return
	}

	traceHttpDump(">", dump)
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

	if isTraceableContent(res.Header) {
		fmt.Fprintf(os.Stderr, "\n\n")
	} else {
		fmt.Fprintf(os.Stderr, "\n")
	}

	traceHttpDump("<", dump)
}

func traceHttpDump(direction string, dump []byte) {
	scanner := bufio.NewScanner(bytes.NewBuffer(dump))

	for scanner.Scan() {
		line := scanner.Text()
		if !Config.isDebuggingHttp && strings.HasPrefix(strings.ToLower(line), "authorization: basic") {
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

func countingRequest(req *http.Request) *countingReadCloser {
	return &countingReadCloser{
		request:         req,
		ReadCloser:      req.Body,
		isTraceableType: isTraceableContent(req.Header),
		useGitTrace:     false,
	}
}

func countingResponse(res *http.Response) *countingReadCloser {
	return &countingReadCloser{
		response:        res,
		ReadCloser:      res.Body,
		isTraceableType: isTraceableContent(res.Header),
		useGitTrace:     true,
	}
}

type countingReadCloser struct {
	Count           int
	request         *http.Request
	response        *http.Response
	isTraceableType bool
	useGitTrace     bool
	io.ReadCloser
}

func (c *countingReadCloser) Read(b []byte) (int, error) {
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

		if Config.isTracingHttp {
			fmt.Fprint(os.Stderr, chunk)
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
