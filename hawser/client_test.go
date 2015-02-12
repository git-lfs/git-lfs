package hawser

import (
	"encoding/json"
	"github.com/bmizerany/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownload(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		head := w.Header()
		head.Set("Content-Type", "application/octet-stream")
		head.Set("Content-Length", "4")
		w.WriteHeader(200)
		w.Write([]byte("test"))
	})

	Config.SetConfig("hawser.url", server.URL+"/media")
	reader, size, wErr := Download(&DownloadRequest{"whatever/oid"})
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

type putRequest struct {
	Oid    string
	Size   int
	Status int
	Body   string
}

func TestExternalPut(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	Config.SetConfig("hawser.url", server.URL+"/media")
	oidPath := filepath.Join(tmp, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatalf("Unable to write oid file: %s", err)
	}

	uploaded := false
	verified := false

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			w.WriteHeader(405)
			return
		}

		if value := r.Header.Get("Content-Length"); value != "4" {
			t.Errorf("bad 'Content-Length' header: %v", value)
		}

		if value := r.Header.Get("a"); value != "1" {
			t.Errorf("bad 'a' header: %v", value)
		}

		by, err := ioutil.ReadAll(r.Body)
		r.Body.Close()

		if err != nil {
			t.Errorf("Error reading uploaded body: %s", err)
		}

		if string(by) != "test" {
			t.Errorf("bad body sent: %s", string(by))
		}

		uploaded = true
		w.WriteHeader(201)
		w.Write([]byte("yup"))
	})

	mux.HandleFunc("/media/objects/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if value := r.Header.Get("b"); value != "2" {
			t.Errorf("bad 'b' header: %v", value)
		}

		putReq := &putRequest{}
		if err := json.NewDecoder(r.Body).Decode(putReq); err != nil {
			t.Errorf("error decoding callback request json: %s", err)
		}

		if putReq.Oid != "oid" {
			t.Errorf("bad oid: %s", putReq.Oid)
		}

		if putReq.Size != 4 {
			t.Errorf("bad size: %s", putReq.Oid)
		}

		if putReq.Status != 201 {
			t.Errorf("bad status: %s", putReq.Oid)
		}

		if putReq.Body != "yup" {
			t.Errorf("bad body: %s", putReq.Oid)
		}

		verified = true
		w.WriteHeader(200)
	})

	link := &linkMeta{
		Links: map[string]*link{
			"upload": {
				Href:   server.URL + "/media/objects/oid",
				Header: map[string]string{"a": "1"},
			},
			"callback": {
				Href:   server.URL + "/media/objects/callback",
				Header: map[string]string{"b": "2"},
			},
		},
	}

	if err := ExternalPut(oidPath, "", link, nil); err != nil {
		t.Error(err)
	}

	if !uploaded {
		t.Error("upload request never called")
	}

	if !verified {
		t.Error("callback never called")
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
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

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
	})

	Config.SetConfig("hawser.url", server.URL+"/media")
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
		if r.Method != "PUT" {
			w.WriteHeader(405)
			return
		}

		by, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Error reading request body: %s", err)
		}

		r.Body.Close()
		if body := string(by); body != "test" {
			t.Errorf("Unexpected body: %s", body)
		}

		w.WriteHeader(200)
	})

	Config.SetConfig("hawser.url", server.URL+"/media")
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
		if r.Method != "OPTIONS" {
			w.WriteHeader(405)
			return
		}

		w.WriteHeader(200)
	})

	Config.SetConfig("hawser.url", server.URL+"/media")
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

	config := Config
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
	dir, err := ioutil.TempDir("", "hawser-test-hawser")
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
