package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	serverDir := dir + ".server"
	if err := os.MkdirAll(serverDir, 0777); err != nil {
		r.Fatal(err)
	}

	repo := &repo{
		dir:          dir,
		largeObjects: make(map[string][]byte),
		serverDir:    serverDir,
	}
	repo.server = httptest.NewServer(r.httpHandler(repo))
	r.repos[name] = repo

	// set up the git server
	if err := os.Chdir(serverDir); err != nil {
		r.Fatal(err)
	}
	r.Git("init")
	r.Git("config", "http.receivepack", "true")
	r.Git("config", "receive.denyCurrentBranch", "ignore")
	r.Logf("git init server: %s/%s", repo.server.URL, name)

	// set up the local git clone
	r.SetRepo(name)
	r.Git("init")
	r.Git("remote", "add", "origin", repo.server.URL+"/"+name+".server")
	r.setupCredentials(repo.server.URL)
	r.WriteFile(".git/hooks/pre-push", []byte("#!/bin/sh\n"+bin+` pre-push "$@"`+"\n"))
	r.Logf("git init: %s", dir)
}

func (r *runner) repo() *repo {
	repo := r.repos[r.repoName]
	if repo == nil {
		r.Fatalf("no repo found for %q", r.repoName)
	}
	return repo
}

func (r *runner) setupCredentials(rawurl string) {
	u, err := url.Parse(rawurl)
	if err != nil {
		r.Fatal(err)
	}

	input := fmt.Sprintf("protocol=http\nhost=%s\nusername=a\npassword=b", u.Host)
	out := &bytes.Buffer{}
	cmd := exec.Command("git", "credential", "approve")
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		r.Fatalf("%s\n\n%s", err, out.String())
	}
}

func (run *runner) httpHandler(repository *repo) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/storage/", run.storageHandler(repository))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, ".server.git/info/lfs") {
			run.Logf("git lfs %s %s", r.Method, r.URL)
			run.Logf("git lfs Accept: %s", r.Header.Get("Accept"))
			run.lfsHandler(repository, w, r)
			return
		}

		run.Logf("git http-backend %s %s", r.Method, r.URL)
		run.gitHandler(w, r)
	})

	return mux
}

type repo struct {
	dir          string
	largeObjects map[string][]byte
	server       *httptest.Server
	serverDir    string
}

func (r *repo) Teardown() {
	r.server.Close()
}
