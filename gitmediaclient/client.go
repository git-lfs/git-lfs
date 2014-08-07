package gitmediaclient

import (
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
)

const (
	gitMediaType     = "application/vnd.git-media"
	gitMediaMetaType = gitMediaType + "+json; charset=utf-8"
	gitMediaHeader   = "--git-media."
)

func Options(filehash string) error {
	oid := filepath.Base(filehash)
	_, err := os.Stat(filehash)
	if err != nil {
		return err
	}

	req, creds, err := clientRequest("OPTIONS", oid)
	if err != nil {
		return err
	}

	_, err = doRequest(req, creds)
	if err != nil {
		return err
	}

	return nil
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

	_, err = doRequest(req, creds)
	if err != nil {
		return err
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
	res, err := doRequest(req, creds)

	if err != nil {
		return nil, 0, gitmedia.Error(err)
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		return nil, 0, gitmedia.Error(errors.New("Invalid Content-Type"))
	}

	if ok, err := validateMediaHeader(contentType, res.Body); !ok {
		return nil, 0, gitmedia.Error(err)
	}

	return res.Body, res.ContentLength, nil
}

func validateMediaHeader(contentType string, reader io.Reader) (bool, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false, errors.New("Invalid Media Type")
	}

	if mediaType != gitMediaType {
		return false, errors.New("Invalid Media Type")
	}

	givenHeader, ok := params["header"]
	if !ok {
		return false, errors.New("Invalid header")
	}

	fullGivenHeader := "--" + givenHeader + "\n"

	header := make([]byte, len(fullGivenHeader))
	_, err = io.ReadAtLeast(reader, header, len(fullGivenHeader))
	if err != nil {
		return false, err
	}

	if string(header) != fullGivenHeader {
		return false, errors.New("Invalid header")
	}

	return true, nil
}

func doRequest(req *http.Request, creds Creds) (*http.Response, error) {
	res, err := http.DefaultClient.Do(req)

	if err == nil {
		if res.StatusCode > 299 {
			execCreds(creds, "reject")

			apierr := &Error{}
			dec := json.NewDecoder(res.Body)
			if err := dec.Decode(apierr); err != nil {
				return res, err
			}

			return res, apierr
		}

		execCreds(creds, "approve")
	}

	return res, err
}

func clientRequest(method, oid string) (*http.Request, Creds, error) {
	u := ObjectUrl(oid)
	req, err := http.NewRequest(method, u.String(), nil)
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
	u.Path = filepath.Join(u.Path, "/objects/"+oid)
	return u
}

type Error struct {
	Message   string `json:"message"`
	RequestId string `json:"request_id,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}
