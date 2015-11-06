package lfs

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

func TestNtlmClientSession(t *testing.T) {

	//Make sure to clear ntlmSession so test order doesn't matter.
	Config.ntlmSession = nil

	creds := Creds{"username": "MOOSEDOMAIN\\canadian", "password": "MooseAntlersYeah"}
	_, err := Config.ntlmClientSession(creds)
	assert.Equal(t, err, nil)

	//The second call should ignore creds and give the session we just created.
	badCreds := Creds{"username": "badusername", "password": "MooseAntlersYeah"}
	_, err = Config.ntlmClientSession(badCreds)
	assert.Equal(t, err, nil)

	//clean up
	Config.ntlmSession = nil
}

func TestNtlmClientSessionBadCreds(t *testing.T) {

	//Make sure to clear ntlmSession so test order doesn't matter.
	Config.ntlmSession = nil

	creds := Creds{"username": "badusername", "password": "MooseAntlersYeah"}
	_, err := Config.ntlmClientSession(creds)
	assert.NotEqual(t, err, nil)

	//clean up
	Config.ntlmSession = nil
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
