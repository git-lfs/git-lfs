package lfs

import (
	"encoding/json"
	"fmt"
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
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
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
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != "" {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
		}

		if r.Header.Get("A") != "1" {
			t.Error("invalid A")
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

// nearly identical to TestSuccessfulDownload
// the api request returns a custom Authorization header
func TestSuccessfulDownloadWithAuthorization(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
		}

		obj := &objectResource{
			Oid:  "oid",
			Size: 4,
			Links: map[string]*linkRelation{
				"download": &linkRelation{
					Href: server.URL + "/download",
					Header: map[string]string{
						"A":             "1",
						"Authorization": "custom",
					},
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
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != "" {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != "custom" {
			t.Error("Invalid Authorization")
		}

		if r.Header.Get("A") != "1" {
			t.Error("invalid A")
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

// nearly identical to TestSuccessfulDownload
// download is served from a second server
func TestSuccessfulDownloadFromSeparateHost(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	mux2 := http.NewServeMux()
	server2 := httptest.NewServer(mux2)
	tmp := tempdir(t)
	defer server.Close()
	defer server2.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
		}

		obj := &objectResource{
			Oid:  "oid",
			Size: 4,
			Links: map[string]*linkRelation{
				"download": &linkRelation{
					Href:   server2.URL + "/download",
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

	mux2.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != "" {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != "" {
			t.Error("Invalid Authorization")
		}

		if r.Header.Get("A") != "1" {
			t.Error("invalid A")
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

func TestDownloadAPIError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")
	_, _, wErr := Download("oid")
	if wErr == nil {
		t.Fatal("no error?")
	}

	if wErr.Panic {
		t.Fatal("should not panic")
	}

	if wErr.Error() != fmt.Sprintf(defaultErrors[404], server.URL+"/media/objects/oid") {
		t.Fatalf("Unexpected error: %s", wErr.Error())
	}
}

func TestDownloadStorageError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
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
		w.WriteHeader(500)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")
	_, _, wErr := Download("oid")
	if wErr == nil {
		t.Fatal("no error?")
	}

	if !wErr.Panic {
		t.Fatal("should panic")
	}

	if wErr.Error() != fmt.Sprintf(defaultErrors[500], server.URL+"/download") {
		t.Fatalf("Unexpected error: %s", wErr.Error())
	}
}
