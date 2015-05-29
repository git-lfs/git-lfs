package lfs

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := os.Getenv("GIT_SSH")
	os.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := os.Getenv("GIT_SSH")
	os.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

type TestStruct struct {
	Name      string
	Something int
}

func TestSSHEncodeJSONRequest(t *testing.T) {
	params := &TestStruct{Name: "Fred", Something: 99}
	req, err := NewJsonRequest("TestMethod", params)
	assert.Equal(t, nil, err)
	reqbytes, err := json.Marshal(req)
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"id":1,"method":"TestMethod","params":{"Name":"Fred","Something":99}}`, string(reqbytes))

}

func TestSSHDecodeJSONResponse(t *testing.T) {
	inputstruct := TestStruct{Name: "Fred", Something: 99}
	resp, err := NewJsonResponse(1, inputstruct)
	assert.Equal(t, nil, err)
	outputstruct := TestStruct{}
	// Now unmarshal nested result; need to extract json first
	innerbytes, err := resp.Result.MarshalJSON()
	assert.Equal(t, nil, err)
	err = json.Unmarshal(innerbytes, &outputstruct)
	assert.Equal(t, inputstruct, outputstruct)
}

// to be intialised
var (
	testoid           string
	testcontent       []byte
	batchExpected     map[string]string // oid -> ["upload" | "download"]
	batchDownloadSize int64             = 500
	batchUploadSize   int64             = 250
)

func init() {
	testcontent = []byte("Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
	hasher := sha256.New()
	inbuf := bytes.NewBuffer(testcontent)
	io.Copy(hasher, inbuf)
	testoid = hex.EncodeToString(hasher.Sum(nil))

	// Add some expected results for batch call
	batchExpected = make(map[string]string)
	batchExpected["111111111"] = "upload"
	batchExpected["111111112"] = "download"
	batchExpected["111111113"] = "download"
	batchExpected["111111114"] = "upload"
	batchExpected["111111115"] = "upload"
	batchExpected["111111116"] = "download"
	batchExpected["111111117"] = "upload"

}

// Test server function here, just called over a pipe to test
var testserve = func(conn net.Conn, t *testing.T) {
	// Man using assertions in a goroutine is much easier with Ginkgo
	defer func() {
		e := recover()
		if e != nil {
			t.Error(e)
		}
	}()
	defer conn.Close()
	// Run in a goroutine, be the server you seek
	// Read a request
	rdr := bufio.NewReader(conn)
	for {
		jsonbytes, err := rdr.ReadBytes(byte(0))
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(fmt.Sprintf("Test persistent server: unable to read from client: %v", err.Error()))
		}
		// slice off the terminator
		jsonbytes = jsonbytes[:len(jsonbytes)-1]
		var req JsonRequest
		err = json.Unmarshal(jsonbytes, &req)
		if err != nil {
			panic(fmt.Sprintf("Test persistent server: unable to unmarshal json request from client:%v", string(jsonbytes)))
		}

		var resp *JsonResponse
		switch req.Method {
		case "UploadCheck":
			upreq := UploadRequest{}
			ExtractStructFromJsonRawMessage(req.Params, &upreq)
			startresult := UploadResponse{}
			if upreq.Oid == testoid {
				startresult.OkToSend = true
			}
			resp, err = NewJsonResponse(req.Id, startresult)
			if err != nil {
				panic("Test persistent server: unable to create response")
			}

		case "Upload":
			upreq := UploadRequest{}
			ExtractStructFromJsonRawMessage(req.Params, &upreq)
			startresult := UploadResponse{}
			startresult.OkToSend = true
			// Send start response immediately
			resp, err = NewJsonResponse(req.Id, startresult)
			if err != nil {
				panic("Test persistent server: unable to create response")
			}
			responseBytes, err := json.Marshal(resp)
			if err != nil {
				panic("Test persistent server: unable to marshal response")
			}
			// null terminate response
			responseBytes = append(responseBytes, byte(0))
			conn.Write(responseBytes)
			// Next should by byte stream
			// Must read from buffered reader since bytes may have been read already
			receivedresult := UploadCompleteResponse{}
			receivedresult.ReceivedOk = true
			var receiveerr error
			// make pre-sized buffer
			contentbuf := bytes.NewBuffer(make([]byte, 0, upreq.Size))
			bytesLeft := upreq.Size
			for bytesLeft > 0 {
				c := int64(100)
				if c > bytesLeft {
					c = bytesLeft
				}
				n, err := io.CopyN(contentbuf, rdr, c)
				bytesLeft -= int64(n)
				if err != nil {
					receivedresult.ReceivedOk = false
					receiveerr = fmt.Errorf("Test persistent server: unable to read data: %v", err.Error())
					break
				}
			}
			// Check we received what we expected to receive
			contentbytes := contentbuf.Bytes()
			assert.Equal(t, contentbytes, testcontent)

			// After we've read all the content bytes, send received response
			if receiveerr != nil {
				resp = NewJsonErrorResponse(req.Id, receiveerr.Error())
			} else {
				resp, _ = NewJsonResponse(req.Id, receivedresult)
			}
		case "DownloadCheck":
			downreq := DownloadCheckRequest{}
			ExtractStructFromJsonRawMessage(req.Params, &downreq)
			result := DownloadCheckResponse{}
			if downreq.Oid == testoid {
				result.Size = int64(len(testcontent))
				resp, err = NewJsonResponse(req.Id, result)
				if err != nil {
					panic("Test persistent server: unable to create response")
				}
			} else {
				// Error response for missing item
				resp = NewJsonErrorResponse(req.Id, "Does not exist")
			}
		case "Download":
			// Can't return any error responses here (byte stream response only), have to just fail
			downreq := DownloadRequest{}
			ExtractStructFromJsonRawMessage(req.Params, &downreq)
			// there is no response to this
			sz := int64(len(testcontent))
			datasrc := bytes.NewReader(testcontent)
			// confirm size for testing
			if sz != downreq.Size {
				panic("Test persistent server: download size incorrect")
			}
			bytesLeft := sz
			for bytesLeft > 0 {
				c := int64(100)
				if c > bytesLeft {
					c = bytesLeft
				}
				n, err := io.CopyN(conn, datasrc, c)
				bytesLeft -= int64(n)
				if err != nil {
					panic(fmt.Sprintf("Test persistent server: unable to read data: %v", err))
				}
			}
		case "Batch":
			batchreq := BatchRequest{}
			ExtractStructFromJsonRawMessage(req.Params, &batchreq)
			result := BatchResponse{}
			for _, o := range batchreq.Objects {
				op, ok := batchExpected[o.Oid]
				if ok {
					var sz int64
					if op == "download" {
						sz = batchDownloadSize // would normally be stat()
					} else {
						sz = o.Size
					}
					result.Results = append(result.Results, BatchResponseObject{o.Oid, op, sz})
				} else {
					// Assume upload
					result.Results = append(result.Results, BatchResponseObject{o.Oid, "upload", o.Size})
				}
			}
			resp, err = NewJsonResponse(req.Id, result)
			if err != nil {
				panic("Test persistent server: unable to create response")
			}
		case "Exit":
			break
		default:
			resp = NewJsonErrorResponse(req.Id, fmt.Sprintf("Unknown method %v", req.Method))
		}
		if resp != nil {
			responseBytes, err := json.Marshal(resp)
			if err != nil {
				panic("Test persistent server: unable to marshal response")
			}

			// null terminate response
			responseBytes = append(responseBytes, byte(0))
			conn.Write(responseBytes)
		}
	}
	conn.Close()
}

func TestSSHDownloadCheck(t *testing.T) {
	cli, srv := net.Pipe()
	go testserve(srv, t)
	defer cli.Close()
	// Create a test SSH context from this which doesn't actually connect in
	// the traditional way
	ctx := NewManualSSHApiContext(cli, cli)
	// First test an invalid oid
	o, err := ctx.DownloadCheck("0000000000")
	// Should be a specific error in this case
	assert.NotEqual(t, (*WrappedError)(nil), err)
	assert.Equal(t, (*ObjectResource)(nil), o)

	// Now test valid one
	o, err = ctx.DownloadCheck(testoid)
	// for some reason the assert lib doesn't behave correctly for *WrappedError
	//assert.Equal(t, nil, err)
	if err != nil {
		t.Error("Should not be an error calling DownloadCheck with the correct Oid")
	}
	assert.Equal(t, int64(len(testcontent)), o.Size)
	ctx.Close()
}

func TestSSHDownload(t *testing.T) {
	cli, srv := net.Pipe()
	go testserve(srv, t)
	defer cli.Close()
	// Create a test SSH context from this which doesn't actually connect in
	// the traditional way
	ctx := NewManualSSHApiContext(cli, cli)
	// First test an invalid oid
	_, sz, err := ctx.Download("00000")
	// Should be an error in this case
	assert.NotEqual(t, nil, err)
	assert.Equal(t, int64(0), sz)

	// Now test valid one
	rdr, sz, err := ctx.Download(testoid)
	// for some reason the assert lib doesn't behave correctly for *WrappedError
	//assert.Equal(t, nil, err)
	assert.Equal(t, (*WrappedError)(nil), err)
	assert.Equal(t, int64(len(testcontent)), sz)
	var buf bytes.Buffer
	io.CopyN(&buf, rdr, sz) // must read before assert otherwise will get clogged
	assert.Equal(t, testcontent, buf.Bytes())

	ctx.Close()
}

func TestSSHUploadCheck(t *testing.T) {
	cli, srv := net.Pipe()
	go testserve(srv, t)
	defer cli.Close()
	ctx := NewManualSSHApiContext(cli, cli)

	// Test one that should work
	o, err := ctx.UploadCheck(testoid, int64(len(testcontent)))
	assert.Equal(t, (*WrappedError)(nil), err)
	assert.NotEqual(t, (*ObjectResource)(nil), o)
	assert.Equal(t, testoid, o.Oid)
	assert.Equal(t, int64(len(testcontent)), o.Size)

	// Test one that should reject upload (already exists)
	o, err = ctx.UploadCheck("0000000000000", int64(len(testcontent)))
	assert.Equal(t, (*WrappedError)(nil), err)
	assert.Equal(t, (*ObjectResource)(nil), o)

	ctx.Close()
}

func TestSSHUploadObject(t *testing.T) {
	cli, srv := net.Pipe()
	go testserve(srv, t)
	defer cli.Close()
	ctx := NewManualSSHApiContext(cli, cli)

	rdr := bytes.NewReader(testcontent)
	err := ctx.UploadObject(&ObjectResource{Oid: testoid, Size: int64(len(testcontent))}, rdr)
	assert.Equal(t, (*WrappedError)(nil), err)
	ctx.Close()
}

func TestSSHBatch(t *testing.T) {
	cli, srv := net.Pipe()
	go testserve(srv, t)
	defer cli.Close()
	ctx := NewManualSSHApiContext(cli, cli)

	var objects []*ObjectResource
	for oid, op := range batchExpected {
		var sz int64
		// only specify size if uploading (otherwise it's a query)
		if op == "upload" {
			sz = 250
		}
		objects = append(objects, &ObjectResource{Oid: oid, Size: sz})
	}
	retobjs, err := ctx.Batch(objects)
	assert.Equal(t, (*WrappedError)(nil), err)
	// Check the results
	for _, o := range retobjs {
		expectedop := batchExpected[o.Oid]
		assert.Equal(t, expectedop == "upload", o.CanUpload())
		assert.Equal(t, expectedop == "download", o.CanDownload())
		var expectedSz int64
		if expectedop == "download" {
			expectedSz = batchDownloadSize
		} else {
			expectedSz = batchUploadSize
		}
		assert.Equal(t, expectedSz, o.Size)
	}
	ctx.Close()

}
