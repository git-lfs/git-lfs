package api_test // prevent import cycles

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errors"
	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/test"
)

func TestExistingUpload(t *testing.T) {
	SetupTestCredentialsFunc()
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
		RestoreCredentialsFunc()
	}()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != api.MediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != api.MediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &api.ObjectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "988881adc9fc3655077dc2d4d757d480b5ea0e11" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		obj := &api.ObjectResource{
			Oid:  reqObj.Oid,
			Size: reqObj.Size,
			Actions: map[string]*api.LinkRelation{
				"upload": &api.LinkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &api.LinkRelation{
					Href:   server.URL + "/verify",
					Header: map[string]string{"B": "2"},
				},
			},
		}

		by, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}

		postCalled = true
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

	oidPath, _ := lfs.LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	oid := filepath.Base(oidPath)
	stat, _ := os.Stat(oidPath)
	o, _, err := api.BatchOrLegacySingle(cfg, &api.ObjectResource{Oid: oid, Size: stat.Size()}, "upload", []string{"basic"})
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatal(err)
	}
	if o != nil {
		t.Errorf("Got an object back")
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

}

func TestUploadWithRedirect(t *testing.T) {
	SetupTestCredentialsFunc()
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
		RestoreCredentialsFunc()
	}()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	mux.HandleFunc("/redirect/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		w.Header().Set("Location", server.URL+"/redirect2/objects")
		w.WriteHeader(307)
	})

	mux.HandleFunc("/redirect2/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		w.Header().Set("Location", server.URL+"/media/objects")
		w.WriteHeader(307)
	})

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != api.MediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != api.MediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &api.ObjectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "988881adc9fc3655077dc2d4d757d480b5ea0e11" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		obj := &api.ObjectResource{
			Actions: map[string]*api.LinkRelation{
				"upload": &api.LinkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &api.LinkRelation{
					Href:   server.URL + "/verify",
					Header: map[string]string{"B": "2"},
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

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.url": server.URL + "/redirect",
		},
	})

	oidPath, _ := lfs.LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	oid := filepath.Base(oidPath)
	stat, _ := os.Stat(oidPath)
	o, _, err := api.BatchOrLegacySingle(cfg, &api.ObjectResource{Oid: oid, Size: stat.Size()}, "upload", []string{"basic"})
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatal(err)
	}

	if o != nil {
		t.Fatal("Received an object")
	}
}

func TestSuccessfulUploadWithVerify(t *testing.T) {
	SetupTestCredentialsFunc()
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
		RestoreCredentialsFunc()
	}()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	verifyCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != api.MediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != api.MediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &api.ObjectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "988881adc9fc3655077dc2d4d757d480b5ea0e11" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		obj := &api.ObjectResource{
			Oid:  reqObj.Oid,
			Size: reqObj.Size,
			Actions: map[string]*api.LinkRelation{
				"upload": &api.LinkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &api.LinkRelation{
					Href:   server.URL + "/verify",
					Header: map[string]string{"B": "2"},
				},
			},
		}

		by, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}

		postCalled = true
		head := w.Header()
		head.Set("Content-Type", api.MediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(202)
		w.Write(by)
	})

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("B") != "2" {
			t.Error("Invalid B")
		}

		if r.Header.Get("Content-Type") != api.MediaType {
			t.Error("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &api.ObjectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "988881adc9fc3655077dc2d4d757d480b5ea0e11" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		verifyCalled = true
		w.WriteHeader(200)
	})

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.url": server.URL + "/media",
		},
	})

	oidPath, _ := lfs.LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	oid := filepath.Base(oidPath)
	stat, _ := os.Stat(oidPath)
	o, _, err := api.BatchOrLegacySingle(cfg, &api.ObjectResource{Oid: oid, Size: stat.Size()}, "upload", []string{"basic"})
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatal(err)
	}
	api.VerifyUpload(cfg, o)

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !verifyCalled {
		t.Errorf("verify not called")
	}

}

func TestUploadApiError(t *testing.T) {
	SetupTestCredentialsFunc()
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
		RestoreCredentialsFunc()
	}()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		postCalled = true
		w.WriteHeader(404)
	})

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.url": server.URL + "/media",
		},
	})

	oidPath, _ := lfs.LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	oid := filepath.Base(oidPath)
	stat, _ := os.Stat(oidPath)
	_, _, err := api.BatchOrLegacySingle(cfg, &api.ObjectResource{Oid: oid, Size: stat.Size()}, "upload", []string{"basic"})
	if err == nil {
		t.Fatal(err)
	}

	if errors.IsFatalError(err) {
		t.Fatal("should not panic")
	}

	if isDockerConnectionError(err) {
		return
	}

	expected := "LFS: " + fmt.Sprintf(httputil.GetDefaultError(404), server.URL+"/media/objects")
	if err.Error() != expected {
		t.Fatalf("Expected: %s\nGot: %s", expected, err.Error())
	}

	if !postCalled {
		t.Errorf("POST not called")
	}
}

func TestUploadVerifyError(t *testing.T) {
	SetupTestCredentialsFunc()
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
		RestoreCredentialsFunc()
	}()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	verifyCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != api.MediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != api.MediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &api.ObjectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "988881adc9fc3655077dc2d4d757d480b5ea0e11" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		obj := &api.ObjectResource{
			Oid:  reqObj.Oid,
			Size: reqObj.Size,
			Actions: map[string]*api.LinkRelation{
				"upload": &api.LinkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &api.LinkRelation{
					Href:   server.URL + "/verify",
					Header: map[string]string{"B": "2"},
				},
			},
		}

		by, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}

		postCalled = true
		head := w.Header()
		head.Set("Content-Type", api.MediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(202)
		w.Write(by)
	})

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		verifyCalled = true
		w.WriteHeader(404)
	})

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.url": server.URL + "/media",
		},
	})

	oidPath, _ := lfs.LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	oid := filepath.Base(oidPath)
	stat, _ := os.Stat(oidPath)
	o, _, err := api.BatchOrLegacySingle(cfg, &api.ObjectResource{Oid: oid, Size: stat.Size()}, "upload", []string{"basic"})
	if err != nil {
		if isDockerConnectionError(err) {
			return
		}
		t.Fatal(err)
	}
	err = api.VerifyUpload(cfg, o)
	if err == nil {
		t.Fatal("verify should fail")
	}

	if errors.IsFatalError(err) {
		t.Fatal("should not panic")
	}

	expected := fmt.Sprintf(httputil.GetDefaultError(404), server.URL+"/verify")
	if err.Error() != expected {
		t.Fatalf("Expected: %s\nGot: %s", expected, err.Error())
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !verifyCalled {
		t.Errorf("verify not called")
	}

}
