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

func TestUploadWithCallback(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	posted := false
	uploaded := false
	verified := false

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

		posted = true

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

		by, err := json.Marshal(link)
		if err != nil {
			t.Errorf("Error marshaling link json: %s", link)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(by)
	})

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

	Config.SetConfig("hawser.url", server.URL+"/media")
	oidPath := filepath.Join(tmp, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatalf("Unable to write oid file: %s", err)
	}

	err := Upload(oidPath, "", nil)
	if err != nil {
		t.Error(err)
	}

	if !posted {
		t.Error("preflight request never called")
	}

	if !uploaded {
		t.Error("upload request never called")
	}

	if !verified {
		t.Error("callback never called")
	}
}
