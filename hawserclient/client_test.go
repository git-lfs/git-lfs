package hawserclient

import (
	"github.com/bmizerany/assert"
	"github.com/hawser/git-hawser/hawser"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestOptions(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	defer os.RemoveAll(TmpPath)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(405)
		}
	})

	hawser.Config.SetConfig("hawser.url", server.URL+"/media")
	oidPath := filepath.Join(TmpPath, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatalf("Unable to write oid file: %s", err)
	}

	status, err := Options(oidPath)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if status != 200 {
		t.Errorf("unexpected status: %d", status)
	}
}

func TestOptionsWithoutExistingObject(t *testing.T) {
	status, err := Options("/this/better/not/work")
	if err == nil {
		t.Errorf("expected an error to be returned")
	}

	if status != 0 {
		t.Errorf("unexpected status: %d", status)
	}
}

func TestObjectUrl(t *testing.T) {
	oid := "oid"
	tests := map[string]string{
		"http://example.com":      "http://example.com/objects/oid",
		"http://example.com/":     "http://example.com/objects/oid",
		"http://example.com/foo":  "http://example.com/foo/objects/oid",
		"http://example.com/foo/": "http://example.com/foo/objects/oid",
	}

	config := hawser.Config
	for endpoint, expected := range tests {
		config.SetConfig("hawser.url", endpoint)
		assert.Equal(t, expected, ObjectUrl(oid).String())
	}
}

var TmpPath string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	TmpPath = filepath.Join(wd, "..", "tmp", "hawserclient")
	os.MkdirAll(TmpPath, 0755)

	execCreds = func(input Creds, subCommand string) (credentialFetcher, error) {
		return &testCredentialFetcher{input}, nil
	}
}

type testCredentialFetcher struct {
	Creds Creds
}

func (c *testCredentialFetcher) Credentials() Creds {
	return c.Creds
}
