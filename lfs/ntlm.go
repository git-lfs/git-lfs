package lfs

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/github/git-lfs/vendor/_nuts/github.com/ThomsonReutersEikon/go-ntlm/ntlm"
)

func (c *Configuration) ntlmClientSession(creds Creds) (ntlm.ClientSession, error) {
	if c.ntlmSession != nil {
		return c.ntlmSession, nil
	}
	splits := strings.Split(creds["username"], "\\")

	if len(splits) != 2 {
		errorMessage := fmt.Sprintf("Your user name must be of the form DOMAIN\\user. It is currently %s", creds["username"], "string")
		return nil, errors.New(errorMessage)
	}

	session, err := ntlm.CreateClientSession(ntlm.Version2, ntlm.ConnectionOrientedMode)

	if err != nil {
		return nil, err
	}

	session.SetUserInfo(splits[1], creds["password"], strings.ToUpper(splits[0]))
	c.ntlmSession = session
	return session, nil
}

func DoNTLMRequest(request *http.Request, retry bool) (*http.Response, error) {
	handReq, err := cloneRequest(request)
	if err != nil {
		return nil, err
	}

	res, err := Config.HttpClient().Do(handReq)
	if err != nil && res == nil {
		return nil, err
	}

	//If the status is 401 then we need to re-authenticate, otherwise it was successful
	if res.StatusCode == 401 {

		creds, err := getCredsForAPI(request)
		if err != nil {
			return nil, err
		}

		negotiateReq, err := cloneRequest(request)
		if err != nil {
			return nil, err
		}

		challengeMessage, err := negotiate(negotiateReq, ntlmNegotiateMessage)
		if err != nil {
			return nil, err
		}

		challengeReq, err := cloneRequest(request)
		if err != nil {
			return nil, err
		}

		res, err := challenge(challengeReq, challengeMessage, creds)
		if err != nil {
			return nil, err
		}

		//If the status is 401 then we need to re-authenticate
		if res.StatusCode == 401 && retry == true {
			return DoNTLMRequest(challengeReq, false)
		}

		saveCredentials(creds, res)

		return res, nil
	}
	return res, nil
}

func negotiate(request *http.Request, message string) ([]byte, error) {
	request.Header.Add("Authorization", message)
	res, err := Config.HttpClient().Do(request)

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

func challenge(request *http.Request, challengeBytes []byte, creds Creds) (*http.Response, error) {
	challenge, err := ntlm.ParseChallengeMessage(challengeBytes)
	if err != nil {
		return nil, err
	}

	session, err := Config.ntlmClientSession(creds)
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
	return Config.HttpClient().Do(request)
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

	clonedReq.ContentLength = request.ContentLength

	return clonedReq, nil
}

func cloneRequestBody(req *http.Request) (io.ReadCloser, error) {
	if req.Body == nil {
		return nil, nil
	}

	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	return ioutil.NopCloser(bytes.NewBuffer(buf)), nil
}

const ntlmNegotiateMessage = "NTLM TlRMTVNTUAABAAAAB7IIogwADAAzAAAACwALACgAAAAKAAAoAAAAD1dJTExISS1NQUlOTk9SVEhBTUVSSUNB"
