package gitmediaclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	req, err := clientRequest("PUT", oid)
	if err != nil {
		return err
	}
	req.Body = file
	req.ContentLength = stat.Size()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode > 299 {
		apierr := &Error{}
		dec := json.NewDecoder(res.Body)
		if err = dec.Decode(apierr); err != nil {
			return err
		}
		return apierr
	}

	fmt.Printf("Sending %s from %s: %d\n", oid, filename, res.StatusCode)
	return nil
}

func Get(filename string) (io.ReadCloser, error) {
	oid := filepath.Base(filename)
	if stat, err := os.Stat(filename); err != nil || stat == nil {
		req, err := clientRequest("GET", oid)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/vnd.git-media")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		return res.Body, nil
	}

	return os.Open(filename)
}

func clientRequest(method, oid string) (*http.Request, error) {
	return http.NewRequest(method, objectUrl(oid), nil)
}

func objectUrl(oid string) string {
	return "http://localhost:8080/objects/" + oid
}

type Error struct {
	Message   string `json:"message"`
	RequestId string `json:"request_id,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}
