package lfsapi

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNTLMAuth(t *testing.T) {
	session, err := ntlm.CreateServerSession(ntlm.Version2, ntlm.ConnectionOrientedMode)
	require.Nil(t, err)
	session.SetUserInfo("ntlmuser", "ntlmpass", "NTLMDOMAIN")

	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reqIndex := atomic.LoadUint32(&called)
		atomic.AddUint32(&called, 1)

		authHeader := req.Header.Get("Authorization")
		t.Logf("REQUEST %d: %s %s", reqIndex, req.Method, req.URL)
		t.Logf("AUTH: %q", authHeader)

		// assert full body is sent each time
		by, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if assert.Nil(t, err) {
			assert.Equal(t, "ntlm", string(by))
		}

		switch authHeader {
		case "":
			w.Header().Set("Www-Authenticate", "ntlm")
			w.WriteHeader(401)
		case ntlmNegotiateMessage:
			assert.True(t, strings.HasPrefix(req.Header.Get("Authorization"), "NTLM "))
			ch, err := session.GenerateChallengeMessage()
			if !assert.Nil(t, err) {
				t.Logf("challenge gen error: %+v", err)
				w.WriteHeader(500)
				return
			}
			chMsg := base64.StdEncoding.EncodeToString(ch.Bytes())
			w.Header().Set("Www-Authenticate", "ntlm "+chMsg)
			w.WriteHeader(401)
		default: // should be an auth msg
			authHeader := req.Header.Get("Authorization")
			assert.True(t, strings.HasPrefix(strings.ToUpper(authHeader), "NTLM "))
			auth := authHeader[5:] // strip "ntlm " prefix
			val, err := base64.StdEncoding.DecodeString(auth)
			if !assert.Nil(t, err) {
				t.Logf("auth base64 error: %+v", err)
				w.WriteHeader(500)
				return
			}

			_, err = ntlm.ParseAuthenticateMessage(val, 2)
			if !assert.Nil(t, err) {
				t.Logf("auth parse error: %+v", err)
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()

	req, err := http.NewRequest("POST", srv.URL+"/ntlm", NewByteBody([]byte("ntlm")))
	require.Nil(t, err)

	credHelper := newMockCredentialHelper()
	cli, err := NewClient(nil, TestEnv(map[string]string{
		"lfs.url":                         srv.URL + "/ntlm",
		"lfs." + srv.URL + "/ntlm.access": "ntlm",
	}))
	cli.Credentials = credHelper
	require.Nil(t, err)

	// ntlm support pulls domain and login info from git credentials
	srvURL, err := url.Parse(srv.URL)
	require.Nil(t, err)
	creds := Creds{
		"protocol": srvURL.Scheme,
		"host":     srvURL.Host,
		"path":     "ntlm",
		"username": "ntlmdomain\\ntlmuser",
		"password": "ntlmpass",
	}
	credHelper.Approve(creds)

	res, err := cli.DoWithAuth("remote", req)
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.True(t, credHelper.IsApproved(creds))
}

func TestNtlmClientSession(t *testing.T) {
	cli, err := NewClient(nil, nil)
	require.Nil(t, err)

	creds := Creds{"username": "MOOSEDOMAIN\\canadian", "password": "MooseAntlersYeah"}
	session1, err := cli.ntlmClientSession(creds)
	assert.Nil(t, err)
	assert.NotNil(t, session1)

	// The second call should ignore creds and give the session we just created.
	badCreds := Creds{"username": "MOOSEDOMAIN\\badusername", "password": "MooseAntlersYeah"}
	session2, err := cli.ntlmClientSession(badCreds)
	assert.Nil(t, err)
	assert.NotNil(t, session2)
	assert.EqualValues(t, session1, session2)
}

func TestNtlmClientSessionBadCreds(t *testing.T) {
	cli, err := NewClient(nil, nil)
	require.Nil(t, err)
	creds := Creds{"username": "badusername", "password": "MooseAntlersYeah"}
	_, err = cli.ntlmClientSession(creds)
	assert.NotNil(t, err)
}

func TestNtlmHeaderParseValid(t *testing.T) {
	res := http.Response{}
	res.Header = make(map[string][]string)
	res.Header.Add("Www-Authenticate", "NTLM "+base64.StdEncoding.EncodeToString([]byte("I am a moose")))
	bytes, err := parseChallengeResponse(&res)
	assert.Nil(t, err)
	assert.False(t, strings.HasPrefix(string(bytes), "NTLM"))
}

func TestNtlmHeaderParseInvalidLength(t *testing.T) {
	res := http.Response{}
	res.Header = make(map[string][]string)
	res.Header.Add("Www-Authenticate", "NTL")
	ret, err := parseChallengeResponse(&res)
	assert.NotNil(t, err)
	assert.Nil(t, ret)
}

func TestNtlmHeaderParseInvalid(t *testing.T) {
	res := http.Response{}
	res.Header = make(map[string][]string)
	res.Header.Add("Www-Authenticate", base64.StdEncoding.EncodeToString([]byte("NTLM I am a moose")))
	ret, err := parseChallengeResponse(&res)
	assert.NotNil(t, err)
	assert.Nil(t, ret)
}

func assertRequestsEqual(t *testing.T, req1 *http.Request, req2 *http.Request, req1Body []byte) {
	assert.Equal(t, req1.Method, req2.Method)

	for k, v := range req1.Header {
		assert.Equal(t, v, req2.Header[k])
	}

	if req1.Body == nil {
		assert.Nil(t, req2.Body)
	} else {
		bytes2, _ := ioutil.ReadAll(req2.Body)
		assert.Equal(t, req1Body, bytes2)
	}
}
