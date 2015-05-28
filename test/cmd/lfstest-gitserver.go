package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"os/exec"
	"strings"
)

var (
	repoDir      string
	largeObjects = make(map[string][]byte)
	server       *httptest.Server
)

func main() {
	repoDir = os.Getenv("LFSTEST_DIR")

	mux := http.NewServeMux()
	server = httptest.NewServer(mux)
	stopch := make(chan bool)

	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		stopch <- true
	})

	mux.HandleFunc("/storage/", storageHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/info/lfs") {
			log.Printf("git lfs %s %s\n", r.Method, r.URL)
			lfsHandler(w, r)
			return
		}

		log.Printf("git http-backend %s %s\n", r.Method, r.URL)
		gitHandler(w, r)
	})

	urlname := os.Getenv("LFSTEST_URL")
	if len(urlname) == 0 {
		urlname = "lfstest-gitserver"
	}

	file, err := os.Create(urlname)
	if err != nil {
		log.Fatalln(err)
	}

	file.Write([]byte(server.URL))
	file.Close()
	log.Println(server.URL)

	defer func() {
		os.RemoveAll(urlname)
	}()

	<-stopch
	log.Println("git server done")
}

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
func lfsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
	switch r.Method {
	case "POST":
		if strings.HasSuffix(r.URL.String(), "batch") {
			lfsBatchHandler(w, r)
		} else {
			lfsPostHandler(w, r)
		}
	case "GET":
		lfsGetHandler(w, r)
	default:
		w.WriteHeader(405)
	}
}

func lfsPostHandler(w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	tee := io.TeeReader(r.Body, buf)
	obj := &lfsObject{}
	err := json.NewDecoder(tee).Decode(obj)
	io.Copy(ioutil.Discard, r.Body)
	r.Body.Close()

	log.Println("REQUEST")
	log.Println(buf.String())
	log.Printf("OID: %s\n", obj.Oid)
	log.Printf("Size: %d\n", obj.Size)

	if err != nil {
		log.Fatal(err)
	}

	res := &lfsObject{
		Oid:  obj.Oid,
		Size: obj.Size,
		Links: map[string]lfsLink{
			"upload": lfsLink{
				Href: server.URL + "/storage/" + obj.Oid,
			},
		},
	}

	by, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("RESPONSE: 202")
	log.Println(string(by))

	w.WriteHeader(202)
	w.Write(by)
}

func lfsGetHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	oid := parts[len(parts)-1]

	by, ok := largeObjects[oid]
	if !ok {
		w.WriteHeader(404)
		return
	}

	obj := &lfsObject{
		Oid:  oid,
		Size: int64(len(by)),
		Links: map[string]lfsLink{
			"download": lfsLink{
				Href: server.URL + "/storage/" + oid,
			},
		},
	}

	by, err := json.Marshal(obj)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("RESPONSE: 200")
	log.Println(string(by))

	w.WriteHeader(200)
	w.Write(by)
}

func lfsBatchHandler(w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	tee := io.TeeReader(r.Body, buf)
	var objs map[string][]lfsObject
	err := json.NewDecoder(tee).Decode(&objs)
	io.Copy(ioutil.Discard, r.Body)
	r.Body.Close()

	log.Println("REQUEST")
	log.Println(buf.String())

	if err != nil {
		log.Fatal(err)
	}

	res := []lfsObject{}
	for _, obj := range objs["objects"] {
		o := lfsObject{
			Oid:  obj.Oid,
			Size: obj.Size,
			Links: map[string]lfsLink{
				"upload": lfsLink{
					Href: server.URL + "/storage/" + obj.Oid,
				},
			},
		}

		res = append(res, o)
	}

	ores := map[string][]lfsObject{"objects": res}

	by, err := json.Marshal(ores)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("RESPONSE: 200")
	log.Println(string(by))

	w.WriteHeader(200)
	w.Write(by)
}

// handles any /storage/{oid} requests
func storageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("storage %s %s\n", r.Method, r.URL)
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

		largeObjects[oid] = buf.Bytes()

	case "GET":
		parts := strings.Split(r.URL.Path, "/")
		oid := parts[len(parts)-1]

		if by, ok := largeObjects[oid]; ok {
			w.Write(by)
			return
		}

		w.WriteHeader(404)
	default:
		w.WriteHeader(405)
	}
}

func gitHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	cmd := exec.Command("git", "http-backend")
	cmd.Env = []string{
		fmt.Sprintf("GIT_PROJECT_ROOT=%s", repoDir),
		fmt.Sprintf("GIT_HTTP_EXPORT_ALL="),
		fmt.Sprintf("PATH_INFO=%s", r.URL.Path),
		fmt.Sprintf("QUERY_STRING=%s", r.URL.RawQuery),
		fmt.Sprintf("REQUEST_METHOD=%s", r.Method),
		fmt.Sprintf("CONTENT_TYPE=%s", r.Header.Get("Content-Type")),
	}

	buffer := &bytes.Buffer{}
	cmd.Stdin = r.Body
	cmd.Stdout = buffer
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	text := textproto.NewReader(bufio.NewReader(buffer))

	code, _, _ := text.ReadCodeLine(-1)

	if code != 0 {
		w.WriteHeader(code)
	}

	headers, _ := text.ReadMIMEHeader()
	head := w.Header()
	for key, values := range headers {
		for _, value := range values {
			head.Add(key, value)
		}
	}

	io.Copy(w, text.R)
}
