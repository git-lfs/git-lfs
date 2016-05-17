package lfs_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"
	. "github.com/github/git-lfs/lfs"
)

func TestSuccessfulDownload(t *testing.T) {
	SetupTestCredentialsFunc()
	defer func() {
		RestoreCredentialsFunc()
	}()

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

		if r.Header.Get("Accept") != api.MediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
		}

		obj := &api.ObjectResource{
			Oid:  "oid",
			Size: 4,
			Actions: map[string]*api.LinkRelation{
				"download": &api.LinkRelation{
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
		head.Set("Content-Type", api.MediaType)
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

		if r.Header.Get("A") != "1" {
			t.Error("invalid A")
		}

		head := w.Header()
		head.Set("Content-Type", "application/octet-stream")
		head.Set("Content-Length", "4")
		w.WriteHeader(200)
		w.Write([]byte("test"))
	})

	defer config.Config.ResetConfig()
	config.Config.SetConfig("lfs.batch", "false")
	config.Config.SetConfig("lfs.url", server.URL+"/media")

	reader, size, err := Download("oid", 0)
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatalf("unexpected error: %s", err)
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
	SetupTestCredentialsFunc()
	defer func() {
		RestoreCredentialsFunc()
	}()

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

		if r.Header.Get("Accept") != api.MediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
		}

		obj := &api.ObjectResource{
			Oid:  "oid",
			Size: 4,
			Actions: map[string]*api.LinkRelation{
				"download": &api.LinkRelation{
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
		head.Set("Content-Type", api.MediaType)
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

		if r.Header.Get("A") != "1" {
			t.Error("invalid A")
		}

		head := w.Header()
		head.Set("Content-Type", "application/octet-stream")
		head.Set("Content-Length", "4")
		w.WriteHeader(200)
		w.Write([]byte("test"))
	})

	defer config.Config.ResetConfig()
	config.Config.SetConfig("lfs.batch", "false")
	config.Config.SetConfig("lfs.url", server.URL+"/redirect")

	for _, redirect := range redirectCodes {
		reader, size, err := Download("oid", 0)
		if err != nil {
			if isDockerConnectionError(err) {
				return
			}
			t.Fatalf("unexpected error for %d status: %s", redirect, err)
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
	SetupTestCredentialsFunc()
	defer func() {
		RestoreCredentialsFunc()
	}()

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

		if r.Header.Get("Accept") != api.MediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
		}

		obj := &api.ObjectResource{
			Oid:  "oid",
			Size: 4,
			Actions: map[string]*api.LinkRelation{
				"download": &api.LinkRelation{
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

	defer config.Config.ResetConfig()
	config.Config.SetConfig("lfs.batch", "false")
	config.Config.SetConfig("lfs.url", server.URL+"/media")
	reader, size, err := Download("oid", 0)
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatalf("unexpected error: %s", err)
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
	SetupTestCredentialsFunc()
	defer func() {
		RestoreCredentialsFunc()
	}()

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

		if r.Header.Get("Accept") != api.MediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
		}

		obj := &api.ObjectResource{
			Oid:  "oid",
			Size: 4,
			Actions: map[string]*api.LinkRelation{
				"download": &api.LinkRelation{
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
		head.Set("Content-Type", api.MediaType)
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

		if r.Header.Get("A") != "1" {
			t.Error("invalid A")
		}

		head := w.Header()
		head.Set("Content-Type", "application/octet-stream")
		head.Set("Content-Length", "4")
		w.WriteHeader(200)
		w.Write([]byte("test"))
	})

	defer config.Config.ResetConfig()
	config.Config.SetConfig("lfs.batch", "false")
	config.Config.SetConfig("lfs.url", server.URL+"/media")
	reader, size, err := Download("oid", 0)
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatalf("unexpected error: %s", err)
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
	SetupTestCredentialsFunc()
	defer func() {
		RestoreCredentialsFunc()
	}()

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

		if r.Header.Get("Accept") != api.MediaType {
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

		if r.Header.Get("Accept") != api.MediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != "" {
			t.Error("Invalid Authorization")
		}

		obj := &api.ObjectResource{
			Oid:  "oid",
			Size: 4,
			Actions: map[string]*api.LinkRelation{
				"download": &api.LinkRelation{
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
		head.Set("Content-Type", api.MediaType)
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

		if r.Header.Get("A") != "1" {
			t.Error("invalid A")
		}

		head := w.Header()
		head.Set("Content-Type", "application/octet-stream")
		head.Set("Content-Length", "4")
		w.WriteHeader(200)
		w.Write([]byte("test"))
	})

	defer config.Config.ResetConfig()
	config.Config.SetConfig("lfs.batch", "false")
	config.Config.SetConfig("lfs.url", server.URL+"/media")

	for _, redirect := range redirectCodes {
		reader, size, err := Download("oid", 0)
		if err != nil {
			if isDockerConnectionError(err) {
				return
			}
			t.Fatalf("unexpected error for %d status: %s", redirect, err)
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
	SetupTestCredentialsFunc()
	defer func() {
		RestoreCredentialsFunc()
	}()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	tmp := tempdir(t)
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/media/objects/oid", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	defer config.Config.ResetConfig()
	config.Config.SetConfig("lfs.batch", "false")
	config.Config.SetConfig("lfs.url", server.URL+"/media")
	_, _, err := Download("oid", 0)
	if err == nil {
		t.Fatal("no error?")
	}

	if errutil.IsFatalError(err) {
		t.Fatal("should not panic")
	}

	if isDockerConnectionError(err) {
		return
	}

	if err.Error() != fmt.Sprintf(httputil.GetDefaultError(404), server.URL+"/media/objects/oid") {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

}

func TestDownloadStorageError(t *testing.T) {
	SetupTestCredentialsFunc()
	defer func() {
		RestoreCredentialsFunc()
	}()

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

		if r.Header.Get("Accept") != api.MediaType {
			t.Error("Invalid Accept")
		}

		if r.Header.Get("Authorization") != expectedAuth(t, server) {
			t.Error("Invalid Authorization")
		}

		obj := &api.ObjectResource{
			Oid:  "oid",
			Size: 4,
			Actions: map[string]*api.LinkRelation{
				"download": &api.LinkRelation{
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
		head.Set("Content-Type", api.MediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	defer config.Config.ResetConfig()
	config.Config.SetConfig("lfs.batch", "false")
	config.Config.SetConfig("lfs.url", server.URL+"/media")
	_, _, err := Download("oid", 0)
	if err == nil {
		t.Fatal("no error?")
	}

	if isDockerConnectionError(err) {
		return
	}

	if !errutil.IsFatalError(err) {
		t.Fatal("should panic")
	}

	if err.Error() != fmt.Sprintf(httputil.GetDefaultError(500), server.URL+"/download") {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}

// guards against connection errors that only seem to happen on debian docker
// images.
func isDockerConnectionError(err error) bool {
	if err == nil {
		return false
	}

	if os.Getenv("TRAVIS") == "true" {
		return false
	}

	e := err.Error()
	return strings.Contains(e, "connection reset by peer") ||
		strings.Contains(e, "connection refused")
}

func tempdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "git-lfs-test")
	if err != nil {
		t.Fatalf("Error getting temp dir: %s", err)
	}
	return dir
}

func expectedAuth(t *testing.T, server *httptest.Server) string {
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	token := fmt.Sprintf("%s:%s", u.Host, "monkey")
	return "Basic " + strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte(token)))
}
