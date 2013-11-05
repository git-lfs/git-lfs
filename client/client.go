package gitmediaclient

import (
	".."
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func Put(filename string) error {
	oid := filepath.Base(filename)
	stat, err := os.Stat(filename)
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	req, creds, err := clientRequest("PUT", oid)
	if err != nil {
		return err
	}

	req.Body = file
	req.ContentLength = stat.Size()

	res, err := doRequest(req, creds)
	if err != nil {
		return err
	}

	fmt.Printf("Sending %s from %s: %d\n", oid, filename, res.StatusCode)
	return nil
}

func Get(filename string) (io.ReadCloser, error) {
	oid := filepath.Base(filename)
	if stat, err := os.Stat(filename); err != nil || stat == nil {
		req, creds, err := clientRequest("GET", oid)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/vnd.git-media")
		res, err := doRequest(req, creds)

		if err != nil {
			return nil, err
		}

		return res.Body, nil
	}

	return os.Open(filename)
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
	} else {
		execCreds(creds, "approve")
	}

	return res, err
}

func clientRequest(method, oid string) (*http.Request, Creds, error) {
	u := objectUrl(oid)
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

func objectUrl(oid string) *url.URL {
	c := gitmedia.Config()
	u, _ := url.Parse(c.Endpoint)
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
