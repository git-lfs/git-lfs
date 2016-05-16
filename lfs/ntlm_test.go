package lfs

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/github/git-lfs/auth"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

func TestNtlmClientSession(t *testing.T) {

	//Make sure to clear ntlmSession so test order doesn't matter.
	config.Config.NtlmSession = nil

	creds := auth.Creds{"username": "MOOSEDOMAIN\\canadian", "password": "MooseAntlersYeah"}
	_, err := ntlmClientSession(config.Config, creds)
	assert.Equal(t, err, nil)

	//The second call should ignore creds and give the session we just created.
	badCreds := auth.Creds{"username": "badusername", "password": "MooseAntlersYeah"}
	_, err = ntlmClientSession(config.Config, badCreds)
	assert.Equal(t, err, nil)

	//clean up
	config.Config.NtlmSession = nil
}

func TestNtlmClientSessionBadCreds(t *testing.T) {

	//Make sure to clear ntlmSession so test order doesn't matter.
	config.Config.NtlmSession = nil

	creds := auth.Creds{"username": "badusername", "password": "MooseAntlersYeah"}
	_, err := ntlmClientSession(config.Config, creds)
	assert.NotEqual(t, err, nil)

	//clean up
	config.Config.NtlmSession = nil
}

func TestNtlmCloneRequest(t *testing.T) {
	req1, _ := http.NewRequest("Method", "url", nil)
	cloneOfReq1, err := cloneRequest(req1)
	assert.Equal(t, err, nil)
	assertRequestsEqual(t, req1, cloneOfReq1)

	req2, _ := http.NewRequest("Method", "url", bytes.NewReader([]byte("Moose can be request bodies")))
	cloneOfReq2, err := cloneRequest(req2)
	assert.Equal(t, err, nil)
	assertRequestsEqual(t, req2, cloneOfReq2)
}

func assertRequestsEqual(t *testing.T, req1 *http.Request, req2 *http.Request) {
	assert.Equal(t, req1.Method, req2.Method)

	for k, v := range req1.Header {
		assert.Equal(t, v, req2.Header[k])
	}

	if req1.Body == nil {
		assert.Equal(t, req2.Body, nil)
	} else {
		bytes1, _ := ioutil.ReadAll(req1.Body)
		bytes2, _ := ioutil.ReadAll(req2.Body)
		assert.Equal(t, bytes.Compare(bytes1, bytes2), 0)
	}
}

func TestNtlmHeaderParseValid(t *testing.T) {
	res := http.Response{}
	res.Header = make(map[string][]string)
	res.Header.Add("Www-Authenticate", "NTLM "+base64.StdEncoding.EncodeToString([]byte("I am a moose")))
	bytes, err := parseChallengeResponse(&res)
	assert.Equal(t, err, nil)
	assert.Equal(t, strings.HasPrefix(string(bytes), "NTLM"), false)
}

func TestNtlmHeaderParseInvalidLength(t *testing.T) {
	res := http.Response{}
	res.Header = make(map[string][]string)
	res.Header.Add("Www-Authenticate", "NTL")
	ret, err := parseChallengeResponse(&res)
	if ret != nil {
		t.Errorf("Unexpected challenge response: %v", ret)
	}

	if err == nil {
		t.Errorf("Expected error, got none!")
	}
}

func TestNtlmHeaderParseInvalid(t *testing.T) {
	res := http.Response{}
	res.Header = make(map[string][]string)
	res.Header.Add("Www-Authenticate", base64.StdEncoding.EncodeToString([]byte("NTLM I am a moose")))
	_, err := parseChallengeResponse(&res)
	assert.NotEqual(t, err, nil)
}

func TestCloneSmallBody(t *testing.T) {
	cloneable1, err := newCloneableBody(strings.NewReader("abc"), 5)
	if err != nil {
		t.Fatal(err)
	}

	cloneable2, err := cloneable1.CloneBody()
	if err != nil {
		t.Fatal(err)
	}

	assertCloneableBody(t, cloneable2, "abc", "abc")
	assertCloneableBody(t, cloneable1, "abc", "abc")
}

func TestCloneBigBody(t *testing.T) {
	cloneable1, err := newCloneableBody(strings.NewReader("abc"), 2)
	if err != nil {
		t.Fatal(err)
	}

	cloneable2, err := cloneable1.CloneBody()
	if err != nil {
		t.Fatal(err)
	}

	assertCloneableBody(t, cloneable2, "abc", "ab")
	assertCloneableBody(t, cloneable1, "abc", "ab")
}

func assertCloneableBody(t *testing.T, cloneable *cloneableBody, expectedBody, expectedBuffer string) {
	buffer := string(cloneable.bytes)
	if buffer != expectedBuffer {
		t.Errorf("Expected buffer %q, got %q", expectedBody, buffer)
	}

	if cloneable.closed {
		t.Errorf("already closed?")
	}

	by, err := ioutil.ReadAll(cloneable)
	if err != nil {
		t.Fatal(err)
	}

	if err := cloneable.Close(); err != nil {
		t.Errorf("Error closing: %v", err)
	}

	actual := string(by)
	if actual != expectedBody {
		t.Errorf("Expected to read %q, got %q", expectedBody, actual)
	}
}
