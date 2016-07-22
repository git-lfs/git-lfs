package httputil

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/github/git-lfs/auth"
	"github.com/github/git-lfs/config"
)

func ntlmClientSession(c *config.Configuration, creds auth.Creds) (ntlm.ClientSession, error) {
	if c.NtlmSession != nil {
		return c.NtlmSession, nil
	}
	splits := strings.Split(creds["username"], "\\")

	if len(splits) != 2 {
		errorMessage := fmt.Sprintf("Your user name must be of the form DOMAIN\\user. It is currently %s", creds["username"])
		return nil, errors.New(errorMessage)
	}

	session, err := ntlm.CreateClientSession(ntlm.Version2, ntlm.ConnectionOrientedMode)

	if err != nil {
		return nil, err
	}

	session.SetUserInfo(splits[1], creds["password"], strings.ToUpper(splits[0]))
	c.NtlmSession = session
	return session, nil
}

func doNTLMRequest(cfg *config.Configuration, request *http.Request, retry bool) (*http.Response, error) {
	handReq, err := cloneRequest(request)
	if err != nil {
		return nil, err
	}

	res, err := NewHttpClient(cfg, handReq.Host).Do(handReq)
	if err != nil && res == nil {
		return nil, err
	}

	//If the status is 401 then we need to re-authenticate, otherwise it was successful
	if res.StatusCode == 401 {
		creds, err := auth.GetCreds(cfg, request)
		if err != nil {
			return nil, err
		}

		negotiateReq, err := cloneRequest(request)
		if err != nil {
			return nil, err
		}

		challengeMessage, err := negotiate(cfg, negotiateReq, ntlmNegotiateMessage)
		if err != nil {
			return nil, err
		}

		challengeReq, err := cloneRequest(request)
		if err != nil {
			return nil, err
		}

		res, err := challenge(cfg, challengeReq, challengeMessage, creds)
		if err != nil {
			return nil, err
		}

		//If the status is 401 then we need to re-authenticate
		if res.StatusCode == 401 && retry == true {
			return doNTLMRequest(cfg, challengeReq, false)
		}

		auth.SaveCredentials(cfg, creds, res)

		return res, nil
	}
	return res, nil
}

func negotiate(cfg *config.Configuration, request *http.Request, message string) ([]byte, error) {
	request.Header.Add("Authorization", message)
	res, err := NewHttpClient(cfg, request.Host).Do(request)

	if res == nil && err != nil {
		return nil, err
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	ret, err := parseChallengeResponse(res)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func challenge(cfg *config.Configuration, request *http.Request, challengeBytes []byte, creds auth.Creds) (*http.Response, error) {
	challenge, err := ntlm.ParseChallengeMessage(challengeBytes)
	if err != nil {
		return nil, err
	}

	session, err := ntlmClientSession(cfg, creds)
	if err != nil {
		return nil, err
	}

	session.ProcessChallengeMessage(challenge)
	authenticate, err := session.GenerateAuthenticateMessage()
	if err != nil {
		return nil, err
	}

	authMsg := base64.StdEncoding.EncodeToString(authenticate.Bytes())
	request.Header.Add("Authorization", "NTLM "+authMsg)
	return NewHttpClient(cfg, request.Host).Do(request)
}

func parseChallengeResponse(response *http.Response) ([]byte, error) {
	header := response.Header.Get("Www-Authenticate")
	if len(header) < 6 {
		return nil, fmt.Errorf("Invalid NTLM challenge response: %q", header)
	}

	//parse out the "NTLM " at the beginning of the response
	challenge := header[5:]
	val, err := base64.StdEncoding.DecodeString(challenge)

	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func cloneRequest(request *http.Request) (*http.Request, error) {
	cloneReqBody, err := cloneRequestBody(request)
	if err != nil {
		return nil, err
	}

	clonedReq, err := http.NewRequest(request.Method, request.URL.String(), cloneReqBody)
	if err != nil {
		return nil, err
	}

	for k, _ := range request.Header {
		clonedReq.Header.Add(k, request.Header.Get(k))
	}

	clonedReq.TransferEncoding = request.TransferEncoding
	clonedReq.ContentLength = request.ContentLength

	return clonedReq, nil
}

func cloneRequestBody(req *http.Request) (io.ReadCloser, error) {
	if req.Body == nil {
		return nil, nil
	}

	var cb *cloneableBody
	var err error
	isCloneableBody := true

	// check to see if the request body is already a cloneableBody
	body := req.Body
	if existingCb, ok := body.(*cloneableBody); ok {
		isCloneableBody = false
		cb, err = existingCb.CloneBody()
	} else {
		cb, err = newCloneableBody(req.Body, 0)
	}

	if err != nil {
		return nil, err
	}

	if isCloneableBody {
		cb2, err := cb.CloneBody()
		if err != nil {
			return nil, err
		}

		req.Body = cb2
	}

	return cb, nil
}

type cloneableBody struct {
	bytes  []byte    // in-memory buffer of body
	file   *os.File  // file buffer of in-memory overflow
	reader io.Reader // internal reader for Read()
	closed bool      // tracks whether body is closed
	dup    *dupTracker
}

func newCloneableBody(r io.Reader, limit int64) (*cloneableBody, error) {
	if limit < 1 {
		limit = 1048576 // default
	}

	b := &cloneableBody{}
	buf := &bytes.Buffer{}
	w, err := io.CopyN(buf, r, limit)
	if err != nil && err != io.EOF {
		return nil, err
	}

	b.bytes = buf.Bytes()
	byReader := bytes.NewBuffer(b.bytes)

	if w >= limit {
		tmp, err := ioutil.TempFile("", "git-lfs-clone-reader")
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(tmp, r)
		tmp.Close()
		if err != nil {
			os.RemoveAll(tmp.Name())
			return nil, err
		}

		f, err := os.Open(tmp.Name())
		if err != nil {
			os.RemoveAll(tmp.Name())
			return nil, err
		}

		dups := int32(0)
		b.dup = &dupTracker{name: f.Name(), dups: &dups}
		b.file = f
		b.reader = io.MultiReader(byReader, b.file)
	} else {
		// no file, so set the reader to just the in-memory buffer
		b.reader = byReader
	}

	return b, nil
}

func (b *cloneableBody) Read(p []byte) (int, error) {
	if b.closed {
		return 0, io.EOF
	}
	return b.reader.Read(p)
}

func (b *cloneableBody) Close() error {
	if !b.closed {
		b.closed = true
		if b.file == nil {
			return nil
		}

		b.file.Close()
		b.dup.Rm()
	}
	return nil
}

func (b *cloneableBody) CloneBody() (*cloneableBody, error) {
	if b.closed {
		return &cloneableBody{closed: true}, nil
	}

	b2 := &cloneableBody{bytes: b.bytes}

	if b.file == nil {
		b2.reader = bytes.NewBuffer(b.bytes)
	} else {
		f, err := os.Open(b.file.Name())
		if err != nil {
			return nil, err
		}
		b2.file = f
		b2.reader = io.MultiReader(bytes.NewBuffer(b.bytes), b2.file)
		b2.dup = b.dup
		b.dup.Add()
	}

	return b2, nil
}

type dupTracker struct {
	name string
	dups *int32
}

func (t *dupTracker) Add() {
	atomic.AddInt32(t.dups, 1)
}

func (t *dupTracker) Rm() {
	newval := atomic.AddInt32(t.dups, -1)
	if newval < 0 {
		os.RemoveAll(t.name)
	}
}

const ntlmNegotiateMessage = "NTLM TlRMTVNTUAABAAAAB7IIogwADAAzAAAACwALACgAAAAKAAAoAAAAD1dJTExISS1NQUlOTk9SVEhBTUVSSUNB"
