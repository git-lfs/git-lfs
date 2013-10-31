package gitmediaclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	u := objectUrl(oid)
	req, err := http.NewRequest(method, u.String(), nil)
	if err == nil {
		creds, err := credentials(u)
		if err != nil {
			return req, err
		}

		token := fmt.Sprintf("%s:%s", creds["username"], creds["password"])
		auth := "Basic " + base64.URLEncoding.EncodeToString([]byte(token))
		req.Header.Set("Authorization", auth)
	}

	return req, err
}

func objectUrl(oid string) *url.URL {
	u, _ := url.Parse("http://localhost:8080")
	u.Path = "/objects/" + oid
	return u
}

func credentials(u *url.URL) (map[string]string, error) {
	creds := make(map[string]string)

	credInput := fmt.Sprintf("protocol=%s\nhost=%s\n", u.Scheme, u.Host)
	buf := new(bytes.Buffer)

	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = bytes.NewBufferString(credInput)
	cmd.Stdout = buf

	err := cmd.Start()
	if err != nil {
		return creds, err
	}

	err = cmd.Wait()
	if err != nil {
		return creds, err
	}

	for _, line := range strings.Split(buf.String(), "\n") {
		pieces := strings.SplitN(line, "=", 2)
		if len(pieces) < 2 {
			continue
		}
		creds[pieces[0]] = pieces[1]
	}

	return creds, nil
}

type Error struct {
	Message   string `json:"message"`
	RequestId string `json:"request_id,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}
