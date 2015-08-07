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
	defer server.Close()

	tmp := tempdir(t)
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
			Actions: map[string]*linkRelation{
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
// called multiple times to return different 3xx status codes
func TestSuccessfulDownloadWithRedirects(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	tmp := tempdir(t)
	defer os.RemoveAll(tmp)

	// all of these should work for GET requests
	redirectCodes := []int{301, 302, 303, 307}
	redirectIndex := 0

	mux.HandleFunc("/redirect/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		w.Header().Set("Location", server.URL+"/redirect2/objects/oid")
		w.WriteHeader(redirectCodes[redirectIndex])
		t.Logf("redirect with %d", redirectCodes[redirectIndex])
	})

	mux.HandleFunc("/redirect2/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		w.Header().Set("Location", server.URL+"/media/objects/oid")
		w.WriteHeader(redirectCodes[redirectIndex])
		t.Logf("redirect again with %d", redirectCodes[redirectIndex])
		redirectIndex += 1
	})

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
			Actions: map[string]*linkRelation{
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

	Config.SetConfig("lfs.url", server.URL+"/redirect")

	for _, redirect := range redirectCodes {
		reader, size, wErr := Download("oid")
		if wErr != nil {
			t.Fatalf("unexpected error for %d status: %s", redirect, wErr)
		}

		if size != 4 {
			t.Errorf("unexpected size for %d status: %d", redirect, size)
		}

		by, err := ioutil.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatalf("unexpected error for %d status: %s", redirect, err)
		}

		if body := string(by); body != "test" {
			t.Errorf("unexpected body for %d status: %s", redirect, body)
		}
	}
}

// nearly identical to TestSuccessfulDownload
// the api request returns a custom Authorization header
func TestSuccessfulDownloadWithAuthorization(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	tmp := tempdir(t)
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
			Actions: map[string]*linkRelation{
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
		head.Set("Content-Type", "application/json; charset=utf-8")
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
	defer server.Close()

	mux2 := http.NewServeMux()
	server2 := httptest.NewServer(mux2)
	defer server2.Close()

	tmp := tempdir(t)
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
			Actions: map[string]*linkRelation{
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

// nearly identical to TestSuccessfulDownload
// download is served from a second server
func TestSuccessfulDownloadFromSeparateRedirectedHost(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux2 := http.NewServeMux()
	server2 := httptest.NewServer(mux2)
	defer server2.Close()

	mux3 := http.NewServeMux()
	server3 := httptest.NewServer(mux3)
	defer server3.Close()

	tmp := tempdir(t)
	defer os.RemoveAll(tmp)

	// all of these should work for GET requests
	redirectCodes := []int{301, 302, 303, 307}
	redirectIndex := 0

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server 1: %s %s", r.Method, r.URL)
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

		w.Header().Set("Location", server2.URL+"/media/objects/oid")
		w.WriteHeader(redirectCodes[redirectIndex])
		t.Logf("redirect with %d", redirectCodes[redirectIndex])
		redirectIndex += 1
	})

	mux2.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server 2: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != "" {
			t.Error("Invalid Authorization")
		}

		obj := &objectResource{
			Oid:  "oid",
			Size: 4,
			Actions: map[string]*linkRelation{
				"download": &linkRelation{
					Href:   server3.URL + "/download",
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

	mux3.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server 3: %s %s", r.Method, r.URL)
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

	for _, redirect := range redirectCodes {
		reader, size, wErr := Download("oid")
		if wErr != nil {
			t.Fatalf("unexpected error for %d status: %s", redirect, wErr)
		}

		if size != 4 {
			t.Errorf("unexpected size for %d status: %d", redirect, size)
		}

		by, err := ioutil.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatalf("unexpected error for %d status: %s", redirect, err)
		}

		if body := string(by); body != "test" {
			t.Errorf("unexpected body for %d status: %s", redirect, body)
		}
	}
}

func TestDownloadAPIError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	tmp := tempdir(t)
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
	defer server.Close()

	tmp := tempdir(t)
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
			Actions: map[string]*linkRelation{
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
