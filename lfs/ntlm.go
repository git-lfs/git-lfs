package lfs

import (
	"io"
	//"bufio"
	//"net"
	"net/http"
	//"encoding/xml"
	"encoding/base64"
	"bytes"
	"io/ioutil"
	"fmt"
	//"os"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)


type myReader struct {
    *bytes.Buffer
}

// So that it implements the io.ReadCloser interface
func (m myReader) Close() error { return nil } 

func (c *Configuration) NTLMSession() ntlm.ClientSession {
	
	if c.ntlmSession != nil {
		return c.ntlmSession
	}
	
	var session, _  = ntlm.CreateClientSession(ntlm.Version2, ntlm.ConnectionOrientedMode)
	session.SetUserInfo("user","pass","northamerica")
	
	c.ntlmSession = session
	
	return session
}

func DoNTLMRequest(request *http.Request) (*http.Response, error) {
			
	tracerx.Printf("DoNTLMRequest ENTER")
	defer tracerx.Printf("DoNTLMRequest LEAVE")
			
	if !Config.NTLM() {
		tracerx.Printf("DoNTLMRequest ntlm is not enabled")
		
		return nil, Error(fmt.Errorf("NTLM is not enabled"))
	}			
			
	handReq := cloneRequest(request)	
	res, nil := InitHandShake(handReq)
	
	//If the status is 401 then we need to re-authenticate, otherwise it was successful
	if res.StatusCode == 401 {
		
		negotiateReq := cloneRequest(request)
		challengeMessage := Negotiate(negotiateReq, getNegotiateMessage())
		
		challengeReq := cloneRequest(request)
		res, nil := Challenge(challengeReq, challengeMessage)
		
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

func Negotiate(request *http.Request, message string) []byte{

	request.Header.Add("Authorization", message)
	var response, err = Config.HttpClient().Do(request)
	
	if err != nil{
		tracerx.Printf("ntlm: Negotiate Error %s", err.Error())
	}
	
	ret := ParseChallengeMessage(response)
	
	//Always close negotiate to keep the connection alive
	//We never return the response from negotiate so we 
	//can't trust decodeApiResponse to decode it
	io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()
	
	return ret;
}

func Challenge(request *http.Request, challengeBytes []byte) (*http.Response, error){
	
	challenge, err := ntlm.ParseChallengeMessage(challengeBytes)
	
	if err != nil {
		return nil, Error(err)
	}
	
	Config.NTLMSession().ProcessChallengeMessage(challenge)
	authenticate, err := Config.NTLMSession().GenerateAuthenticateMessage()
	
	if err != nil {
		return nil, Error(err)
	}
	
	authenticateMessage := string(Concat([]byte("NTLM "), []byte(base64.StdEncoding.EncodeToString(authenticate.Bytes()))))
	
	request.Header.Add("Authorization", authenticateMessage)
	response, err := Config.HttpClient().Do(request)
	
	//io.Copy(ioutil.Discard, response.Body)
	//response.Body.Close()
	
	return response, nil
}

// get the bytes for the Type2 message
func ParseChallengeMessage(response *http.Response) []byte{
	
	if headers, ok := response.Header["Www-Authenticate"]; ok{
		
		//parse out the "NTLM " at the beginning of the resposne
		challenge := headers[0][5:]
		
		val, err := base64.StdEncoding.DecodeString(challenge)
		
		if err != nil{
			tracerx.Printf("ntlm: ParseChallengeMessage Error %s", err.Error())
		}
		
		return []byte(val)
	}
	
	panic("www-Authenticate header is not present")
}

func cloneRequest(request *http.Request) *http.Request {
	
	//We need to do some magic to copy the request without closing the body stream
	buf, _ := ioutil.ReadAll(request.Body)
	rdr1 := myReader{bytes.NewBuffer(buf)}
	rdr2 := myReader{bytes.NewBuffer(buf)}

	request.Body = rdr2 // OK since rdr2 implements the io.ReadCloser interface	
	
	clonedReq, _ := http.NewRequest(request.Method, request.URL.String(), rdr1)
	
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

func ConcatS(ar ...string) string {
	
    var buffer bytes.Buffer

    for _, s := range ar{
        buffer.WriteString(s)
    }

    return buffer.String()
}

func Concat(ar ...[]byte) []byte {
	return bytes.Join(ar, nil)
}