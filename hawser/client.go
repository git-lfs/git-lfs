package hawser

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/rubyist/tracerx"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

const (
	gitMediaType     = "application/vnd.git-media"
	gitMediaMetaType = gitMediaType + "+json; charset=utf-8"
)

type linkMeta struct {
	Links map[string]*link `json:"_links,omitempty"`
}

func (l *linkMeta) Rel(name string) (*link, bool) {
	if l.Links == nil {
		return nil, false
	}

	lnk, ok := l.Links[name]
	return lnk, ok
}

type link struct {
	Href   string            `json:"href"`
	Header map[string]string `json:"header,omitempty"`
}

type UploadRequest struct {
	OidPath      string
	Filename     string
	CopyCallback CopyCallback
}

func Download(oidPath string) (io.ReadCloser, int64, *WrappedError) {
	oid := filepath.Base(oidPath)
	req, creds, err := clientRequest("GET", oid)
	if err != nil {
		return nil, 0, Error(err)
	}

	req.Header.Set("Accept", gitMediaType)
	res, wErr := doRequest(req, creds)

	if wErr != nil {
		return nil, 0, wErr
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		wErr = Error(errors.New("Empty Content-Type"))
		setErrorResponseContext(wErr, res)
		return nil, 0, wErr
	}

	ok, headerSize, wErr := validateMediaHeader(contentType, res.Body)
	if !ok {
		setErrorResponseContext(wErr, res)
		return nil, 0, wErr
	}

	return res.Body, res.ContentLength - int64(headerSize), nil
}

func Upload(oidPath, filename string, cb CopyCallback) *WrappedError {
	linkMeta, status, err := callPost(oidPath, filename)
	if err != nil && status != 302 {
		return Errorf(err, "Error starting file upload.")
	}

	oid := filepath.Base(oidPath)

	switch status {
	case 200: // object exists on the server
	case 405, 302:
		// Do the old style OPTIONS + PUT
		status, err := callOptions(oidPath)
		if err != nil {
			return Errorf(err, "Error getting options for file %s (%s)", filename, oid)
		}

		if status != 200 {
			err = callPut(oidPath, filename, cb)
			if err != nil {
				return Errorf(err, "Error uploading file %s (%s)", filename, oid)
			}
		}
	case 202:
		// the server responded with hypermedia links to upload and verify the object.
		err = callExternalPut(oidPath, filename, linkMeta, cb)
		if err != nil {
			return Errorf(err, "Error uploading file %s (%s)", filename, oid)
		}
	default:
		return Errorf(err, "Unexpected HTTP response: %d", status)
	}

	return nil
}

func callOptions(filehash string) (int, error) {
	oid := filepath.Base(filehash)
	_, err := os.Stat(filehash)
	if err != nil {
		return 0, err
	}

	tracerx.Printf("api_options: %s", oid)
	req, creds, err := clientRequest("OPTIONS", oid)
	if err != nil {
		return 0, err
	}

	res, wErr := doRequest(req, creds)
	if wErr != nil {
		return 0, wErr
	}
	tracerx.Printf("api_options_status: %d", res.StatusCode)

	return res.StatusCode, nil
}

func callPut(filehash, filename string, cb CopyCallback) error {
	if filename == "" {
		filename = filehash
	}

	oid := filepath.Base(filehash)
	file, err := os.Open(filehash)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	req, creds, err := clientRequest("PUT", oid)
	if err != nil {
		return err
	}

	fileSize := stat.Size()
	reader := &CallbackReader{
		C:         cb,
		TotalSize: fileSize,
		Reader:    file,
	}

	bar := pb.StartNew(int(fileSize))
	bar.SetUnits(pb.U_BYTES)
	bar.Start()

	req.Header.Set("Content-Type", gitMediaType)
	req.Header.Set("Accept", gitMediaMetaType)
	req.Body = ioutil.NopCloser(bar.NewProxyReader(reader))
	req.ContentLength = fileSize

	fmt.Printf("Sending %s\n", filename)

	tracerx.Printf("api_put: %s %s", oid, filename)
	res, wErr := doRequest(req, creds)
	if wErr != nil {
		return wErr
	}
	tracerx.Printf("api_put_status: %d", res.StatusCode)

	return nil
}

func callExternalPut(filehash, filename string, lm *linkMeta, cb CopyCallback) error {
	if lm == nil {
		return Error(errors.New("No hypermedia links provided"))
	}

	link, ok := lm.Rel("upload")
	if !ok {
		return Error(errors.New("No upload link provided"))
	}

	file, err := os.Open(filehash)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := stat.Size()
	reader := &CallbackReader{
		C:         cb,
		TotalSize: fileSize,
		Reader:    file,
	}

	req, err := http.NewRequest("PUT", link.Href, nil)
	if err != nil {
		return Error(err)
	}
	for h, v := range link.Header {
		req.Header.Set(h, v)
	}

	bar := pb.StartNew(int(fileSize))
	bar.SetUnits(pb.U_BYTES)
	bar.Start()

	req.Body = ioutil.NopCloser(bar.NewProxyReader(reader))
	req.ContentLength = fileSize

	tracerx.Printf("external_put: %s %s", filepath.Base(filehash), req.URL)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Error(err)
	}
	tracerx.Printf("external_put_status: %d", res.StatusCode)

	// Run the verify callback
	if cb, ok := lm.Rel("verify"); ok {
		oid := filepath.Base(filehash)

		cbreq, err := http.NewRequest("POST", cb.Href, nil)
		if err != nil {
			return Error(err)
		}
		for h, v := range cb.Header {
			cbreq.Header.Set(h, v)
		}

		d := fmt.Sprintf(`{"oid":"%s", "size":%d}`, oid, fileSize)
		cbreq.Body = ioutil.NopCloser(bytes.NewBufferString(d))

		tracerx.Printf("verify: %s %s", oid, cb.Href)
		cbres, err := http.DefaultClient.Do(cbreq)
		if err != nil {
			return Error(err)
		}
		tracerx.Printf("verify_status: %d", cbres.StatusCode)
	}

	return nil
}

func callPost(filehash, filename string) (*linkMeta, int, error) {
	oid := filepath.Base(filehash)
	req, creds, err := clientRequest("POST", "")
	if err != nil {
		return nil, 0, Error(err)
	}

	file, err := os.Open(filehash)
	if err != nil {
		return nil, 0, Error(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, 0, Error(err)
	}
	fileSize := stat.Size()

	d := fmt.Sprintf(`{"oid":"%s", "size":%d}`, oid, fileSize)
	req.Body = ioutil.NopCloser(bytes.NewBufferString(d))

	req.Header.Set("Accept", gitMediaMetaType)

	tracerx.Printf("api_post: %s %s", oid, filename)
	res, wErr := doRequest(req, creds)
	if wErr != nil {
		return nil, 0, wErr
	}
	tracerx.Printf("api_post_status: %d", res.StatusCode)

	if res.StatusCode == 202 {
		lm := &linkMeta{}
		dec := json.NewDecoder(res.Body)
		err := dec.Decode(lm)
		if err != nil {
			return nil, res.StatusCode, Errorf(err, "Error decoding JSON from %s %s.", req.Method, req.URL)
		}

		return lm, res.StatusCode, nil
	}

	return nil, res.StatusCode, nil
}

func validateMediaHeader(contentType string, reader io.Reader) (bool, int, *WrappedError) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	var headerSize int

	if err != nil {
		return false, headerSize, Errorf(err, "Invalid Media Type: %s", contentType)
	}

	if mediaType == gitMediaType {

		givenHeader, ok := params["header"]
		if !ok {
			return false, headerSize, Error(fmt.Errorf("Missing Git Media header in %s", contentType))
		}

		fullGivenHeader := "--" + givenHeader + "\n"
		headerSize = len(fullGivenHeader)

		header := make([]byte, headerSize)
		_, err = io.ReadAtLeast(reader, header, len(fullGivenHeader))
		if err != nil {
			return false, headerSize, Errorf(err, "Error reading response body.")
		}

		if string(header) != fullGivenHeader {
			return false, headerSize, Error(fmt.Errorf("Invalid header: %s expected, got %s", fullGivenHeader, header))
		}
	}
	return true, headerSize, nil
}

func doRequest(req *http.Request, creds Creds) (*http.Response, *WrappedError) {
	res, err := HttpClient().Do(req)

	var wErr *WrappedError

	if err == RedirectError {
		err = nil
	}

	if err == nil {
		if res.StatusCode > 299 {
			// An auth error should be 403.  Could be 404 also.
			if res.StatusCode < 405 {
				execCreds(creds, "reject")

				apierr := &ClientError{}
				dec := json.NewDecoder(res.Body)
				if err := dec.Decode(apierr); err != nil {
					wErr = Errorf(err, "Error decoding JSON from response")
				} else {
					wErr = Errorf(apierr, "Invalid response: %d", res.StatusCode)
				}
			}
		} else {
			execCreds(creds, "approve")
		}
	} else if res.StatusCode != 302 { // hack for pre-release
		wErr = Errorf(err, "Error sending HTTP request to %s", req.URL.String())
	}

	if wErr != nil {
		if res != nil {
			setErrorResponseContext(wErr, res)
		} else {
			setErrorRequestContext(wErr, req)
		}
	}

	return res, wErr
}

var hiddenHeaders = map[string]bool{
	"Authorization": true,
}

func setErrorRequestContext(err *WrappedError, req *http.Request) {
	err.Set("Endpoint", Config.Endpoint())
	err.Set("URL", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	setErrorHeaderContext(err, "Response", req.Header)
}

func setErrorResponseContext(err *WrappedError, res *http.Response) {
	err.Set("Status", res.Status)
	setErrorHeaderContext(err, "Request", res.Header)
	setErrorRequestContext(err, res.Request)
}

func setErrorHeaderContext(err *WrappedError, prefix string, head http.Header) {
	for key, _ := range head {
		contextKey := fmt.Sprintf("%s:%s", prefix, key)
		if _, skip := hiddenHeaders[key]; skip {
			err.Set(contextKey, "--")
		} else {
			err.Set(contextKey, head.Get(key))
		}
	}
}

func clientRequest(method, oid string) (*http.Request, Creds, error) {
	u := Config.ObjectUrl(oid)
	req, err := http.NewRequest(method, u.String(), nil)
	req.Header.Set("User-Agent", UserAgent)
	if err == nil {
		creds, err := credentials(u)
		if err != nil {
			return req, nil, err
		}

		token := fmt.Sprintf("%s:%s", creds["username"], creds["password"])
		auth := "Basic " + base64.URLEncoding.EncodeToString([]byte(token))
		req.Header.Set("Authorization", auth)
		return req, creds, nil
	}

	return req, nil, err
}

type ClientError struct {
	Message   string `json:"message"`
	RequestId string `json:"request_id,omitempty"`
}

func (e *ClientError) Error() string {
	return e.Message
}
