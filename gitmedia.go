package gitmedia

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const Version = "0.0.1"

var (
	MediaWarning       = []byte("# This is a placeholder for large media, please install GitHub git-media to retrieve content\n# It is also possible you did not have the media locally, run 'git media sync' to retrieve it\n")
	LargeSizeThreshold = 5 * 1024 * 1024
	TempDir            = filepath.Join(os.TempDir(), "git-media")
	LocalMediaDir      string
)

type LargeAsset struct {
	MediaType string `json:"media_type"`
	Size      int64  `json:"size"`
	MD5       string `json:"md5"`
	SHA1      string `json:"sha1"`
}

func TempFile() (*os.File, error) {
	return ioutil.TempFile(TempDir, "")
}

func SimpleExec(name string, arg ...string) string {
	output, err := exec.Command(name, arg...).Output()
	if _, ok := err.(*exec.ExitError); ok {
		return ""
	} else if err != nil {
		fmt.Printf("error running: %s %s\n", name, arg)
		panic(err)
	}

	return strings.Trim(string(output), " \n")
}

func LocalMediaPath(sha string) string {
	path := filepath.Join(LocalMediaDir, sha[0:2], sha[2:4])
	if err := os.MkdirAll(path, 0744); err != nil {
		fmt.Printf("Error trying to create local media directory: %s\n", path)
		panic(err)
	}

	return filepath.Join(path, sha)
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	LocalMediaDir = filepath.Join(wd, ".git", "media")

	if err = os.MkdirAll(TempDir, 0744); err != nil {
		fmt.Printf("Error trying to create temp directory: %s\n", TempDir)
		panic(err)
	}
}
