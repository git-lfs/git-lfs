package tests

import (
	"os"
	"path/filepath"
)

func (r *runner) initRepo(path string) {
	dir := filepath.Join(r.tempDir, path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		r.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		r.Fatal(err)
	}

	r.Git("init")
	r.Logf("git init: %s", dir)
	r.repoDir = dir
}
