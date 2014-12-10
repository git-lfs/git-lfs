package gitmediaclient

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/github/git-media/gitmedia"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	gitMediaType     = "application/vnd.git-media"
	gitMediaMetaType = gitMediaType + "+json; charset=utf-8"
	gitMediaHeader   = "--git-media."
)

func Options(filehash string) (int, error) {
	oid := filepath.Base(filehash)
	_, err := os.Stat(filehash)
	if err != nil {
		return 0, err
	}

	req, creds, err := clientRequest("OPTIONS", oid)
	if err != nil {
		return 0, err
	}

	resp, wErr := doRequest(req, creds)
	if wErr != nil {
		return 0, wErr
	}

	return resp.StatusCode, nil
}

func Put(filehash, filename string, cb gitmedia.CopyCallback) error {
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
	reader := &gitmedia.CallbackReader{
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

	_, wErr := doRequest(req, creds)
	if wErr != nil {
		return wErr
	}

	return nil
}

func Get(filename string) (io.ReadCloser, int64, *gitmedia.WrappedError) {
	oid := filepath.Base(filename)
	req, creds, err := clientRequest("GET", oid)
	if err != nil {
		return nil, 0, gitmedia.Error(err)
	}

	req.Header.Set("Accept", gitMediaType)
	res, wErr := doRequest(req, creds)

	if wErr != nil {
		return nil, 0, wErr
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		wErr = gitmedia.Error(errors.New("Empty Content-Type"))
		setErrorResponseContext(wErr, res)
		return nil, 0, wErr
	}

	if ok, wErr := validateMediaHeader(contentType, res.Body); !ok {
		setErrorResponseContext(wErr, res)
		return nil, 0, wErr
	}

	return res.Body, res.ContentLength, nil
}

func validateMediaHeader(contentType string, reader io.Reader) (bool, *gitmedia.WrappedError) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false, gitmedia.Errorf(err, "Invalid Media Type: %s", contentType)
	}

	if mediaType != gitMediaType {
		return false, gitmedia.Error(fmt.Errorf("Invalid Media Type: %s expected, got %s", gitMediaType, mediaType))
	}

	givenHeader, ok := params["header"]
	if !ok {
		return false, gitmedia.Error(fmt.Errorf("Missing Git Media header in %s", contentType))
	}

	fullGivenHeader := "--" + givenHeader + "\n"

	header := make([]byte, len(fullGivenHeader))
	_, err = io.ReadAtLeast(reader, header, len(fullGivenHeader))
	if err != nil {
		return false, gitmedia.Errorf(err, "Error reading response body.")
	}

	if string(header) != fullGivenHeader {
		return false, gitmedia.Error(fmt.Errorf("Invalid header: %s expected, got %s", fullGivenHeader, header))
	}

	return true, nil
}

var httpClient *http.Client

func getHttpClient() *http.Client {
	if httpClient == nil {
		if len(os.Getenv("GIT_SSL_NO_VERIFY")) > 0 {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			httpClient = &http.Client{Transport: tr}
		} else {
			httpClient = http.DefaultClient
		}
	}

	return httpClient
}

func doRequest(req *http.Request, creds Creds) (*http.Response, *gitmedia.WrappedError) {
	res, err := getHttpClient().Do(req)

	var wErr *gitmedia.WrappedError

	if err == nil {
		if res.StatusCode > 299 {
			// An auth error should be 403.  Could be 404 also.
			if res.StatusCode < 405 {
				execCreds(creds, "reject")
			}

			apierr := &Error{}
			dec := json.NewDecoder(res.Body)
			if err := dec.Decode(apierr); err != nil {
				wErr = gitmedia.Errorf(err, "Error decoding JSON from response")
			} else {
				wErr = gitmedia.Errorf(apierr, "Invalid response: %d", res.StatusCode)
			}
		} else {
			execCreds(creds, "approve")
		}
	} else {
		wErr = gitmedia.Errorf(err, "Error sending HTTP request to %s", req.URL.String())
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

func setErrorRequestContext(err *gitmedia.WrappedError, req *http.Request) {
	err.Set("Endpoint", gitmedia.Config.Endpoint())
	err.Set("URL", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	setErrorHeaderContext(err, "Response", req.Header)
}

func setErrorResponseContext(err *gitmedia.WrappedError, res *http.Response) {
	err.Set("Status", res.Status)
	setErrorHeaderContext(err, "Request", res.Header)
	setErrorRequestContext(err, res.Request)
}

func setErrorHeaderContext(err *gitmedia.WrappedError, prefix string, head http.Header) {
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
	req.Header.Set("User-Agent", gitmedia.UserAgent)
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
	c := gitmedia.Config
	u, _ := url.Parse(c.Endpoint())
	if strings.HasSuffix(u.Path, "/") {
		u.Path = fmt.Sprintf("%sobjects/%s", u.Path, oid)
	} else {
		u.Path = fmt.Sprintf("%s/objects/%s", u.Path, oid)
	}
	return u
}

type Error struct {
	Message   string `json:"message"`
	RequestId string `json:"request_id,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}
