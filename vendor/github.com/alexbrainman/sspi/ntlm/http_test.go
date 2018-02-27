// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package ntlm_test

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexbrainman/sspi/ntlm"
)

var (
	testURL = flag.String("url", "", "server URL for TestNTLMHTTPClient")
)

func newRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET", *testURL, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func get(req *http.Request) (*http.Response, string, error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}
	return res, string(body), nil
}

func canDoNTLM() error {
	req, err := newRequest()
	if err != nil {
		return err
	}
	res, _, err := get(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("Unauthorized expected, but got %v", res.StatusCode)
	}
	authHeaders, found := res.Header["Www-Authenticate"]
	if !found {
		return fmt.Errorf("Www-Authenticate not found")
	}
	for _, h := range authHeaders {
		if h == "NTLM" {
			return nil
		}
	}
	return fmt.Errorf("Www-Authenticate header does not contain NTLM, but has %v", authHeaders)
}

func doNTLMNegotiate(negotiate []byte) ([]byte, error) {
	req, err := newRequest()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "NTLM "+base64.StdEncoding.EncodeToString(negotiate))
	res, _, err := get(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusUnauthorized {
		return nil, fmt.Errorf("Unauthorized expected, but got %v", res.StatusCode)
	}
	authHeaders, found := res.Header["Www-Authenticate"]
	if !found {
		return nil, fmt.Errorf("Www-Authenticate not found")
	}
	if len(authHeaders) != 1 {
		return nil, fmt.Errorf("Only one Www-Authenticate header expected, but %d found: %v", len(authHeaders), authHeaders)
	}
	if len(authHeaders[0]) < 6 {
		return nil, fmt.Errorf("Www-Authenticate header is to short: %q", authHeaders[0])
	}
	if !strings.HasPrefix(authHeaders[0], "NTLM ") {
		return nil, fmt.Errorf("Www-Authenticate header is suppose to starts with \"NTLM \", but is %q", authHeaders[0])
	}
	authenticate, err := base64.StdEncoding.DecodeString(authHeaders[0][5:])
	if err != nil {
		return nil, err
	}
	return authenticate, nil
}

func doNTLMAuthenticate(authenticate []byte) (string, error) {
	req, err := newRequest()
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "NTLM "+base64.StdEncoding.EncodeToString(authenticate))
	res, body, err := get(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OK expected, but got %v", res.StatusCode)
	}
	return body, nil
}

func TestNTLMHTTPClient(t *testing.T) {
	// TODO: combine client and server tests so we don't need external server
	if len(*testURL) == 0 {
		t.Skip("Skipping due to empty \"url\" parameter")
	}

	cred, err := ntlm.AcquireCurrentUserCredentials()
	if err != nil {
		t.Fatal(err)
	}
	defer cred.Release()

	secctx, negotiate, err := ntlm.NewClientContext(cred)
	if err != nil {
		t.Fatal(err)
	}
	defer secctx.Release()

	err = canDoNTLM()
	if err != nil {
		t.Fatal(err)
	}
	challenge, err := doNTLMNegotiate(negotiate)
	if err != nil {
		t.Fatal(err)
	}
	authenticate, err := secctx.Update(challenge)
	if err != nil {
		t.Fatal(err)
	}
	_, err = doNTLMAuthenticate(authenticate)
	if err != nil {
		t.Fatal(err)
	}
}

// TODO: See http://www.innovation.ch/personal/ronald/ntlm.html#connections about needed to keep connection alive during authentication.

func TestNTLMHTTPServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement NTLM authentication here
		w.Write([]byte("hello"))
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q, want hello", string(got))
	}
}
