package api_test

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
	"github.com/github/git-lfs/auth"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errors"
	"github.com/github/git-lfs/httputil"
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

	mux.HandleFunc("/media/objects/batch", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "POST" {
			t.Error("bad http verb")
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != api.MediaType {
			t.Error("Invalid Accept")
			w.WriteHeader(405)
			return
		}

		auth := r.Header.Get("Authorization")
		if len(auth) == 0 {
			w.WriteHeader(401)
			return
		}

		if auth != expectedAuth(t, server) {
			t.Errorf("Invalid Authorization: %q", auth)
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

		by, err := json.Marshal(newBatchResponse("", obj))
		if err != nil {
			t.Fatal(err)
		}

		head := w.Header()
		head.Set("Content-Type", api.MediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.url": server.URL + "/media",
		},
	})

	obj, _, err := api.BatchSingle(cfg, &api.ObjectResource{Oid: "oid"}, "download", []string{"basic"})
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatalf("unexpected error: %s", err)
	}

	if obj.Size != 4 {
		t.Errorf("unexpected size: %d", obj.Size)
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

	mux.HandleFunc("/redirect/objects/batch", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		w.Header().Set("Location", server.URL+"/redirect2/objects/batch")
		w.WriteHeader(307)
		t.Log("redirect with 307")
	})

	mux.HandleFunc("/redirect2/objects/batch", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		w.Header().Set("Location", server.URL+"/media/objects/batch")
		w.WriteHeader(307)
		t.Log("redirect with 307")
	})

	mux.HandleFunc("/media/objects/batch", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != api.MediaType {
			t.Error("Invalid Accept")
			w.WriteHeader(406)
			return
		}

		auth := r.Header.Get("Authorization")
		if len(auth) == 0 {
			w.WriteHeader(401)
			return
		}

		if auth != expectedAuth(t, server) {
			t.Errorf("Invalid Authorization: %q", auth)
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

		by, err := json.Marshal(newBatchResponse("", obj))
		if err != nil {
			t.Fatal(err)
		}

		head := w.Header()
		head.Set("Content-Type", api.MediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.url": server.URL + "/redirect",
		},
	})

	obj, _, err := api.BatchSingle(cfg, &api.ObjectResource{Oid: "oid"}, "download", []string{"basic"})
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatalf("unexpected error for 307 status: %s", err)
	}

	if obj.Size != 4 {
		t.Errorf("unexpected size for 307 status: %d", obj.Size)
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

	mux.HandleFunc("/media/objects/batch", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		t.Logf("request header: %v", r.Header)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != api.MediaType {
			t.Error("Invalid Accept")
		}

		auth := r.Header.Get("Authorization")
		if len(auth) == 0 {
			w.WriteHeader(401)
			return
		}

		if auth != expectedAuth(t, server) {
			t.Errorf("Invalid Authorization: %q", auth)
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

		by, err := json.Marshal(newBatchResponse("", obj))
		if err != nil {
			t.Fatal(err)
		}

		head := w.Header()
		head.Set("Content-Type", "application/json; charset=utf-8")
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.url": server.URL + "/media",
		},
	})

	obj, _, err := api.BatchSingle(cfg, &api.ObjectResource{Oid: "oid"}, "download", []string{"basic"})
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatalf("unexpected error: %s", err)
	}

	if obj.Size != 4 {
		t.Errorf("unexpected size: %d", obj.Size)
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

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.batch": "false",
			"lfs.url":   server.URL + "/media",
		},
	})

	_, _, err := api.BatchSingle(cfg, &api.ObjectResource{Oid: "oid"}, "download", []string{"basic"})
	if err == nil {
		t.Fatal("no error?")
	}

	if errors.IsFatalError(err) {
		t.Fatal("should not panic")
	}

	if isDockerConnectionError(err) {
		return
	}

	expected := "batch response: " + fmt.Sprintf(httputil.GetDefaultError(404), server.URL+"/media/objects/batch")
	if err.Error() != expected {
		t.Fatalf("Expected: %s\nGot: %s", expected, err.Error())
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

var (
	TestCredentialsFunc auth.CredentialFunc
	origCredentialsFunc auth.CredentialFunc
)

func init() {
	TestCredentialsFunc = func(cfg *config.Configuration, input auth.Creds, subCommand string) (auth.Creds, error) {
		output := make(auth.Creds)
		for key, value := range input {
			output[key] = value
		}
		if _, ok := output["username"]; !ok {
			output["username"] = input["host"]
		}
		output["password"] = "monkey"
		return output, nil
	}
}

// Override the credentials func for testing
func SetupTestCredentialsFunc() {
	origCredentialsFunc = auth.SetCredentialsFunc(TestCredentialsFunc)
}

// Put the original credentials func back
func RestoreCredentialsFunc() {
	auth.SetCredentialsFunc(origCredentialsFunc)
}
