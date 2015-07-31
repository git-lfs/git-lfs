package lfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestExistingUpload(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	olddir := LocalMediaDir
	LocalMediaDir = tmp
	defer func() {
		LocalMediaDir = olddir
	}()
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	putCalled := false
	verifyCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
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

		obj := &objectResource{
			Oid:  reqObj.Oid,
			Size: reqObj.Size,
			Actions: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &linkRelation{
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
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		putCalled = true
		w.WriteHeader(200)
	})

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)
		verifyCalled = true
		w.WriteHeader(200)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath, _ := LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	o, wErr := UploadCheck(oidPath)
	if wErr != nil {
		t.Fatal(wErr)
	}
	if o != nil {
		t.Errorf("Got an object back")
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if putCalled {
		t.Errorf("PUT not skipped")
	}

	if verifyCalled {
		t.Errorf("verify not skipped")
	}
}

func TestUploadWithRedirect(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	olddir := LocalMediaDir
	LocalMediaDir = tmp
	defer func() {
		LocalMediaDir = olddir
	}()
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

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
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

		obj := &objectResource{
			Actions: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &linkRelation{
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
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	Config.SetConfig("lfs.url", server.URL+"/redirect")

	oidPath, _ := LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	obj, wErr := UploadCheck(oidPath)
	if wErr != nil {
		t.Fatal(wErr)
	}

	if obj != nil {
		t.Fatal("Received an object")
	}
}

func TestSuccessfulUploadWithVerify(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	olddir := LocalMediaDir
	LocalMediaDir = tmp
	defer func() {
		LocalMediaDir = olddir
	}()
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	putCalled := false
	verifyCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
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

		obj := &objectResource{
			Oid:  reqObj.Oid,
			Size: reqObj.Size,
			Actions: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &linkRelation{
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
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(202)
		w.Write(by)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "PUT" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("A") != "1" {
			t.Error("Invalid A")
		}

		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Error("Invalid Content-Type")
		}

		if r.Header.Get("Content-Length") != "4" {
			t.Error("Invalid Content-Length")
		}

		if r.Header.Get("Transfer-Encoding") != "" {
			t.Fatal("Transfer-Encoding is set")
		}

		by, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", string(by))

		if str := string(by); str != "test" {
			t.Errorf("unexpected body: %s", str)
		}

		putCalled = true
		w.WriteHeader(200)
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

		if r.Header.Get("Content-Type") != mediaType {
			t.Error("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
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

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath, _ := LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	// stores callbacks
	calls := make([][]int64, 0, 5)
	cb := func(total int64, written int64, current int) error {
		calls = append(calls, []int64{total, written})
		return nil
	}

	obj, wErr := UploadCheck(oidPath)
	if wErr != nil {
		t.Fatal(wErr)
	}
	wErr = UploadObject(obj, cb)
	if wErr != nil {
		t.Fatal(wErr)
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !putCalled {
		t.Errorf("PUT not called")
	}

	if !verifyCalled {
		t.Errorf("verify not called")
	}

	t.Logf("CopyCallback: %v", calls)

	if len(calls) < 1 {
		t.Errorf("CopyCallback was not used")
	}

	lastCall := calls[len(calls)-1]
	if lastCall[0] != 4 || lastCall[1] != 4 {
		t.Errorf("Last CopyCallback call should be the total")
	}
}

func TestSuccessfulUploadWithoutVerify(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	olddir := LocalMediaDir
	LocalMediaDir = tmp
	defer func() {
		LocalMediaDir = olddir
	}()
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	putCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
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

		obj := &objectResource{
			Oid:  reqObj.Oid,
			Size: reqObj.Size,
			Actions: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
			},
		}

		by, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}

		postCalled = true
		head := w.Header()
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(202)
		w.Write(by)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "PUT" {
			w.WriteHeader(405)
			return
		}

		if a := r.Header.Get("A"); a != "1" {
			t.Errorf("Invalid A: %s", a)
		}

		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Error("Invalid Content-Type")
		}

		if r.Header.Get("Content-Length") != "4" {
			t.Error("Invalid Content-Length")
		}

		if r.Header.Get("Transfer-Encoding") != "" {
			t.Fatal("Transfer-Encoding is set")
		}

		by, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", string(by))

		if str := string(by); str != "test" {
			t.Errorf("unexpected body: %s", str)
		}

		putCalled = true
		w.WriteHeader(200)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath, _ := LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	obj, wErr := UploadCheck(oidPath)
	if wErr != nil {
		t.Fatal(wErr)
	}
	wErr = UploadObject(obj, nil)
	if wErr != nil {
		t.Fatal(wErr)
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !putCalled {
		t.Errorf("PUT not called")
	}
}

func TestUploadApiError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	olddir := LocalMediaDir
	LocalMediaDir = olddir
	defer func() {
		LocalMediaDir = olddir
	}()
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		postCalled = true
		w.WriteHeader(404)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath, _ := LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	_, wErr := UploadCheck(oidPath)
	if wErr == nil {
		t.Fatal(wErr)
	}

	if wErr.Panic {
		t.Fatal("should not panic")
	}

	if wErr.Error() != fmt.Sprintf(defaultErrors[404], server.URL+"/media/objects") {
		t.Fatalf("Unexpected error: %s", wErr.Error())
	}

	if !postCalled {
		t.Errorf("POST not called")
	}
}

func TestUploadStorageError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	olddir := LocalMediaDir
	LocalMediaDir = tmp
	defer func() {
		LocalMediaDir = olddir
	}()
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	putCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
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

		obj := &objectResource{
			Oid:  reqObj.Oid,
			Size: reqObj.Size,
			Actions: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &linkRelation{
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
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(202)
		w.Write(by)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		putCalled = true
		w.WriteHeader(404)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath, _ := LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	obj, wErr := UploadCheck(oidPath)
	if wErr != nil {
		t.Fatal(wErr)
	}
	wErr = UploadObject(obj, nil)
	if wErr == nil {
		t.Fatal("Expected an error")
	}

	if wErr.Panic {
		t.Fatal("should not panic")
	}

	if wErr.Error() != fmt.Sprintf(defaultErrors[404], server.URL+"/upload") {
		t.Fatalf("Unexpected error: %s", wErr.Error())
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !putCalled {
		t.Errorf("PUT not called")
	}
}

func TestUploadVerifyError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	olddir := LocalMediaDir
	LocalMediaDir = tmp
	defer func() {
		LocalMediaDir = olddir
	}()
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	putCalled := false
	verifyCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
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

		obj := &objectResource{
			Oid:  reqObj.Oid,
			Size: reqObj.Size,
			Actions: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &linkRelation{
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
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(202)
		w.Write(by)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "PUT" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("A") != "1" {
			t.Error("Invalid A")
		}

		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Error("Invalid Content-Type")
		}

		by, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", string(by))

		if str := string(by); str != "test" {
			t.Errorf("unexpected body: %s", str)
		}

		putCalled = true
		w.WriteHeader(200)
	})

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		verifyCalled = true
		w.WriteHeader(404)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath, _ := LocalMediaPath("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	obj, wErr := UploadCheck(oidPath)
	if wErr != nil {
		t.Fatal(wErr)
	}
	wErr = UploadObject(obj, nil)
	if wErr == nil {
		t.Fatal("Expected an error")
	}

	if wErr.Panic {
		t.Fatal("should not panic")
	}

	if wErr.Error() != fmt.Sprintf(defaultErrors[404], server.URL+"/verify") {
		t.Fatalf("Unexpected error: %s", wErr.Error())
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !putCalled {
		t.Errorf("PUT not called")
	}

	if !verifyCalled {
		t.Errorf("verify not called")
	}
}
