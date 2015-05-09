package tests

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
)

// SetRepo changes the current test to the repository's working directory.  The
// given name refers to the subdirectory in the runner temp directory that the
// repository lives.
func (r *runner) SetRepo(name string) {
	r.repoName = name
	if err := os.Chdir(r.repo().dir); err != nil {
		r.Fatal(err)
	}
}

// InitRepo initializes an empty git repository.  The given name is a
// subdirectory in the runner temp directory that the repository will live.
func (r *runner) InitRepo(name string) {
	dir := filepath.Join(r.dir, name)
	if err := os.MkdirAll(dir, 0777); err != nil {
		r.Fatal(err)
	}

	repo := &repo{
		dir:    dir,
		server: httptest.NewServer(r.gitHandler()),
	}
	r.repos[name] = repo

	r.SetRepo(name)
	r.Git("init")
	r.Git("config", "http.receivepack", "true")
	r.Git("remote", "add", "origin", repo.server.URL+"/"+name)
	r.Logf("git init: %s", dir)
}

func (r *runner) repo() *repo {
	repo := r.repos[r.repoName]
	if repo == nil {
		r.Fatalf("no repo found for %q", r.repoName)
	}
	return repo
}

func (r *runner) gitHandler() http.Handler {
	run := r
	t := r.T // testing.T
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			io.Copy(ioutil.Discard, r.Body)
			r.Body.Close()
		}()

		cmd := exec.Command("git", "http-backend")
		cmd.Env = []string{
			fmt.Sprintf("GIT_PROJECT_ROOT=%s", run.dir),
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
			t.Fatal(err)
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
	})
}

type repo struct {
	dir    string
	server *httptest.Server
}

func (r *repo) Teardown() {
	r.server.Close()
}
