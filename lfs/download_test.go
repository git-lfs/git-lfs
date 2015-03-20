package lfs

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestSuccessfulDownload(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %s", r.Method)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if accept := r.Header.Get("Accept"); accept != mediaType {
			t.Errorf("Invalid Accept: %s", accept)
		}

		obj := &objectResource{
			Oid:  "oid",
			Size: 4,
			Links: map[string]*linkRelation{
				"download": &linkRelation{
					Href:   server.URL + "/download",
					Header: map[string]string{"A": "1"},
				},
			},
		}

		by, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}

		head := w.Header()
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %s", r.Method)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if accept := r.Header.Get("Accept"); accept != "" {
			t.Errorf("Accept: %s", accept)
		}

		if a := r.Header.Get("A"); a != "1" {
			t.Logf("A: %s", a)
		}

		head := w.Header()
		head.Set("Content-Type", "application/octet-stream")
		head.Set("Content-Length", "4")
		w.WriteHeader(200)
		w.Write([]byte("test"))
	})

	Config.SetConfig("lfs.url", server.URL+"/media")
	reader, size, wErr := Download("oid")
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
