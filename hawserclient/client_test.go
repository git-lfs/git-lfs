package hawserclient

import (
	"encoding/json"
	"github.com/bmizerany/assert"
	"github.com/hawser/git-hawser/hawser"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGet(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			head := w.Header()
			head.Set("Content-Type", "application/octet-stream")
			head.Set("Content-Length", "4")
			w.WriteHeader(200)
			w.Write([]byte("test"))
			return
		}

		w.WriteHeader(405)
	})

	hawser.Config.SetConfig("hawser.url", server.URL+"/media")
	reader, size, wErr := Get("whatever/oid")
	if wErr != nil {
		t.Fatalf("unexpected error: %s", wErr)
	}
	defer reader.Close()

	if size != 4 {
		t.Errorf("unexpected size: %d", size)
	}

	by, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if body := string(by); body != "test" {
		t.Errorf("unexpected body: %s", body)
	}
}

type postRequest struct {
	Oid  string
	Size int
}

func TestPost(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			postReq := &postRequest{}
			if err := json.NewDecoder(r.Body).Decode(postReq); err != nil {
				t.Errorf("Error parsing json: %s", err)
			}
			r.Body.Close()

			if postReq.Size != 4 {
				t.Errorf("Unexpected size: %d", postReq.Size)
			}

			if postReq.Oid != "oid" {
				t.Errorf("unexpected oid: %s", postReq.Oid)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			w.Write([]byte(`{
				"_links": {
					"abc": {
						"href": "def",
						"header": {
							"a": "1",
							"b": "2"
						}
					}
				}
			}`))
			return
		}

		w.WriteHeader(405)
	})

	hawser.Config.SetConfig("hawser.url", server.URL+"/media")
	oidPath := filepath.Join(tmp, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatalf("Unable to write oid file: %s", err)
	}

	link, status, err := Post(oidPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if status != 201 {
		t.Errorf("unexpected status: %d", status)
	}

	if link == nil || link.Links == nil {
		t.Error("expected a link object, got none")
	}

	for key, rel := range link.Links {
		t.Logf("%s: %v", key, rel)
	}

	if len(link.Links) != 1 {
		t.Error("wrong number of link relations")
	}

	linkRel, ok := link.Links["abc"]
	if !ok {
		t.Error("no 'abc' rel")
	}

	if linkRel.Href != "def" {
		t.Errorf("bad href: %s", linkRel.Href)
	}

	if linkRel.Header["a"] != "1" {
		t.Errorf("bad 'a': %s", linkRel.Header["a"])
	}

	if linkRel.Header["b"] != "2" {
		t.Errorf("bad 'b': %s", linkRel.Header["b"])
	}
}

func TestPut(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			by, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Error reading request body: %s", err)
			}

			r.Body.Close()
			if body := string(by); body != "test" {
				t.Errorf("Unexpected body: %s", body)
			}

			w.WriteHeader(200)
			return
		}

		w.WriteHeader(405)
	})

	hawser.Config.SetConfig("hawser.url", server.URL+"/media")
	oidPath := filepath.Join(tmp, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatalf("Unable to write oid file: %s", err)
	}

	if err := Put(oidPath, "", nil); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestOptions(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(405)
		}
	})

	hawser.Config.SetConfig("hawser.url", server.URL+"/media")
	oidPath := filepath.Join(tmp, "oid")
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

func init() {
	execCreds = func(input Creds, subCommand string) (credentialFetcher, error) {
		return &testCredentialFetcher{input}, nil
	}
}

func tempdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "hawser-test-hawserclient")
	if err != nil {
		t.Fatalf("Error getting temp dir: %s", err)
	}
	return dir
}

type testCredentialFetcher struct {
	Creds Creds
}

func (c *testCredentialFetcher) Credentials() Creds {
	return c.Creds
}
