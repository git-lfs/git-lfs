package tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Git executes a Git command and returns the combined STDOUT and STDERR.
func (r *runner) Git(args ...string) string {
	name := "git"
	if args != nil && len(args) > 0 && args[0] == "lfs" {
		name = bin
		args = args[1:len(args)]
	}

	cmd := exec.Command(name, args...)
	return r.execCmd(cmd)
}

// GitBlob gets the blob OID of the given path at the given commit.
func (r *runner) GitBlob(commitish, path string) string {
	out := r.Git("ls-tree", commitish)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		tabs := strings.Split(line, "\t")
		if len(tabs) < 2 {
			continue
		}

		attrs := strings.Split(tabs[0], " ")
		if len(attrs) < 3 {
			continue
		}

		if tabs[1] == path {
			return attrs[2]
		}
	}

	return ""
}

// SetRepo changes the current test to the repository's working directory.  The
// given name refers to the subdirectory in the runner temp directory that the
// repository lives.
func (r *runner) SetRepo(name string) {
	r.Logf("$ cd %s", name)
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
		configFile:   filepath.Join(r.dir, name+".gitconfig"),
	}
	repo.server = httptest.NewServer(r.httpHandler(repo))
	repo.serverURL = repo.server.URL + "/" + name + ".server"
	r.repos[name] = repo

	// write a config file for any clones of this repo
	cfg := fmt.Sprintf(`[filter "lfs"]
	required = true
	smudge = %s smudge %%f
	clean = %s clean %%f

[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*
`, bin, bin, repo.serverURL)
	if err := ioutil.WriteFile(repo.configFile, []byte(cfg), 0755); err != nil {
		panic(err)
	}

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
	r.configRepo()
	r.setupCredentials(repo.server.URL)
	r.WriteFile(".git/hooks/pre-push", []byte("#!/bin/sh\n"+bin+` pre-push "$@"`+"\n"))
	r.Logf("git init: %s", dir)
}

func (r *runner) CloneTo(name string) string {
	repository := r.repo()
	if err := os.Chdir(r.dir); err != nil {
		r.Fatal(err)
	}

	cmdName := "git"
	args := []string{"clone", repository.serverURL, name}

	cmd := exec.Command(cmdName, args...)
	currEnv := os.Environ()
	cmd.Env = make([]string, len(currEnv)+1)
	for idx, e := range currEnv {
		cmd.Env[idx] = e
	}

	// ensures that filter.lfs.* config settings point to the test's git-lfs.
	cmd.Env[len(cmd.Env)-1] = "GIT_CONFIG=" + repository.configFile

	out := r.execCmd(cmd)

	r.repos[name] = &repo{
		runner:       r,
		dir:          filepath.Join(r.dir, name),
		largeObjects: repository.largeObjects,
		configFile:   repository.configFile,
		server:       repository.server,
		serverURL:    repository.serverURL,
	}

	r.SetRepo(name)
	r.configRepo()
	return out
}

func (r *runner) configRepo() {
	r.Git("config", "filter.lfs.smudge", fmt.Sprintf("%s smudge %%f", bin))
	r.Git("config", "filter.lfs.clean", fmt.Sprintf("%s clean %%f", bin))
	r.Git("config", "user.name", "Git LFS Tests")
	r.Git("config", "user.email", "example@git-lfs.com")
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
	cmd := exec.Command("git", "credential", "approve")
	cmd.Stdin = strings.NewReader(input)
	r.execCmd(cmd)
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
	runner       *runner
	dir          string
	largeObjects map[string][]byte
	configFile   string
	server       *httptest.Server
	serverDir    string
	serverURL    string
}

func (r *repo) Teardown() {
	r.server.Close()

	u, err := url.Parse(r.server.URL)
	if err != nil {
		r.runner.Fatal(err)
	}

	input := fmt.Sprintf("protocol=http\nhost=%s", u.Host)
	cmd := exec.Command("git", "credential", "reject")
	cmd.Stdin = strings.NewReader(input)
	cmd.Run()
}
