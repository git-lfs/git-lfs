package lfs

import (
	"bytes"
	"encoding/base64"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"	
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func (c *Configuration) NTLMSession(creds Creds) ntlm.ClientSession {
	
	if c.ntlmSession != nil {
		return c.ntlmSession
	}
	
	splits := strings.Split(creds["username"], "\\")
	
	var session, _  = ntlm.CreateClientSession(ntlm.Version2, ntlm.ConnectionOrientedMode)
	session.SetUserInfo(splits[0], creds["password"], splits[1])
	
	c.ntlmSession = session
	
	return session
}

func DoNTLMRequest(request *http.Request, retry bool) (*http.Response, error) {					
	handReq := cloneRequest(request)	
	res, nil := InitHandShake(handReq)

	//If the status is 401 then we need to re-authenticate, otherwise it was successful
	if res.StatusCode == 401 {
		
		creds, _ := getCredsForNTLM(request)
		
		negotiateReq := cloneRequest(request)
		challengeMessage := negotiate(negotiateReq, getNegotiateMessage())
		
		challengeReq := cloneRequest(request)
		res, _ := challenge(challengeReq, challengeMessage, creds)
		
		//If the status is 401 then we need to re-authenticate
		if res.StatusCode == 401 && retry == true {
			return DoNTLMRequest(challengeReq, false)
		}
		
		saveCredentials(creds, res)	
		
		return res, nil	
	}
	return res, nil	
}

func InitHandShake(request *http.Request) (*http.Response, error){
	var response, err = Config.HttpClient().Do(request)
	
	if err != nil {
		return nil, Error(err)
	}
	
	return response, nil
}

func negotiate(request *http.Request, message string) []byte{
	request.Header.Add("Authorization", message)
	var response, err = Config.HttpClient().Do(request)
	
	if err != nil{
		panic(err.Error())
	}
	
	ret := parseChallengeResponse(response)
	
	//Always close negotiate to keep the connection alive
	//We never return the response from negotiate so we 
	//can't trust decodeApiResponse to close it
	io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()
	
	return ret;
}

func challenge(request *http.Request, challengeBytes []byte, creds Creds) (*http.Response, error){
	challenge, err := ntlm.ParseChallengeMessage(challengeBytes)
	
	if err != nil {
		return nil, Error(err)
	}
	
	Config.NTLMSession(creds).ProcessChallengeMessage(challenge)
	authenticate, err := Config.NTLMSession(creds).GenerateAuthenticateMessage()
	
	if err != nil {
		return nil, Error(err)
	}
	
	authenticateMessage := concatS("NTLM ", base64.StdEncoding.EncodeToString(authenticate.Bytes()))
	
	request.Header.Add("Authorization", authenticateMessage)
	response, err := Config.HttpClient().Do(request)
	
	return response, nil
}

func parseChallengeResponse(response *http.Response) []byte{
	if headers, ok := response.Header["Www-Authenticate"]; ok{
		
		//parse out the "NTLM " at the beginning of the response
		challenge := headers[0][5:]
		val, err := base64.StdEncoding.DecodeString(challenge)
		
		if err != nil{
			panic(err.Error())
		}
		return []byte(val)
	}
	
	panic("www-Authenticate header is not present")
}

func cloneRequest(request *http.Request) *http.Request {
	var rdr1, rdr2 myReader
	var clonedReq *http.Request
	
	if request.Body != nil {
		//If we have a body (POST/PUT etc.)
		//We need to do some magic to copy the request without closing the body stream
		
		buf, _ := ioutil.ReadAll(request.Body)
		rdr1 = myReader{bytes.NewBuffer(buf)}
		rdr2 = myReader{bytes.NewBuffer(buf)}	
		request.Body = rdr2 // OK since rdr2 implements the io.ReadCloser interface	
		clonedReq, _ = http.NewRequest(request.Method, request.URL.String(), rdr1)	
	}else{
		clonedReq, _ = http.NewRequest(request.Method, request.URL.String(), nil)
	}
	
	for k, v := range request.Header {
		clonedReq.Header.Add(k,v[0])
	}
	
	clonedReq.ContentLength = request.ContentLength	
	
	return clonedReq
}

//Get Type 1 message
func getNegotiateMessage() string{
		
	//var negotiate, _ = session.GenerateNegotiateMessage()
	//return negotiate.Bytes
	
	return "NTLM TlRMTVNTUAABAAAAB7IIogwADAAzAAAACwALACgAAAAKAAAoAAAAD1dJTExISS1NQUlOTk9SVEhBTUVSSUNB"
}

func concatS(ar ...string) string {
	
    var buffer bytes.Buffer

    for _, s := range ar{
        buffer.WriteString(s)
    }

    return buffer.String()
}

func concat(ar ...[]byte) []byte {
	return bytes.Join(ar, nil)
}

type myReader struct {
    *bytes.Buffer
}

// So that myReader implements the io.ReadCloser interface
func (m myReader) Close() error { return nil } 