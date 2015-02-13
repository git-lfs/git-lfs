package hawser

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type putRequest struct {
	Oid  string
	Size int
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

	mux.HandleFunc("/media/objects/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if value := r.Header.Get("b"); value != "2" {
			t.Errorf("bad 'b' header: %v", value)
		}

		putReq := &putRequest{}
		if err := json.NewDecoder(r.Body).Decode(putReq); err != nil {
			t.Errorf("error decoding verify request json: %s", err)
		}

		if putReq.Oid != "oid" {
			t.Errorf("bad oid: %s", putReq.Oid)
		}

		if putReq.Size != 4 {
			t.Errorf("bad size: %d", putReq.Size)
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
			"verify": {
				Href:   server.URL + "/media/objects/verify",
				Header: map[string]string{"b": "2"},
			},
		},
	}

	if err := callExternalPut(oidPath, "", link, nil); err != nil {
		t.Error(err)
	}

	if !uploaded {
		t.Error("upload request never called")
	}

	if !verified {
		t.Error("verify request never called")
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
		w.WriteHeader(202)
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

	link, status, err := callPost(oidPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if status != 202 {
		t.Errorf("unexpected status: %d", status)
	}

	if link == nil || link.Links == nil {
		t.Fatal("expected a link object, got none")
	}

	for key, rel := range link.Links {
		t.Logf("%s: %v", key, rel)
	}

	if len(link.Links) != 1 {
		t.Error("wrong number of link relations")
	}

	linkRel, ok := link.Links["abc"]
	if !ok {
		t.Fatal("no 'abc' rel")
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
