package hawserclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/hawser/git-hawser/hawser"
	"github.com/rubyist/tracerx"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

const (
	gitMediaType     = "application/vnd.hawser"
	gitMediaMetaType = gitMediaType + "+json; charset=utf-8"
)

type linkMeta struct {
	Links map[string]*link `json:"_links,omitempty"`
}

type link struct {
	Href   string            `json:"href"`
	Header map[string]string `json:"header,omitempty"`
}

func Options(filehash string) (int, error) {
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

func Put(filehash, filename string, cb hawser.CopyCallback) error {
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
	reader := &hawser.CallbackReader{
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

func ExternalPut(filehash, filename string, lm *linkMeta, cb hawser.CopyCallback) error {
	link, ok := lm.Links["upload"]
	if !ok {
		return hawser.Error(errors.New("No upload link provided"))
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
	reader := &hawser.CallbackReader{
		C:         cb,
		TotalSize: fileSize,
		Reader:    file,
	}

	req, err := http.NewRequest("PUT", link.Href, nil)
	if err != nil {
		return hawser.Error(err)
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
		return hawser.Error(err)
	}
	tracerx.Printf("external_put_status: %d", res.StatusCode)

	// Run the callback
	if cb, ok := lm.Links["callback"]; ok {
		oid := filepath.Base(filehash)
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return hawser.Error(err)
		}

		cbreq, err := http.NewRequest("POST", cb.Href, nil)
		if err != nil {
			return hawser.Error(err)
		}
		for h, v := range cb.Header {
			cbreq.Header.Set(h, v)
		}

		d := fmt.Sprintf(`{"oid":"%s", "size":%d, "status":%d, "body":"%s"}`, oid, fileSize, res.StatusCode, string(body))
		cbreq.Body = ioutil.NopCloser(bytes.NewBufferString(d))

		tracerx.Printf("callback: %s %s", oid, cb.Href)
		cbres, err := http.DefaultClient.Do(cbreq)
		if err != nil {
			return hawser.Error(err)
		}
		tracerx.Printf("callback_status: %d", cbres.StatusCode)
	}

	return nil
}

func Post(filehash, filename string) (*linkMeta, int, error) {
	oid := filepath.Base(filehash)
	req, creds, err := clientRequest("POST", "")
	if err != nil {
		return nil, 0, hawser.Error(err)
	}

	file, err := os.Open(filehash)
	if err != nil {
		return nil, 0, hawser.Error(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, 0, hawser.Error(err)
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

	if res.StatusCode == 201 {
		var lm linkMeta
		dec := json.NewDecoder(res.Body)
		err := dec.Decode(&lm)
		if err != nil {
			return nil, res.StatusCode, hawser.Error(err)
		}

		return &lm, res.StatusCode, nil
	}

	return nil, res.StatusCode, nil
}

func Get(filename string) (io.ReadCloser, int64, *hawser.WrappedError) {
	oid := filepath.Base(filename)
	req, creds, err := clientRequest("GET", oid)
	if err != nil {
		return nil, 0, hawser.Error(err)
	}

	req.Header.Set("Accept", gitMediaType)
	res, wErr := doRequest(req, creds)

	if wErr != nil {
		return nil, 0, wErr
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		wErr = hawser.Error(errors.New("Empty Content-Type"))
		setErrorResponseContext(wErr, res)
		return nil, 0, wErr
	}

	if ok, wErr := validateMediaHeader(contentType, res.Body); !ok {
		setErrorResponseContext(wErr, res)
		return nil, 0, wErr
	}

	return res.Body, res.ContentLength, nil
}

func validateMediaHeader(contentType string, reader io.Reader) (bool, *hawser.WrappedError) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false, hawser.Errorf(err, "Invalid Media Type: %s", contentType)
	}

	if mediaType == gitMediaType {

		givenHeader, ok := params["header"]
		if !ok {
			return false, hawser.Error(fmt.Errorf("Missing Git Media header in %s", contentType))
		}

		fullGivenHeader := "--" + givenHeader + "\n"

		header := make([]byte, len(fullGivenHeader))
		_, err = io.ReadAtLeast(reader, header, len(fullGivenHeader))
		if err != nil {
			return false, hawser.Errorf(err, "Error reading response body.")
		}

		if string(header) != fullGivenHeader {
			return false, hawser.Error(fmt.Errorf("Invalid header: %s expected, got %s", fullGivenHeader, header))
		}
	}
	return true, nil
}

func doRequest(req *http.Request, creds Creds) (*http.Response, *hawser.WrappedError) {
	res, err := hawser.HttpClient().Do(req)

	var wErr *hawser.WrappedError

	if err == nil {
		if res.StatusCode > 299 {
			// An auth error should be 403.  Could be 404 also.
			if res.StatusCode < 405 {
				execCreds(creds, "reject")

				apierr := &Error{}
				dec := json.NewDecoder(res.Body)
				if err := dec.Decode(apierr); err != nil {
					wErr = hawser.Errorf(err, "Error decoding JSON from response")
				} else {
					wErr = hawser.Errorf(apierr, "Invalid response: %d", res.StatusCode)
				}
			}
		} else {
			execCreds(creds, "approve")
		}
	} else {
		wErr = hawser.Errorf(err, "Error sending HTTP request to %s", req.URL.String())
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

func setErrorRequestContext(err *hawser.WrappedError, req *http.Request) {
	err.Set("Endpoint", hawser.Config.Endpoint())
	err.Set("URL", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	setErrorHeaderContext(err, "Response", req.Header)
}

func setErrorResponseContext(err *hawser.WrappedError, res *http.Response) {
	err.Set("Status", res.Status)
	setErrorHeaderContext(err, "Request", res.Header)
	setErrorRequestContext(err, res.Request)
}

func setErrorHeaderContext(err *hawser.WrappedError, prefix string, head http.Header) {
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
	u := ObjectUrl(oid)
	req, err := http.NewRequest(method, u.String(), nil)
	req.Header.Set("User-Agent", hawser.UserAgent)
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

func ObjectUrl(oid string) *url.URL {
	c := hawser.Config
	u, _ := url.Parse(c.Endpoint())
	u.Path = path.Join(u.Path, "objects", oid)
	return u
}

type Error struct {
	Message   string `json:"message"`
	RequestId string `json:"request_id,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}
