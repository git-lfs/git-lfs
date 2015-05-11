package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type lfsObject struct {
	Oid   string             `json:"oid,omitempty"`
	Size  int64              `json:"size,omitempty"`
	Links map[string]lfsLink `json:"_links,omitempty"`
}

type lfsLink struct {
	Href   string            `json:"href"`
	Header map[string]string `json:"header,omitempty"`
}

// handles any requests with "{name}.server.git/info/lfs" in the path
func (run *runner) lfsHandler(repository *repo, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
	if r.Method == "POST" {
		run.lfsPostHandler(repository, w, r)
	} else {
		run.lfsGetHandler(repository, w, r)
	}
}

func (run *runner) lfsPostHandler(repository *repo, w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	tee := io.TeeReader(r.Body, buf)
	obj := &lfsObject{}
	err := json.NewDecoder(tee).Decode(obj)
	io.Copy(ioutil.Discard, r.Body)
	r.Body.Close()

	run.Log("REQUEST")
	run.Logf(buf.String())
	run.Logf("OID: %s", obj.Oid)
	run.Logf("Size: %d", obj.Size)

	if err != nil {
		run.Fatal(err)
	}

	res := &lfsObject{
		Links: map[string]lfsLink{
			"upload": lfsLink{
				Href: repository.server.URL + "/storage/" + obj.Oid,
			},
		},
	}

	by, err := json.Marshal(res)
	if err != nil {
		run.Fatal(err)
	}

	run.Log("RESPONSE: 202")
	run.Log(string(by))

	w.WriteHeader(202)
	w.Write(by)
}

func (run *runner) lfsGetHandler(repository *repo, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
}

// handles any /storage/{oid} requests
func (run *runner) storageHandler(repository *repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		run.Logf("storage %s %s", r.Method, r.URL)
		switch r.Method {
		case "PUT":
			hash := sha256.New()
			buf := &bytes.Buffer{}
			io.Copy(io.MultiWriter(hash, buf), r.Body)
			oid := hex.EncodeToString(hash.Sum(nil))
			if !strings.HasSuffix(r.URL.Path, "/"+oid) {
				w.WriteHeader(403)
				return
			}

			repository.largeObjects[oid] = buf.Bytes()

		case "GET":
			w.WriteHeader(500)
		default:
			w.WriteHeader(500)
		}
	}
}
