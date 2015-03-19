package lfs

import (
	"io/ioutil"
	"testing"
)

type putRequest struct {
	Oid  string
	Size int
}

func tempdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "git-lfs-test")
	if err != nil {
		t.Fatalf("Error getting temp dir: %s", err)
	}
	return dir
}
