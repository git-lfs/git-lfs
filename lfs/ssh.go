package lfs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

type sshAuthResponse struct {
	Message   string            `json:"-"`
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt string            `json:"expires_at"`
}

func sshAuthenticate(endpoint Endpoint, operation, oid string) (sshAuthResponse, error) {

	// This is only used as a fallback where the Git URL is SSH but server doesn't support a full SSH binary protocol
	// and therefore we derive a HTTPS endpoint for binaries instead; but check authentication here via SSH

	res := sshAuthResponse{}
	if len(endpoint.SshUserAndHost) == 0 {
		return res, nil
	}

	tracerx.Printf("ssh: %s git-lfs-authenticate %s %s %s",
		endpoint.SshUserAndHost, endpoint.SshPath, operation, oid)

	exe, args := sshGetExeAndArgs(endpoint)
	args = append(args,
		"git-lfs-authenticate",
		endpoint.SshPath,
		operation, oid)

	cmd := exec.Command(exe, args...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		res.Message = string(out)
	} else {
		err = json.Unmarshal(out, &res)
	}

	return res, err
}

// Return the executable name for ssh on this machine and the base args
// Base args includes port settings, user/host, everything pre the command to execute
func sshGetExeAndArgs(endpoint Endpoint) (exe string, baseargs []string) {
	if len(endpoint.SshUserAndHost) == 0 {
		return "", nil
	}

	ssh := os.Getenv("GIT_SSH")
	basessh := filepath.Base(ssh)
	// Strip extension for easier comparison
	if ext := filepath.Ext(basessh); len(ext) > 0 {
		basessh = basessh[:len(basessh)-len(ext)]
	}
	isPlink := strings.EqualFold(basessh, "plink")
	isTortoise := strings.EqualFold(basessh, "tortoiseplink")
	if ssh == "" {
		ssh = "ssh"
	}

	args := make([]string, 0, 4)
	if isTortoise {
		// TortoisePlink requires the -batch argument to behave like ssh/plink
		args = append(args, "-batch")
	}

	if len(endpoint.SshPort) > 0 {
		if isPlink || isTortoise {
			args = append(args, "-P")
		} else {
			args = append(args, "-p")
		}
		args = append(args, endpoint.SshPort)
	}
	args = append(args, endpoint.SshUserAndHost)

	return ssh, args
}

// Below here is the pure-SSH API interface
// The API is basically the same except there's no need for hypermedia links
func NewSshApiContext(endpoint Endpoint) ApiContext {
	ctx := &SshApiContext{endpoint: endpoint}

	err := ctx.connect()
	if err != nil {
		// Rejected SSH connection, use fallback
		return nil
	}
	return ctx
}

// Create a manually initialised API context where I/O is already running
func NewManualSSHApiContext(in io.WriteCloser, out io.ReadCloser) *SshApiContext {
	return &SshApiContext{
		stdin:     in,
		stdout:    out,
		bufReader: bufio.NewReader(out),
	}

}

type SshApiContext struct {
	// Endpoint which was used to open this connection
	endpoint Endpoint

	// The command which is running ssh
	cmd *exec.Cmd
	// Native streams for communicating
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	// Buffered reader to allow searching for delimiters between JSON and blob data
	bufReader *bufio.Reader
}

func (self *SshApiContext) Endpoint() Endpoint {
	return self.endpoint
}

type ExitRequest struct {
}
type ExitResponse struct {
}

func (self *SshApiContext) Close() error {

	if self.stdin != nil && self.stdout != nil {
		// terminate server-side
		params := ExitRequest{}
		req, err := NewJsonRequest("Exit", params)
		if err != nil {
			return err
		}
		err = self.sendJSONRequest(req)
		if err != nil {
			return err
		}
	}
	var errbytes []byte
	if self.stderr != nil {
		var readerr error
		errbytes, readerr = ioutil.ReadAll(self.stderr)
		if readerr == nil && len(errbytes) > 0 {
			// Copy to our stderr for info
			fmt.Fprintf(os.Stderr, "Messages from SSH server:\n%v", string(errbytes))
		}
	}

	if self.cmd != nil {
		err := self.cmd.Wait()
		if err != nil {
			return fmt.Errorf("Error closing ssh connection: %v\nstderr: %v", err.Error(), string(errbytes))
		}
	}

	self.stdin = nil
	self.stdout = nil
	self.stderr = nil
	self.bufReader = nil
	self.cmd = nil

	return nil
}

func (self *SshApiContext) connect() error {
	ssh, args := sshGetExeAndArgs(self.endpoint)

	// Now add remote program and path
	serverCommand := "git-lfs-ssh-serve"
	if c, ok := Config.GitConfig("lfs.sshservercmd"); ok {
		serverCommand = c
	}
	args = append(args, serverCommand)
	args = append(args, self.endpoint.SshPath)

	cmd := exec.Command(ssh, args...)

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to connect to ssh stdout: %v", err.Error())
	}
	errp, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to connect to ssh stderr: %v", err.Error())
	}
	inp, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Unable to connect to ssh stdin: %v", err.Error())
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Unable to start ssh command: %v", err.Error())
	}

	self.cmd = cmd
	self.stdin = inp
	self.stdout = outp
	self.stderr = errp
	self.bufReader = bufio.NewReader(outp)

	// test that the server works
	_, _, _, err = self.ServerVersion()
	if err != nil {
		return fmt.Errorf("Failure to start up full SSH server, probably not supported: %v", err)
	}

	return nil

}

// Utility methods for JSON-RPC style request/response over SSH
// This works for any persistent connection and could be used elsewhere too but for now, localised
// Note *not* using net/rpc and net/rpc/jsonrpc because we want more control
// golang's rpc requires a certain method format (Object.Method) and also doesn't easily
// support interleaving with raw byte streams like we need to.
// as per http://www.jsonrpc.org/specification
type JsonRequest struct {
	Id     int    `json:"id"`
	Method string `json:"method"`
	// RawMessage allows us to store late-resolved, message-specific nested types
	// requires an extra couple of steps though; even though RawMessage is a []byte, it's not
	// JSON itself. You need to convert JSON to/from RawMessage as well as JSON to/from the structure
	// - see RawMessage's own UnmarshalJSON/MarshalJSON for this extra step
	Params *json.RawMessage `json:"params"`
}
type JsonResponse struct {
	Id    int         `json:"id"`
	Error interface{} `json:"error"`
	// RawMessage allows us to store late-resolved, message-specific nested types
	// requires an extra couple of steps though; even though RawMessage is a []byte, it's not
	// JSON itself. You need to convert JSON to/from RawMessage as well as JSON to/from the structure
	// - see RawMessage's own UnmarshalJSON/MarshalJSON for this extra step
	Result *json.RawMessage `json:"result"`
}

var (
	latestRequestId int = 1
)

func NewJsonRequest(method string, params interface{}) (*JsonRequest, error) {
	ret := &JsonRequest{
		Id:     latestRequestId,
		Method: method,
	}
	var err error
	ret.Params, err = EmbedStructInJsonRawMessage(params)
	latestRequestId++
	return ret, err
}

func NewJsonResponse(id int, result interface{}) (*JsonResponse, error) {
	ret := &JsonResponse{
		Id: id,
	}
	var err error
	ret.Result, err = EmbedStructInJsonRawMessage(result)
	return ret, err
}
func NewJsonErrorResponse(id int, err interface{}) *JsonResponse {
	ret := &JsonResponse{
		Id:    id,
		Error: err,
	}
	return ret
}

func EmbedStructInJsonRawMessage(in interface{}) (*json.RawMessage, error) {
	// Encode nested struct ready for transmission so that it can be late unmarshalled at the other end
	// Need to do this & declare as RawMessage rather than interface{} in struct otherwise unmarshalling
	// at other end will turn it into a simple array/map
	// Doesn't affect the wire bytes; they're still nested JSON in the same way as if you marshalled the whole struct
	// this is just a golang method to defer resolving on unmarshal
	ret := &json.RawMessage{}
	innerbytes, err := json.Marshal(in)
	if err != nil {
		return ret, fmt.Errorf("Unable to marshal struct to JSON: %v %v", in, err.Error())
	}
	err = ret.UnmarshalJSON(innerbytes)
	if err != nil {
		return ret, fmt.Errorf("Unable to convert JSON to RawMessage: %v %v", string(innerbytes), err.Error())
	}

	return ret, nil

}

// Perform a full JSON-RPC style call with JSON request and response
func (self *SshApiContext) doFullJSONRequestResponse(method string, params interface{}, result interface{}) error {

	req, err := NewJsonRequest(method, params)
	if err != nil {
		return err
	}
	err = self.sendJSONRequest(req)
	if err != nil {
		return err
	}
	err = self.readFullJSONResponse(req, result)
	if err != nil {
		return err
	}
	// result is now populated
	return nil

}

// A wrapper around a reader for a multi-purpose stream which reads exactly a number of
// bytes before returning EOF, so that 'read to end' and 'close' just read the fixed number
// of bytes and leaves the stream available for subsequent use by others
// This is just like io.LimitedReader but responds to Close() too
type LimitedReadCloser struct {
	r io.Reader
}

func (self *LimitedReadCloser) Read(p []byte) (n int, err error) {
	return self.r.Read(p)
}
func (self *LimitedReadCloser) Close() error {
	// Don't actually close, leave stream open
	// but always make sure we read the full size data from the stream so it's left at the right point
	// r is already a LimitedReader so we'll stop at the right place
	// ignore errors, might already be at the end
	io.Copy(ioutil.Discard, self.r)

	return nil
}

// Return an initialised LimitedReadCloser
func LimitReadCloser(r io.Reader, sz int64) io.ReadCloser {
	return &LimitedReadCloser{io.LimitReader(r, sz)}
}

// Perform a JSON request that results in a byte stream as a response
func (self *SshApiContext) doJSONRequestDownload(method string, params interface{}, sz int64) (io.ReadCloser, error) {

	req, err := NewJsonRequest(method, params)
	if err != nil {
		return nil, err
	}
	err = self.sendJSONRequest(req)
	if err != nil {
		return nil, err
	}
	// Next response from the server is a raw byte stream of sz bytes
	// Wrap a fixed-lenth reader around the actual reader so any attempt to
	// read to end reads exactly the correct number of bytes
	return LimitReadCloser(self.bufReader, sz), nil

}

// Late-bind a method-specific structure from the raw message
func ExtractStructFromJsonRawMessage(raw *json.RawMessage, out interface{}) error {
	nestedbytes, err := raw.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Unable to extract type-specific JSON from: %v\n%v", string(*raw), err.Error())
	}
	err = json.Unmarshal(nestedbytes, &out)
	if err != nil {
		return fmt.Errorf("Unable to decode type-specific result: %v\n%v", string(nestedbytes), err.Error())
	}
	return nil

}

// Send a JSON request but don't read any response
func (self *SshApiContext) sendJSONRequest(req interface{}) error {
	if self.stdout == nil || self.bufReader == nil {
		return fmt.Errorf("Not connected")
	}

	reqbytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("Error encoding %v to JSON: %v", err.Error())
	}
	// Append the binary 0 delimiter that server uses to read up to
	reqbytes = append(reqbytes, byte(0))
	_, err = self.stdin.Write(reqbytes)
	if err != nil {
		return fmt.Errorf("Error writing request bytes to connection: %v", err.Error())
	}

	return nil
}

func (self *SshApiContext) readJSONResponse() (*JsonResponse, error) {
	jsonbytes, err := self.bufReader.ReadBytes(byte(0))
	if err != nil {
		return nil, fmt.Errorf("Unable to read response from server: %v", err.Error())
	}
	// remove terminator before unmarshalling
	jsonbytes = jsonbytes[:len(jsonbytes)-1]
	response := &JsonResponse{}
	err = json.Unmarshal(jsonbytes, response)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode JSON response from server: %v\n%v", string(jsonbytes), err.Error())
	}
	return response, nil
}

// Check a response object; req can be nil, if so doesn't check that Ids match
func (self *SshApiContext) checkJSONResponse(req *JsonRequest, resp *JsonResponse) error {
	if resp.Error != nil {
		return fmt.Errorf("Error response from server: %v", resp.Error)
	}
	if req != nil && req.Id != resp.Id {
		return fmt.Errorf("Response from server has wrong Id, request: %d response: %d", req.Id, resp.Id)
	}
	return nil
}

// Read a JSON response, check it, and pull out the nested method-specific & write to result
// originalReq is optional and can be left nil but if supplied Ids will be checked for matching
func (self *SshApiContext) readFullJSONResponse(originalReq *JsonRequest, result interface{}) error {
	// read response (buffered) up to binary 0 which terminates JSON
	response, err := self.readJSONResponse()
	if err != nil {
		return err
	}
	// early validation
	err = self.checkJSONResponse(originalReq, response)
	if err != nil {
		return err
	}
	// response.Result is left as raw since it depends on the type of the expected result
	// so now unmarshal the nested part
	err = ExtractStructFromJsonRawMessage(response.Result, &result)
	if err != nil {
		return err
	}
	return nil
}

func (self *SshApiContext) sendRawData(sz int64, source io.Reader) error {

	if sz == 0 {
		return nil
	}

	// no need to copy in blocks now since callback is handled in reader now
	n, err := io.CopyN(self.stdin, source, sz)
	if err != nil {
		return err
	}
	if n != sz {
		return fmt.Errorf("Transferred bytes did not match expected size; transferred %d, expected %d", n, sz)
	}

	return nil
}

type ServerVersionRequest struct {
}
type ServerVersionResponse struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// Get the version of the SSH server
func (self *SshApiContext) ServerVersion() (major, minor, patch int, e error) {
	verparams := ServerVersionRequest{}
	resp := ServerVersionResponse{}
	err := self.doFullJSONRequestResponse("Version", &verparams, &resp)
	if err != nil {
		return 0, 0, 0, Errorf(err, "Error calling Version: %v", err)
	}

	return resp.Major, resp.Minor, resp.Patch, nil

}

type DownloadCheckRequest struct {
	Oid string `json:"oid"`
}
type DownloadCheckResponse struct {
	Size int64 `json:"size"`
}
type DownloadRequest struct {
	Oid  string `json:"oid"`
	Size int64  `json:"size"`
}

func (self *SshApiContext) Download(oid string) (io.ReadCloser, int64, *WrappedError) {
	infoparams := DownloadCheckRequest{oid}
	resp := DownloadCheckResponse{}
	err := self.doFullJSONRequestResponse("DownloadCheck", &infoparams, &resp)
	if err != nil {
		sendApiEvent(apiEventFail)
		return nil, 0, Errorf(err, "Error while downloading %v (DownloadCheck): %v", oid, err)
	} else if resp.Size == -1 {
		sendApiEvent(apiEventFail)
		return nil, 0, Error(objectRelationDoesNotExist)
	}
	contentparams := DownloadRequest{
		Oid:  oid,
		Size: resp.Size,
	}

	r, err := self.doJSONRequestDownload("Download", &contentparams, resp.Size)
	if err != nil {
		sendApiEvent(apiEventFail)
		return nil, 0, Errorf(err, "Error while downloading %v (Download): %v", oid, err.Error())
	}

	sendApiEvent(apiEventSuccess)

	return r, resp.Size, nil
}

func (self *SshApiContext) DownloadCheck(oid string) (*ObjectResource, *WrappedError) {
	infoparams := DownloadCheckRequest{oid}
	resp := DownloadCheckResponse{}
	err := self.doFullJSONRequestResponse("DownloadCheck", &infoparams, &resp)
	if err != nil {
		sendApiEvent(apiEventFail)
		return nil, Errorf(err, "Error calling DownloadCheck for %v: %v", oid, err)
	}

	sendApiEvent(apiEventSuccess)

	if resp.Size >= 0 {
		return &ObjectResource{
			Oid:   oid,
			Size:  resp.Size,
			Links: self.makeDownloadLinks(oid)}, nil
	} else {
		// does not exist
		return nil, Error(objectRelationDoesNotExist)
	}

}

func (self *SshApiContext) DownloadObject(obj *ObjectResource) (io.ReadCloser, int64, *WrappedError) {
	contentparams := DownloadRequest{
		Oid:  obj.Oid,
		Size: obj.Size,
	}

	r, err := self.doJSONRequestDownload("Download", &contentparams, obj.Size)
	if err != nil {
		sendApiEvent(apiEventFail)
		return nil, 0, Errorf(err, "Error while downloading %v: %v", obj.Oid, err.Error())
	}

	sendApiEvent(apiEventSuccess)

	return r, obj.Size, nil

}

type UploadRequest struct {
	Oid  string `json:"oid"`
	Size int64  `json:"size"`
}

type UploadResponse struct {
	OkToSend bool `json:"okToSend"`
}
type UploadCompleteResponse struct {
	ReceivedOk bool `json:"receivedOk"`
}

func (self *SshApiContext) makeUploadLinks(oid string) map[string]*linkRelation {
	// SSH doesn't need links to upload, but include dummy ones to play nice
	return map[string]*linkRelation{
		"upload": &linkRelation{Href: oid},
		"verify": &linkRelation{Href: oid},
	}
}
func (self *SshApiContext) makeDownloadLinks(oid string) map[string]*linkRelation {
	// SSH doesn't need links to upload, but include dummy ones to play nice
	return map[string]*linkRelation{
		"download": &linkRelation{Href: oid},
	}
}

func (self *SshApiContext) UploadCheck(oid string, sz int64) (*ObjectResource, *WrappedError) {
	params := UploadRequest{oid, sz}

	resp := UploadResponse{}
	err := self.doFullJSONRequestResponse("UploadCheck", &params, &resp)
	if err != nil {
		sendApiEvent(apiEventFail)
		return nil, Errorf(err, "Error while uploading %v (while sending Upload JSON request): %v", oid, err.Error())
	}
	sendApiEvent(apiEventSuccess)
	if !resp.OkToSend {
		// File already exists, behave like HTTP 200
		return nil, nil
	}

	return &ObjectResource{
		Oid:   oid,
		Size:  sz,
		Links: self.makeUploadLinks(oid)}, nil

}
func (self *SshApiContext) UploadObject(o *ObjectResource, content io.Reader) *WrappedError {
	params := UploadRequest{o.Oid, o.Size}

	resp := UploadResponse{}
	err := self.doFullJSONRequestResponse("Upload", &params, &resp)
	if err != nil {
		sendApiEvent(apiEventFail)
		return Errorf(err, "Error while uploading %v (while sending Upload JSON request): %v", o.Oid, err.Error())
	}
	if resp.OkToSend {
		// Send data, this does it in batches and calls back
		err = self.sendRawData(o.Size, content)
		if err != nil {
			sendApiEvent(apiEventFail)
			return Errorf(err, "Error while uploading %v (while sending raw content): %v", o.Oid, err.Error())
		}
		// Now read response to sent data
		received := UploadCompleteResponse{}
		err = self.readFullJSONResponse(nil, &received)
		if err != nil {
			sendApiEvent(apiEventFail)
			return Errorf(err, "Error while uploading %v (response to raw content): %v", o.Oid, err.Error())
		}
		if !received.ReceivedOk {
			sendApiEvent(apiEventFail)
			return Errorf(err, "Data not fully received while uploading %v", o.Oid)
		}

	}

	sendApiEvent(apiEventSuccess)

	return nil

}

type BatchRequestObject struct {
	Oid  string `json:"oid"`
	Size int64  `json:"size"`
}
type BatchResponseObject struct {
	Oid    string `json:"oid"`
	Action string `json:"action"`
	Size   int64  `json:"size"`
}
type BatchRequest struct {
	Objects []BatchRequestObject `json:"objects"`
}
type BatchResponse struct {
	Results []BatchResponseObject `json:"results"`
}

func (self *SshApiContext) Batch(objects []*ObjectResource) ([]*ObjectResource, *WrappedError) {
	params := BatchRequest{make([]BatchRequestObject, 0, len(objects))}
	for _, o := range objects {
		params.Objects = append(params.Objects, BatchRequestObject{o.Oid, o.Size})
	}

	resp := BatchResponse{}
	err := self.doFullJSONRequestResponse("Batch", &params, &resp)
	if err != nil {
		sendApiEvent(apiEventFail)
		return nil, Errorf(err, "Error in Batch: %v", err)
	}

	retObjs := make([]*ObjectResource, 0, len(resp.Results))
	for _, r := range resp.Results {
		newobj := &ObjectResource{Oid: r.Oid, Size: r.Size}
		if r.Action == "download" {
			newobj.Links = self.makeDownloadLinks(r.Oid)
		} else if r.Action == "upload" {
			newobj.Links = self.makeUploadLinks(r.Oid)
		}
		retObjs = append(retObjs, newobj)
	}

	return retObjs, nil

}
