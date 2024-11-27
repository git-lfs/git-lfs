//go:build testtools
// +build testtools

package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	repoDir          string
	largeObjects     = newLfsStorage()
	server           *httptest.Server
	serverTLS        *httptest.Server
	serverClientCert *httptest.Server

	// maps OIDs to content strings. Both the LFS and Storage test servers below
	// see OIDs.
	oidHandlers map[string]string

	// These magic strings tell the test lfs server change their behavior so the
	// integration tests can check those use cases. Tests will create objects with
	// the magic strings as the contents.
	//
	//   printf "status:lfs:404" > 404.dat
	//
	contentHandlers = []string{
		"status-batch-403", "status-batch-404", "status-batch-410", "status-batch-422", "status-batch-500",
		"status-storage-403", "status-storage-404", "status-storage-410", "status-storage-422", "status-storage-500", "status-storage-503",
		"status-batch-resume-206", "batch-resume-fail-fallback", "return-expired-action", "return-expired-action-forever", "return-invalid-size",
		"object-authenticated", "storage-download-retry", "storage-upload-retry", "storage-upload-retry-later", "storage-upload-retry-later-no-header", "unknown-oid",
		"send-verify-action", "send-deprecated-links", "redirect-storage-upload", "storage-compress", "batch-hash-algo-empty", "batch-hash-algo-invalid",
		"auth-bearer", "auth-multistage",
	}

	reqCookieReposRE = regexp.MustCompile(`\A/require-cookie-`)
	dekInfoRE        = regexp.MustCompile(`DEK-Info: AES-128-CBC,([a-fA-F0-9]*)`)
)

func main() {
	repoDir = os.Getenv("LFSTEST_DIR")

	mux := http.NewServeMux()
	server = httptest.NewServer(mux)
	serverTLS = httptest.NewTLSServer(mux)
	serverClientCert = httptest.NewUnstartedServer(mux)

	//setup Client Cert server
	rootKey, rootCert := generateCARootCertificates()
	_, clientCertPEM, clientKeyPEM, clientKeyEncPEM := generateClientCertificates(rootCert, rootKey)

	certPool := x509.NewCertPool()
	certPool.AddCert(rootCert)

	serverClientCert.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverTLS.TLS.Certificates[0]},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	serverClientCert.StartTLS()

	stopch := make(chan bool)

	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		stopch <- true
	})

	mux.HandleFunc("/storage/", storageHandler)
	mux.HandleFunc("/verify", verifyHandler)
	mux.HandleFunc("/redirect307/", redirect307Handler)
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s\n", time.Now().String())
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id, ok := reqId(w)
		if !ok {
			return
		}

		if reqCookieReposRE.MatchString(r.URL.Path) {
			if skipIfNoCookie(w, r, id) {
				return
			}
		}

		if strings.Contains(r.URL.Path, "/info/lfs") {
			if !skipIfBadAuth(w, r, id) {
				lfsHandler(w, r, id)
			}

			return
		}

		debug(id, "git http-backend %s %s", r.Method, r.URL)
		gitHandler(w, r)
	})

	urlname := writeTestStateFile([]byte(server.URL), "LFSTEST_URL", "lfstest-gitserver")
	defer os.RemoveAll(urlname)

	sslurlname := writeTestStateFile([]byte(serverTLS.URL), "LFSTEST_SSL_URL", "lfstest-gitserver-ssl")
	defer os.RemoveAll(sslurlname)

	clientCertUrlname := writeTestStateFile([]byte(serverClientCert.URL), "LFSTEST_CLIENT_CERT_URL", "lfstest-gitserver-client-cert-url")
	defer os.RemoveAll(clientCertUrlname)

	block := &pem.Block{}
	block.Type = "CERTIFICATE"
	block.Bytes = serverTLS.TLS.Certificates[0].Certificate[0]
	pembytes := pem.EncodeToMemory(block)

	certname := writeTestStateFile(pembytes, "LFSTEST_CERT", "lfstest-gitserver-cert")
	defer os.RemoveAll(certname)

	cccertname := writeTestStateFile(clientCertPEM, "LFSTEST_CLIENT_CERT", "lfstest-gitserver-client-cert")
	defer os.RemoveAll(cccertname)

	ckcertname := writeTestStateFile(clientKeyPEM, "LFSTEST_CLIENT_KEY", "lfstest-gitserver-client-key")
	defer os.RemoveAll(ckcertname)

	ckecertname := writeTestStateFile(clientKeyEncPEM, "LFSTEST_CLIENT_KEY_ENCRYPTED", "lfstest-gitserver-client-key-enc")
	defer os.RemoveAll(ckecertname)

	debug("init", "server url: %s", server.URL)
	debug("init", "server tls url: %s", serverTLS.URL)
	debug("init", "server client cert url: %s", serverClientCert.URL)

	<-stopch
	server.Close()
	serverTLS.Close()
	serverClientCert.Close()
	debug("close", "git server done")
}

// writeTestStateFile writes contents to either the file referenced by the
// environment variable envVar, or defaultFilename if that's not set. Returns
// the filename that was used
func writeTestStateFile(contents []byte, envVar, defaultFilename string) string {
	f := os.Getenv(envVar)
	if len(f) == 0 {
		f = defaultFilename
	}
	file, err := os.Create(f)
	if err != nil {
		log.Fatalln(err)
	}
	file.Write(contents)
	file.Close()
	return f
}

type lfsObject struct {
	Oid           string              `json:"oid,omitempty"`
	Size          int64               `json:"size,omitempty"`
	Authenticated bool                `json:"authenticated,omitempty"`
	Actions       map[string]*lfsLink `json:"actions,omitempty"`
	Links         map[string]*lfsLink `json:"_links,omitempty"`
	Err           *lfsError           `json:"error,omitempty"`
}

type lfsLink struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header,omitempty"`
	ExpiresAt time.Time         `json:"expires_at,omitempty"`
	ExpiresIn int               `json:"expires_in,omitempty"`
}

type lfsError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

func writeLFSError(w http.ResponseWriter, code int, msg string) {
	by, err := json.Marshal(&lfsError{Message: msg})
	if err != nil {
		http.Error(w, "json encoding error: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
	w.WriteHeader(code)
	w.Write(by)
}

// handles any requests with "{name}.server.git/info/lfs" in the path
func lfsHandler(w http.ResponseWriter, r *http.Request, id string) {
	repo, err := repoFromLfsUrl(r.URL.Path)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	// Check that we're sending valid data.
	if !strings.Contains(r.Header.Get("Accept"), "application/vnd.git-lfs+json") {
		w.WriteHeader(406)
		return
	}

	debug(id, "git lfs %s %s repo: %s", r.Method, r.URL, repo)
	w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
	switch r.Method {
	case "POST":
		// Reject invalid data.
		if !strings.Contains(r.Header.Get("Content-Type"), "application/vnd.git-lfs+json") {
			w.WriteHeader(400)
			return
		}

		if strings.HasSuffix(r.URL.String(), "batch") {
			lfsBatchHandler(w, r, id, repo)
		} else {
			locksHandler(w, r, repo)
		}
	case "DELETE":
		lfsDeleteHandler(w, r, id, repo)
	case "GET":
		if strings.Contains(r.URL.String(), "/locks") {
			locksHandler(w, r, repo)
		} else {
			w.WriteHeader(404)
			w.Write([]byte("lock request"))
		}
	default:
		w.WriteHeader(405)
	}
}

func lfsUrl(repo, oid string, redirect bool) string {
	repo = url.QueryEscape(repo)
	if redirect {
		return server.URL + "/redirect307/objects/" + oid + "?r=" + repo
	}
	return server.URL + "/storage/" + oid + "?r=" + repo
}

const (
	secondsToRefillTokens = 10
	refillTokenCount      = 5
)

var (
	requestTokens   = make(map[string]int)
	retryStartTimes = make(map[string]time.Time)
	laterRetriesMu  sync.Mutex
)

// checkRateLimit tracks the various requests to the git-server. If it is the first
// request of its kind, then a times is started, that when it is finished, a certain
// number of requests become available.
func checkRateLimit(api, direction, repo, oid string) (seconds int, isWait bool) {
	laterRetriesMu.Lock()
	defer laterRetriesMu.Unlock()
	key := strings.Join([]string{direction, repo, oid}, ":")
	if requestsRemaining, ok := requestTokens[key]; !ok || requestsRemaining == 0 {
		if retryStartTimes[key] == (time.Time{}) {
			// If time is not initialized, set it to now
			retryStartTimes[key] = time.Now()
		}
		// The user is not allowed to make a request now and must wait for the required
		// time to pass.
		secsPassed := time.Since(retryStartTimes[key]).Seconds()
		if secsPassed >= float64(secondsToRefillTokens) {
			// The required time has passed.
			requestTokens[key] = refillTokenCount
			return 0, false
		}
		return secondsToRefillTokens - int(secsPassed) + 1, true
	}

	requestTokens[key]--

	// Tokens are now over, record time.
	if requestTokens[key] == 0 {
		retryStartTimes[key] = time.Now()
	}

	return 0, false
}

var (
	retries   = make(map[string]uint32)
	retriesMu sync.Mutex
)

func incrementRetriesFor(api, direction, repo, oid string, check bool) (after uint32, ok bool) {
	// fmtStr formats a string like "<api>-<direction>-[check]-<retry>",
	// i.e., "legacy-upload-check-retry", or "storage-download-retry".
	var fmtStr string
	if check {
		fmtStr = "%s-%s-check-retry"
	} else {
		fmtStr = "%s-%s-retry"
	}

	if oidHandlers[oid] != fmt.Sprintf(fmtStr, api, direction) {
		return 0, false
	}

	retriesMu.Lock()
	defer retriesMu.Unlock()

	retryKey := strings.Join([]string{direction, repo, oid}, ":")

	retries[retryKey]++
	retries := retries[retryKey]

	return retries, true
}

func lfsDeleteHandler(w http.ResponseWriter, r *http.Request, id, repo string) {
	parts := strings.Split(r.URL.Path, "/")
	oid := parts[len(parts)-1]

	largeObjects.Delete(repo, oid)
	debug(id, "DELETE:", oid)
	w.WriteHeader(200)
}

type batchReq struct {
	Transfers []string    `json:"transfers"`
	Operation string      `json:"operation"`
	Objects   []lfsObject `json:"objects"`
	Ref       *Ref        `json:"ref,omitempty"`
}

func (r *batchReq) RefName() string {
	if r.Ref == nil {
		return ""
	}
	return r.Ref.Name
}

type batchResp struct {
	Transfer      string      `json:"transfer,omitempty"`
	Objects       []lfsObject `json:"objects"`
	HashAlgorithm string      `json:"hash_algo,omitempty"`
}

func lfsBatchHandler(w http.ResponseWriter, r *http.Request, id, repo string) {
	checkingObject := r.Header.Get("X-Check-Object") == "1"
	if !checkingObject && repo == "batchunsupported" {
		w.WriteHeader(404)
		return
	}

	if !checkingObject && repo == "badbatch" {
		w.WriteHeader(203)
		return
	}

	if repo == "netrctest" {
		_, user, pass, err := extractAuth(r.Header.Get("Authorization"))
		if err != nil || (user != "netrcuser" || pass != "netrcpass") {
			w.WriteHeader(403)
			return
		}
	}

	if missingRequiredCreds(w, r, repo) {
		return
	}

	buf := &bytes.Buffer{}
	tee := io.TeeReader(r.Body, buf)
	objs := &batchReq{}
	err := json.NewDecoder(tee).Decode(objs)
	io.Copy(io.Discard, r.Body)
	r.Body.Close()

	debug(id, "REQUEST")
	debug(id, buf.String())

	if err != nil {
		log.Fatal(err)
	}

	if strings.HasSuffix(repo, "branch-required") {
		parts := strings.Split(repo, "-")
		lenParts := len(parts)
		if lenParts > 3 && "refs/heads/"+parts[lenParts-3] != objs.RefName() {
			w.WriteHeader(403)
			json.NewEncoder(w).Encode(struct {
				Message string `json:"message"`
			}{fmt.Sprintf("Expected ref %q, got %q", "refs/heads/"+parts[lenParts-3], objs.RefName())})
			return
		}
	}

	if strings.HasSuffix(repo, "batch-retry-later") {
		if timeLeft, isWaiting := checkRateLimit("batch", "", repo, ""); isWaiting {
			w.Header().Set("Retry-After", strconv.Itoa(timeLeft))
			w.WriteHeader(http.StatusTooManyRequests)

			w.Write([]byte("rate limit reached"))
			fmt.Println("Setting header to: ", strconv.Itoa(timeLeft))
			return
		}
	}

	if strings.HasSuffix(repo, "batch-retry-later-no-header") {
		if _, isWaiting := checkRateLimit("batch", "", repo, ""); isWaiting {
			w.WriteHeader(http.StatusTooManyRequests)

			w.Write([]byte("rate limit reached"))
			fmt.Println("Not setting Retry-After header")
			return
		}
	}

	res := []lfsObject{}
	testingChunked := testingChunkedTransferEncoding(r)
	testingTus := testingTusUploadInBatchReq(r)
	testingTusInterrupt := testingTusUploadInterruptedInBatchReq(r)
	testingCustomTransfer := testingCustomTransfer(r)
	var transferChoice string
	var searchForTransfer string
	hashAlgo := "sha256"
	if testingTus {
		searchForTransfer = "tus"
	} else if testingCustomTransfer {
		searchForTransfer = "testcustom"
	}
	if len(searchForTransfer) > 0 {
		for _, t := range objs.Transfers {
			if t == searchForTransfer {
				transferChoice = searchForTransfer
				break
			}

		}
	}
	for _, obj := range objs.Objects {
		handler := oidHandlers[obj.Oid]
		action := objs.Operation

		o := lfsObject{
			Size:    obj.Size,
			Actions: make(map[string]*lfsLink),
		}

		// Clobber the OID if told to do so.
		if handler == "unknown-oid" {
			o.Oid = "unknown-oid"
		} else {
			o.Oid = obj.Oid
		}

		exists := largeObjects.Has(repo, obj.Oid)
		addAction := true
		if action == "download" {
			if !exists {
				o.Err = &lfsError{Code: 404, Message: fmt.Sprintf("Object %v does not exist", obj.Oid)}
				addAction = false
			}
		} else {
			if exists {
				// not an error but don't add an action
				addAction = false
			}
		}

		if handler == "object-authenticated" {
			o.Authenticated = true
		}

		switch handler {
		case "status-batch-403":
			o.Err = &lfsError{Code: 403, Message: "welp"}
		case "status-batch-404":
			o.Err = &lfsError{Code: 404, Message: "welp"}
		case "status-batch-410":
			o.Err = &lfsError{Code: 410, Message: "welp"}
		case "status-batch-422":
			o.Err = &lfsError{Code: 422, Message: "welp"}
		case "status-batch-500":
			o.Err = &lfsError{Code: 500, Message: "welp"}
		default: // regular 200 response
			if handler == "return-invalid-size" {
				o.Size = -1
			}

			if handler == "batch-hash-algo-empty" {
				hashAlgo = ""
			} else if handler == "batch-hash-algo-invalid" {
				hashAlgo = "invalid"
			}

			if handler == "send-deprecated-links" {
				o.Links = make(map[string]*lfsLink)
			}
			if addAction {
				a := &lfsLink{
					Href:   lfsUrl(repo, obj.Oid, handler == "redirect-storage-upload"),
					Header: map[string]string{},
				}
				a = serveExpired(a, repo, handler)

				if handler == "send-deprecated-links" {
					o.Links[action] = a
				} else {
					o.Actions[action] = a
				}
			}

			if handler == "send-verify-action" {
				o.Actions["verify"] = &lfsLink{
					Href: server.URL + "/verify",
					Header: map[string]string{
						"repo": repo,
					},
				}
			}
		}

		if testingChunked && addAction {
			if handler == "send-deprecated-links" {
				o.Links[action].Header["Transfer-Encoding"] = "chunked"
			} else {
				o.Actions[action].Header["Transfer-Encoding"] = "chunked"
			}
		}
		if testingTusInterrupt && addAction {
			if handler == "send-deprecated-links" {
				o.Links[action].Header["Lfs-Tus-Interrupt"] = "true"
			} else {
				o.Actions[action].Header["Lfs-Tus-Interrupt"] = "true"
			}
		}

		res = append(res, o)
	}

	ores := batchResp{HashAlgorithm: hashAlgo, Transfer: transferChoice, Objects: res}

	by, err := json.Marshal(ores)
	if err != nil {
		log.Fatal(err)
	}

	debug(id, "RESPONSE: 200")
	debug(id, string(by))

	w.WriteHeader(200)
	w.Write(by)
}

// emu guards expiredRepos
var emu sync.Mutex

// expiredRepos is a map keyed by repository name, valuing to whether or not it
// has yet served an expired object.
var expiredRepos = map[string]bool{}

// serveExpired marks the given repo as having served an expired object, making
// it unable for that same repository to return an expired object in the future,
func serveExpired(a *lfsLink, repo, handler string) *lfsLink {
	var (
		dur = -5 * time.Minute
		at  = time.Now().Add(dur)
	)

	if handler == "return-expired-action-forever" ||
		(handler == "return-expired-action" && canServeExpired(repo)) {

		emu.Lock()
		expiredRepos[repo] = true
		emu.Unlock()

		a.ExpiresAt = at
		return a
	}

	switch repo {
	case "expired-absolute":
		a.ExpiresAt = at
	case "expired-relative":
		a.ExpiresIn = -5
	case "expired-both":
		a.ExpiresAt = at
		a.ExpiresIn = -5
	}

	return a
}

// canServeExpired returns whether or not a repository is capable of serving an
// expired object. In other words, canServeExpired returns whether or not the
// given repo has yet served an expired object.
func canServeExpired(repo string) bool {
	emu.Lock()
	defer emu.Unlock()

	return !expiredRepos[repo]
}

// Persistent state across requests
var batchResumeFailFallbackStorageAttempts = 0
var tusStorageAttempts = 0

var (
	vmu           sync.Mutex
	verifyCounts  = make(map[string]int)
	verifyRetryRe = regexp.MustCompile(`verify-fail-(\d+)-times?$`)
)

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	repo := r.Header.Get("repo")
	var payload struct {
		Oid  string `json:"oid"`
		Size int64  `json:"size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeLFSError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	var max int
	if matches := verifyRetryRe.FindStringSubmatch(repo); len(matches) < 2 {
		return
	} else {
		max, _ = strconv.Atoi(matches[1])
	}

	key := strings.Join([]string{repo, payload.Oid}, ":")

	vmu.Lock()
	verifyCounts[key] = verifyCounts[key] + 1
	count := verifyCounts[key]
	vmu.Unlock()

	if count < max {
		writeLFSError(w, http.StatusServiceUnavailable, fmt.Sprintf(
			"intentionally failing verify request %d (out of %d)", count, max,
		))
		return
	}
}

// handles any /storage/{oid} requests
func storageHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := reqId(w)
	if !ok {
		return
	}
	repo := r.URL.Query().Get("r")
	parts := strings.Split(r.URL.Path, "/")
	oid := parts[len(parts)-1]
	if missingRequiredCreds(w, r, repo) {
		return
	}

	debug(id, "storage %s %s repo: %s", r.Method, oid, repo)
	switch r.Method {
	case "PUT":
		switch oidHandlers[oid] {
		case "status-storage-403":
			w.WriteHeader(403)
			return
		case "status-storage-404":
			w.WriteHeader(404)
			return
		case "status-storage-410":
			w.WriteHeader(410)
			return
		case "status-storage-422":
			w.WriteHeader(422)
			return
		case "status-storage-500":
			w.WriteHeader(500)
			return
		case "status-storage-503":
			writeLFSError(w, 503, "LFS is temporarily unavailable")
			return
		case "object-authenticated":
			if len(r.Header.Get("Authorization")) > 0 {
				w.WriteHeader(400)
				w.Write([]byte("Should not send authentication"))
			}
			return
		case "storage-upload-retry":
			if retries, ok := incrementRetriesFor("storage", "upload", repo, oid, false); ok && retries < 3 {
				w.WriteHeader(500)
				w.Write([]byte("malformed content"))

				return
			}
		case "storage-upload-retry-later":
			if timeLeft, isWaiting := checkRateLimit("storage", "upload", repo, oid); isWaiting {
				w.Header().Set("Retry-After", strconv.Itoa(timeLeft))
				w.WriteHeader(http.StatusTooManyRequests)

				w.Write([]byte("rate limit reached"))
				fmt.Println("Setting header to: ", strconv.Itoa(timeLeft))
				return
			}
		case "storage-upload-retry-later-no-header":
			if _, isWaiting := checkRateLimit("storage", "upload", repo, oid); isWaiting {
				w.WriteHeader(http.StatusTooManyRequests)

				w.Write([]byte("rate limit reached"))
				fmt.Println("Not setting Retry-After header")
				return
			}
		case "storage-compress":
			if r.Header.Get("Accept-Encoding") != "gzip" {
				w.WriteHeader(500)
				w.Write([]byte("not encoded"))
				return
			}
		}

		if testingChunkedTransferEncoding(r) {
			valid := false
			for _, value := range r.TransferEncoding {
				if value == "chunked" {
					valid = true
					break
				}
			}
			if !valid {
				debug(id, "Chunked transfer encoding expected")
			}
		}

		hash := sha256.New()
		buf := &bytes.Buffer{}

		io.Copy(io.MultiWriter(hash, buf), r.Body)
		oid := hex.EncodeToString(hash.Sum(nil))
		if !strings.HasSuffix(r.URL.Path, "/"+oid) {
			w.WriteHeader(403)
			return
		}

		largeObjects.Set(repo, oid, buf.Bytes())

	case "GET":
		parts := strings.Split(r.URL.Path, "/")
		oid := parts[len(parts)-1]
		statusCode := 200
		byteLimit := 0
		resumeAt := int64(0)
		compress := false

		if by, ok := largeObjects.Get(repo, oid); ok {
			if len(by) == len("storage-download-retry-later") && string(by) == "storage-download-retry-later" {
				if secsToWait, wait := checkRateLimit("storage", "download", repo, oid); wait {
					statusCode = http.StatusTooManyRequests
					w.Header().Set("Retry-After", strconv.Itoa(secsToWait))
					by = []byte("rate limit reached")
					fmt.Println("Setting header to: ", strconv.Itoa(secsToWait))
				}
			} else if len(by) == len("storage-download-retry-later-no-header") && string(by) == "storage-download-retry-later-no-header" {
				if _, wait := checkRateLimit("storage", "download", repo, oid); wait {
					statusCode = http.StatusTooManyRequests
					by = []byte("rate limit reached")
					fmt.Println("Not setting Retry-After header")
				}
			} else if len(by) == len("storage-download-retry") && string(by) == "storage-download-retry" {
				if retries, ok := incrementRetriesFor("storage", "download", repo, oid, false); ok && retries < 3 {
					statusCode = 500
					by = []byte("malformed content")
				}
			} else if len(by) == len("storage-compress") && string(by) == "storage-compress" {
				if r.Header.Get("Accept-Encoding") != "gzip" {
					statusCode = 500
					by = []byte("not encoded")
				} else {
					compress = true
				}
			} else if len(by) == len("status-batch-resume-206") && string(by) == "status-batch-resume-206" {
				// Resume if header includes range, otherwise deliberately interrupt
				if rangeHdr := r.Header.Get("Range"); rangeHdr != "" {
					regex := regexp.MustCompile(`bytes=(\d+)\-.*`)
					match := regex.FindStringSubmatch(rangeHdr)
					if match != nil && len(match) > 1 {
						statusCode = 206
						resumeAt, _ = strconv.ParseInt(match[1], 10, 32)
						w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", resumeAt, len(by), resumeAt-int64(len(by))))
					}
				} else {
					byteLimit = 10
				}
			} else if len(by) == len("batch-resume-fail-fallback") && string(by) == "batch-resume-fail-fallback" {
				// Fail any Range: request even though we said we supported it
				// To make sure client can fall back
				if rangeHdr := r.Header.Get("Range"); rangeHdr != "" {
					w.WriteHeader(416)
					return
				}
				if batchResumeFailFallbackStorageAttempts == 0 {
					// Truncate output on FIRST attempt to cause resume
					// Second attempt (without range header) is fallback, complete successfully
					byteLimit = 8
					batchResumeFailFallbackStorageAttempts++
				}
			} else if string(by) == "status-batch-retry" {
				if rangeHdr := r.Header.Get("Range"); rangeHdr != "" {
					regex := regexp.MustCompile(`bytes=(\d+)\-(.*)`)
					match := regex.FindStringSubmatch(rangeHdr)
					// We have a Range header with two
					// non-empty values.
					if match != nil && len(match) > 2 && len(match[2]) != 0 {
						first, _ := strconv.ParseInt(match[1], 10, 32)
						second, _ := strconv.ParseInt(match[2], 10, 32)
						// The second part of the range
						// is smaller than the first
						// part (or the latter part of
						// the range is non-integral).
						// This is invalid; reject it.
						if second < first {
							w.WriteHeader(400)
							return
						}
						// The range is valid; we'll
						// take the branch below.
					}
					// We got a valid range header, so
					// provide a 206 Partial Content. We
					// ignore the upper bound if one was
					// provided.
					if match != nil && len(match) > 1 {
						statusCode = 206
						resumeAt, _ = strconv.ParseInt(match[1], 10, 32)
						w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", resumeAt, len(by), resumeAt-int64(len(by))))
					}
				}
			}
			var wrtr io.Writer = w
			if compress {
				w.Header().Set("Content-Encoding", "gzip")
				gz := gzip.NewWriter(w)
				defer gz.Close()

				wrtr = gz
			}
			w.WriteHeader(statusCode)
			if byteLimit > 0 {
				wrtr.Write(by[0:byteLimit])
			} else if resumeAt > 0 {
				wrtr.Write(by[resumeAt:])
			} else {
				wrtr.Write(by)
			}
			return
		}

		w.WriteHeader(404)
	case "HEAD":
		// tus.io
		if !validateTusHeaders(r, id) {
			w.WriteHeader(400)
			return
		}
		parts := strings.Split(r.URL.Path, "/")
		oid := parts[len(parts)-1]
		var offset int64
		if by, ok := largeObjects.GetIncomplete(repo, oid); ok {
			offset = int64(len(by))
		}
		w.Header().Set("Upload-Offset", strconv.FormatInt(offset, 10))
		w.WriteHeader(200)
	case "PATCH":
		// tus.io
		if !validateTusHeaders(r, id) {
			w.WriteHeader(400)
			return
		}
		parts := strings.Split(r.URL.Path, "/")
		oid := parts[len(parts)-1]

		offsetHdr := r.Header.Get("Upload-Offset")
		offset, err := strconv.ParseInt(offsetHdr, 10, 64)
		if err != nil {
			log.Fatal("Unable to parse Upload-Offset header in request: ", err)
			w.WriteHeader(400)
			return
		}
		hash := sha256.New()
		buf := &bytes.Buffer{}
		out := io.MultiWriter(hash, buf)

		if by, ok := largeObjects.GetIncomplete(repo, oid); ok {
			if offset != int64(len(by)) {
				log.Fatal(fmt.Sprintf("Incorrect offset in request, got %d expected %d", offset, len(by)))
				w.WriteHeader(400)
				return
			}
			_, err := out.Write(by)
			if err != nil {
				log.Fatal("Error reading incomplete bytes from store: ", err)
				w.WriteHeader(500)
				return
			}
			largeObjects.DeleteIncomplete(repo, oid)
			debug(id, "Resuming upload of %v at byte %d", oid, offset)
		}

		// As a test, we intentionally break the upload from byte 0 by only
		// reading some bytes the quitting & erroring, this forces a resume
		// any offset > 0 will work ok
		var copyErr error
		if r.Header.Get("Lfs-Tus-Interrupt") == "true" && offset == 0 {
			chdr := r.Header.Get("Content-Length")
			contentLen, err := strconv.ParseInt(chdr, 10, 64)
			if err != nil {
				log.Fatal(fmt.Sprintf("Invalid Content-Length %q", chdr))
				w.WriteHeader(400)
				return
			}
			truncated := contentLen / 3
			_, _ = io.CopyN(out, r.Body, truncated)
			r.Body.Close()
			copyErr = fmt.Errorf("Simulated copy error")
		} else {
			_, copyErr = io.Copy(out, r.Body)
		}
		if copyErr != nil {
			b := buf.Bytes()
			if len(b) > 0 {
				debug(id, "Incomplete upload of %v, %d bytes", oid, len(b))
				largeObjects.SetIncomplete(repo, oid, b)
			}
			w.WriteHeader(500)
		} else {
			checkoid := hex.EncodeToString(hash.Sum(nil))
			if checkoid != oid {
				log.Fatal(fmt.Sprintf("Incorrect oid after calculation, got %q expected %q", checkoid, oid))
				w.WriteHeader(403)
				return
			}

			b := buf.Bytes()
			largeObjects.Set(repo, oid, b)
			w.Header().Set("Upload-Offset", strconv.FormatInt(int64(len(b)), 10))
			w.WriteHeader(204)
		}

	default:
		w.WriteHeader(405)
	}
}

func validateTusHeaders(r *http.Request, id string) bool {
	if len(r.Header.Get("Tus-Resumable")) == 0 {
		debug(id, "Missing Tus-Resumable header in request")
		return false
	}
	return true
}

func gitHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
	}()

	cmd := exec.Command("git", "http-backend")
	cmd.Env = []string{
		fmt.Sprintf("GIT_PROJECT_ROOT=%s", repoDir),
		fmt.Sprintf("GIT_HTTP_EXPORT_ALL="),
		fmt.Sprintf("PATH_INFO=%s", r.URL.Path),
		fmt.Sprintf("QUERY_STRING=%s", r.URL.RawQuery),
		fmt.Sprintf("REQUEST_METHOD=%s", r.Method),
		fmt.Sprintf("CONTENT_TYPE=%s", r.Header.Get("Content-Type")),
	}

	if vals := r.Header.Values("Git-Protocol"); len(vals) == 1 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_PROTOCOL=%s", vals[0]))
	}

	buffer := &bytes.Buffer{}
	cmd.Stdin = r.Body
	cmd.Stdout = buffer
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	text := textproto.NewReader(bufio.NewReader(buffer))

	code, _, _ := text.ReadCodeLine(-1)

	if code != 0 {
		w.WriteHeader(code)
	}

	headers, _ := text.ReadMIMEHeader()
	head := w.Header()
	for key, values := range headers {
		for _, value := range values {
			head.Add(key, value)
		}
	}

	io.Copy(w, text.R)
}

func redirect307Handler(w http.ResponseWriter, r *http.Request) {
	id, ok := reqId(w)
	if !ok {
		return
	}

	// Send a redirect to info/lfs
	// Make it either absolute or relative depending on subpath
	parts := strings.Split(r.URL.Path, "/")
	// first element is always blank since rooted
	var redirectTo string
	if parts[2] == "rel" {
		redirectTo = "/" + strings.Join(parts[3:], "/")
	} else if parts[2] == "abs" {
		redirectTo = server.URL + "/" + strings.Join(parts[3:], "/")
	} else if parts[2] == "objects" {
		repo := r.URL.Query().Get("r")
		redirectTo = server.URL + "/storage/" + strings.Join(parts[3:], "/") + "?r=" + repo
	} else {
		debug(id, "Invalid URL for redirect: %v", r.URL)
		w.WriteHeader(404)
		return
	}
	w.Header().Set("Location", redirectTo)
	w.WriteHeader(307)
}

type User struct {
	Name string `json:"name"`
}

type Lock struct {
	Id       string    `json:"id"`
	Path     string    `json:"path"`
	Owner    User      `json:"owner"`
	LockedAt time.Time `json:"locked_at"`
}

type LockRequest struct {
	Path string `json:"path"`
	Ref  *Ref   `json:"ref,omitempty"`
}

func (r *LockRequest) RefName() string {
	if r.Ref == nil {
		return ""
	}
	return r.Ref.Name
}

type LockResponse struct {
	Lock    *Lock  `json:"lock"`
	Message string `json:"message,omitempty"`
}

type UnlockRequest struct {
	Force bool `json:"force"`
	Ref   *Ref `json:"ref,omitempty"`
}

func (r *UnlockRequest) RefName() string {
	if r.Ref == nil {
		return ""
	}
	return r.Ref.Name
}

type UnlockResponse struct {
	Lock    *Lock  `json:"lock"`
	Message string `json:"message,omitempty"`
}

type LockList struct {
	Locks      []Lock `json:"locks"`
	NextCursor string `json:"next_cursor,omitempty"`
	Message    string `json:"message,omitempty"`
}

type Ref struct {
	Name string `json:"name,omitempty"`
}

type VerifiableLockRequest struct {
	Ref    *Ref   `json:"ref,omitempty"`
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func (r *VerifiableLockRequest) RefName() string {
	if r.Ref == nil {
		return ""
	}
	return r.Ref.Name
}

type VerifiableLockList struct {
	Ours       []Lock `json:"ours"`
	Theirs     []Lock `json:"theirs"`
	NextCursor string `json:"next_cursor,omitempty"`
	Message    string `json:"message,omitempty"`
}

var (
	lmu       sync.RWMutex
	repoLocks = map[string][]Lock{}
)

func addLocks(repo string, l ...Lock) {
	lmu.Lock()
	defer lmu.Unlock()
	repoLocks[repo] = append(repoLocks[repo], l...)
	sort.Sort(LocksByCreatedAt(repoLocks[repo]))
}

func getLocks(repo string) []Lock {
	lmu.RLock()
	defer lmu.RUnlock()

	locks := repoLocks[repo]
	cp := make([]Lock, len(locks))
	for i, l := range locks {
		cp[i] = l
	}

	return cp
}

func getFilteredLocks(repo, path, cursor, limit string) ([]Lock, string, error) {
	locks := getLocks(repo)
	if cursor != "" {
		lastSeen := -1
		for i, l := range locks {
			if l.Id == cursor {
				lastSeen = i
				break
			}
		}

		if lastSeen > -1 {
			locks = locks[lastSeen:]
		} else {
			return nil, "", fmt.Errorf("cursor (%s) not found", cursor)
		}
	}

	if path != "" {
		var filtered []Lock
		for _, l := range locks {
			if l.Path == path {
				filtered = append(filtered, l)
			}
		}

		locks = filtered
	}

	if limit != "" {
		size, err := strconv.Atoi(limit)
		if err != nil {
			return nil, "", errors.New("unable to parse limit amount")
		}

		size = int(math.Min(float64(len(locks)), 3))
		if size < 0 {
			return nil, "", nil
		}

		if size+1 < len(locks) {
			return locks[:size], locks[size+1].Id, nil
		}
	}

	return locks, "", nil
}

func delLock(repo string, id string) *Lock {
	lmu.RLock()
	defer lmu.RUnlock()

	var deleted *Lock
	locks := make([]Lock, 0, len(repoLocks[repo]))
	for _, l := range repoLocks[repo] {
		if l.Id == id {
			deleted = &l
			continue
		}
		locks = append(locks, l)
	}
	repoLocks[repo] = locks
	return deleted
}

type LocksByCreatedAt []Lock

func (c LocksByCreatedAt) Len() int           { return len(c) }
func (c LocksByCreatedAt) Less(i, j int) bool { return c[i].LockedAt.Before(c[j].LockedAt) }
func (c LocksByCreatedAt) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

var (
	lockRe   = regexp.MustCompile(`/locks/?$`)
	unlockRe = regexp.MustCompile(`locks/([^/]+)/unlock\z`)
)

func locksHandler(w http.ResponseWriter, r *http.Request, repo string) {
	dec := json.NewDecoder(r.Body)
	enc := json.NewEncoder(w)

	if repo == "netrctest" {
		_, user, pass, err := extractAuth(r.Header.Get("Authorization"))
		if err != nil || (user == "netrcuser" && pass == "badpassretry") {
			writeLFSError(w, 401, "Error: Bad Auth")
			return
		}
	}

	switch r.Method {
	case "GET":
		if !lockRe.MatchString(r.URL.Path) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"unknown path: ` + r.URL.Path + `"}`))
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "could not parse form values", http.StatusInternalServerError)
			return
		}

		if strings.HasSuffix(repo, "branch-required") {
			parts := strings.Split(repo, "-")
			lenParts := len(parts)
			if lenParts > 3 && "refs/heads/"+parts[lenParts-3] != r.FormValue("refspec") {
				w.WriteHeader(403)
				enc.Encode(struct {
					Message string `json:"message"`
				}{fmt.Sprintf("Expected ref %q, got %q", "refs/heads/"+parts[lenParts-3], r.FormValue("refspec"))})
				return
			}
		}

		ll := &LockList{}
		w.Header().Set("Content-Type", "application/json")
		locks, nextCursor, err := getFilteredLocks(repo,
			r.FormValue("path"),
			r.FormValue("cursor"),
			r.FormValue("limit"))

		if err != nil {
			ll.Message = err.Error()
		} else {
			ll.Locks = locks
			ll.NextCursor = nextCursor
		}

		enc.Encode(ll)
		return
	case "POST":
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "unlock") {
			var lockId string
			if matches := unlockRe.FindStringSubmatch(r.URL.Path); len(matches) > 1 {
				lockId = matches[1]
			}

			if len(lockId) == 0 {
				enc.Encode(&UnlockResponse{Message: "Invalid lock"})
			}

			unlockRequest := &UnlockRequest{}
			if err := dec.Decode(unlockRequest); err != nil {
				enc.Encode(&UnlockResponse{Message: err.Error()})
				return
			}

			if strings.HasSuffix(repo, "branch-required") {
				parts := strings.Split(repo, "-")
				lenParts := len(parts)
				if lenParts > 3 && "refs/heads/"+parts[lenParts-3] != unlockRequest.RefName() {
					w.WriteHeader(403)
					enc.Encode(struct {
						Message string `json:"message"`
					}{fmt.Sprintf("Expected ref %q, got %q", "refs/heads/"+parts[lenParts-3], unlockRequest.RefName())})
					return
				}
			}

			if l := delLock(repo, lockId); l != nil {
				enc.Encode(&UnlockResponse{Lock: l})
			} else {
				enc.Encode(&UnlockResponse{Message: "unable to find lock"})
			}
			return
		}

		if strings.HasSuffix(r.URL.Path, "/locks/verify") {
			if strings.HasSuffix(repo, "verify-5xx") {
				w.WriteHeader(500)
				return
			}
			if strings.HasSuffix(repo, "verify-501") {
				w.WriteHeader(501)
				return
			}
			if strings.HasSuffix(repo, "verify-403") {
				w.WriteHeader(403)
				return
			}

			switch repo {
			case "pre_push_locks_verify_404":
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message":"pre_push_locks_verify_404"}`))
				return
			case "pre_push_locks_verify_410":
				w.WriteHeader(http.StatusGone)
				w.Write([]byte(`{"message":"pre_push_locks_verify_410"}`))
				return
			}

			reqBody := &VerifiableLockRequest{}
			if err := dec.Decode(reqBody); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				enc.Encode(struct {
					Message string `json:"message"`
				}{"json decode error: " + err.Error()})
				return
			}

			if strings.HasSuffix(repo, "branch-required") {
				parts := strings.Split(repo, "-")
				lenParts := len(parts)
				if lenParts > 3 && "refs/heads/"+parts[lenParts-3] != reqBody.RefName() {
					w.WriteHeader(403)
					enc.Encode(struct {
						Message string `json:"message"`
					}{fmt.Sprintf("Expected ref %q, got %q", "refs/heads/"+parts[lenParts-3], reqBody.RefName())})
					return
				}
			}

			ll := &VerifiableLockList{}
			locks, nextCursor, err := getFilteredLocks(repo, "",
				reqBody.Cursor,
				strconv.Itoa(reqBody.Limit))
			if err != nil {
				ll.Message = err.Error()
			} else {
				ll.NextCursor = nextCursor

				for _, l := range locks {
					if strings.Contains(l.Path, "theirs") {
						ll.Theirs = append(ll.Theirs, l)
					} else {
						ll.Ours = append(ll.Ours, l)
					}
				}
			}

			enc.Encode(ll)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/locks") {
			lockRequest := &LockRequest{}
			if err := dec.Decode(lockRequest); err != nil {
				enc.Encode(&LockResponse{Message: err.Error()})
			}

			if strings.HasSuffix(repo, "branch-required") {
				parts := strings.Split(repo, "-")
				lenParts := len(parts)
				if lenParts > 3 && "refs/heads/"+parts[lenParts-3] != lockRequest.RefName() {
					w.WriteHeader(403)
					enc.Encode(struct {
						Message string `json:"message"`
					}{fmt.Sprintf("Expected ref %q, got %q", "refs/heads/"+parts[lenParts-3], lockRequest.RefName())})
					return
				}
			}

			for _, l := range getLocks(repo) {
				if l.Path == lockRequest.Path {
					enc.Encode(&LockResponse{Message: "lock already created"})
					return
				}
			}

			var id [20]byte
			rand.Read(id[:])

			lock := &Lock{
				Id:       fmt.Sprintf("%x", id[:]),
				Path:     lockRequest.Path,
				Owner:    User{Name: "Git LFS Tests"},
				LockedAt: time.Now(),
			}

			addLocks(repo, *lock)

			// TODO(taylor): commit_needed case
			// TODO(taylor): err case

			enc.Encode(&LockResponse{
				Lock: lock,
			})
			return
		}
	}

	http.NotFound(w, r)
}

func missingRequiredCreds(w http.ResponseWriter, r *http.Request, repo string) bool {
	if !strings.HasPrefix(repo, "requirecreds") {
		return false
	}

	auth := r.Header.Get("Authorization")
	if len(auth) == 0 {
		writeLFSError(w, 401, "Error: Authorization Required")
		return true
	}

	_, user, pass, err := extractAuth(auth)
	if err != nil {
		writeLFSError(w, 403, err.Error())
		return true
	}

	if user != "requirecreds" || pass != "pass" {
		writeLFSError(w, 403, fmt.Sprintf("Got: '%s' => '%s' : '%s'", auth, user, pass))
		return true
	}

	return false
}

func testingChunkedTransferEncoding(r *http.Request) bool {
	return strings.HasPrefix(r.URL.String(), "/test-chunked-transfer-encoding")
}

func testingTusUploadInBatchReq(r *http.Request) bool {
	return strings.HasPrefix(r.URL.String(), "/test-tus-upload")
}
func testingTusUploadInterruptedInBatchReq(r *http.Request) bool {
	return strings.HasPrefix(r.URL.String(), "/test-tus-upload-interrupt")
}
func testingCustomTransfer(r *http.Request) bool {
	return strings.HasPrefix(r.URL.String(), "/test-custom-transfer")
}

var lfsUrlRE = regexp.MustCompile(`\A/?([^/]+)/info/lfs`)

func repoFromLfsUrl(urlpath string) (string, error) {
	matches := lfsUrlRE.FindStringSubmatch(urlpath)
	if len(matches) != 2 {
		return "", fmt.Errorf("LFS url '%s' does not match %v", urlpath, lfsUrlRE)
	}

	repo := matches[1]
	if strings.HasSuffix(repo, ".git") {
		return repo[0 : len(repo)-4], nil
	}
	return repo, nil
}

type lfsStorage struct {
	objects    map[string]map[string][]byte
	incomplete map[string]map[string][]byte
	mutex      *sync.Mutex
}

func (s *lfsStorage) Get(repo, oid string) ([]byte, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	repoObjects, ok := s.objects[repo]
	if !ok {
		return nil, ok
	}

	by, ok := repoObjects[oid]
	return by, ok
}

func (s *lfsStorage) Has(repo, oid string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	repoObjects, ok := s.objects[repo]
	if !ok {
		return false
	}

	_, ok = repoObjects[oid]
	return ok
}

func (s *lfsStorage) Set(repo, oid string, by []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	repoObjects, ok := s.objects[repo]
	if !ok {
		repoObjects = make(map[string][]byte)
		s.objects[repo] = repoObjects
	}
	repoObjects[oid] = by
}

func (s *lfsStorage) Delete(repo, oid string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	repoObjects, ok := s.objects[repo]
	if ok {
		delete(repoObjects, oid)
	}
}

func (s *lfsStorage) GetIncomplete(repo, oid string) ([]byte, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	repoObjects, ok := s.incomplete[repo]
	if !ok {
		return nil, ok
	}

	by, ok := repoObjects[oid]
	return by, ok
}

func (s *lfsStorage) SetIncomplete(repo, oid string, by []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	repoObjects, ok := s.incomplete[repo]
	if !ok {
		repoObjects = make(map[string][]byte)
		s.incomplete[repo] = repoObjects
	}
	repoObjects[oid] = by
}

func (s *lfsStorage) DeleteIncomplete(repo, oid string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	repoObjects, ok := s.incomplete[repo]
	if ok {
		delete(repoObjects, oid)
	}
}

func newLfsStorage() *lfsStorage {
	return &lfsStorage{
		objects:    make(map[string]map[string][]byte),
		incomplete: make(map[string]map[string][]byte),
		mutex:      &sync.Mutex{},
	}
}

func extractAuth(auth string) (string, string, string, error) {
	if strings.HasPrefix(auth, "Basic ") {
		decodeBy, err := base64.StdEncoding.DecodeString(auth[6:len(auth)])
		decoded := string(decodeBy)

		if err != nil {
			return "", "", "", err
		}

		parts := strings.SplitN(decoded, ":", 2)
		if len(parts) == 2 {
			return "Basic", parts[0], parts[1], nil
		}
		return "", "", "", nil
	} else if strings.HasPrefix(auth, "Bearer ") || strings.HasPrefix(auth, "Multistage ") {
		authtype, cred, _ := strings.Cut(auth, " ")
		return authtype, "", cred, nil
	}

	return "", "", "", nil
}

func skipIfNoCookie(w http.ResponseWriter, r *http.Request, id string) bool {
	cookie := r.Header.Get("Cookie")
	if strings.Contains(cookie, "secret") {
		return false
	}

	w.WriteHeader(403)
	debug(id, "No cookie received: %q", r.URL.Path)
	return true
}

func skipIfBadAuth(w http.ResponseWriter, r *http.Request, id string) bool {
	wantedAuth := "Basic realm=\"testsuite\""
	authHeader := "Lfs-Authenticate"
	if strings.HasPrefix(r.URL.Path, "/auth-bearer") {
		wantedAuth = "Bearer"
		authHeader = "Www-Authenticate"
	}

	if strings.HasPrefix(r.URL.Path, "/auth-multistage") {
		wantedAuth = "Multistage type=foo"
		authHeader = "Www-Authenticate"
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.Header().Add(authHeader, wantedAuth)
		w.WriteHeader(401)
		return true
	}

	authtype, user, cred, err := extractAuth(auth)
	if err != nil {
		w.WriteHeader(403)
		debug(id, "Error decoding auth: %s", err)
		return true
	}

	if !strings.HasPrefix(wantedAuth, authtype) {
		w.WriteHeader(403)
		debug(id, "Unwanted auth: %s (wanted %q)", authtype, wantedAuth)
		return true
	}

	switch authtype {
	case "Basic":
		switch user {
		case "user":
			if cred == "pass" {
				return false
			}
		case "netrcuser", "requirecreds":
			return false
		case "path":
			if strings.HasPrefix(r.URL.Path, "/"+cred) {
				return false
			}
			debug(id, "auth attempt against: %q", r.URL.Path)
		}
	case "Bearer":
		if cred == "token" {
			return false
		}
	case "Multistage":
		if cred == "cred1" {
			wantedAuth = "Multistage type=bar"
			w.Header().Add(authHeader, wantedAuth)
			w.WriteHeader(401)
			debug(id, "auth stage 1 succeeded: %q", auth)
			return true
		} else if cred == "cred2" {
			return false
		}
	}

	w.WriteHeader(403)
	debug(id, "Bad auth: %q", auth)
	return true
}

func init() {
	oidHandlers = make(map[string]string)
	for _, content := range contentHandlers {
		h := sha256.New()
		h.Write([]byte(content))
		oidHandlers[hex.EncodeToString(h.Sum(nil))] = content
	}
}

func debug(reqid, msg string, args ...interface{}) {
	fullargs := make([]interface{}, len(args)+1)
	fullargs[0] = reqid
	for i, a := range args {
		fullargs[i+1] = a
	}
	log.Printf("[%s] "+msg+"\n", fullargs...)
}

func reqId(w http.ResponseWriter) (string, bool) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, "error generating id: "+err.Error(), 500)
		return "", false
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), true
}

// https://ericchiang.github.io/post/go-tls/
func generateCARootCertificates() (rootKey *rsa.PrivateKey, rootCert *x509.Certificate) {

	// generate a new key-pair
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generating random key: %v", err)
	}

	rootCertTmpl, err := CertTemplate()
	if err != nil {
		log.Fatalf("creating cert template: %v", err)
	}
	// describe what the certificate will be used for
	rootCertTmpl.IsCA = true
	rootCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	rootCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	//	rootCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	rootCert, _, err = CreateCert(rootCertTmpl, rootCertTmpl, &rootKey.PublicKey, rootKey)

	return
}

func generateClientCertificates(rootCert *x509.Certificate, rootKey interface{}) (clientKey *rsa.PrivateKey, clientCertPEM []byte, clientKeyPEM []byte, clientKeyEncPEM []byte) {

	// create a key-pair for the client
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generating random key: %v", err)
	}

	// create a template for the client
	clientCertTmpl, err1 := CertTemplate()
	if err1 != nil {
		log.Fatalf("creating cert template: %v", err1)
	}
	clientCertTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	clientCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}

	// the root cert signs the cert by again providing its private key
	_, clientCertPEM, err2 := CreateCert(clientCertTmpl, rootCert, &clientKey.PublicKey, rootKey)
	if err2 != nil {
		log.Fatalf("error creating cert: %v", err2)
	}

	privKey := x509.MarshalPKCS1PrivateKey(clientKey)

	// encode and load the cert and private key for the client
	clientKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: privKey,
	})

	clientKeyEnc, err := x509.EncryptPEMBlock(bytes.NewBuffer(privKey), "RSA PRIVATE KEY", privKey, ([]byte)("pass"), x509.PEMCipherAES128)
	if err != nil {
		log.Fatalf("creating encrypted private key: %v", err)
	}
	clientKeyEncPEM = pem.EncodeToMemory(clientKeyEnc)

	// ensure salt is in uppercase hexadecimal for gnutls library v3.7.x:
	// https://github.com/gnutls/gnutls/commit/4604bbde14d2c6adb2af5315f9063ad65ab50aa6
	// https://github.com/gnutls/gnutls/blob/a0aa4780892dcc3c14cc10d823f8766ac75bcd85/lib/x509/privkey_openssl.c#L205-L206
	dekInfoIndexes := dekInfoRE.FindSubmatchIndex(clientKeyEncPEM)
	if dekInfoIndexes == nil || len(dekInfoIndexes) != 4 {
		log.Fatalf("DEK-Info header not found in encrypted private key: %s", string(clientKeyEncPEM))
	}
	for i := dekInfoIndexes[2]; i < dekInfoIndexes[3]; i++ {
		c := clientKeyEncPEM[i]
		if c >= 'a' && c <= 'f' {
			clientKeyEncPEM[i] = byte(unicode.ToUpper(rune(c)))
		}
	}

	return
}

// helper function to create a cert template with a serial number and other required fields
func CertTemplate() (*x509.Certificate, error) {
	// generate a random serial number (a real cert authority would have some logic behind this)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.New("failed to generate serial number: " + err.Error())
	}

	tmpl := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"Yhat, Inc."}},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour), // valid for an hour
		BasicConstraintsValid: true,
	}
	return &tmpl, nil
}

func CreateCert(template, parent *x509.Certificate, pub interface{}, parentPriv interface{}) (
	cert *x509.Certificate, certPEM []byte, err error) {

	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, parentPriv)
	if err != nil {
		return
	}
	// parse the resulting certificate so we can use it again
	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return
	}
	// PEM encode the certificate (this is a standard TLS encoding)
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM = pem.EncodeToMemory(&b)
	return
}
