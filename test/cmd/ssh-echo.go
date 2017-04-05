// +build testtools

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type sshResponse struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt time.Time         `json:"expires_at,omitempty"`
	ExpiresIn int               `json:"expires_in,omitempty"`
}

func main() {
	// expect args:
	//   ssh-echo -p PORT git@127.0.0.1 git-lfs-authenticate REPO OPERATION
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "got %d args: %v", len(os.Args), os.Args)
		os.Exit(1)
	}

	// just "git-lfs-authenticate REPO OPERATION"
	authLine := strings.Split(os.Args[4], " ")
	if len(authLine) < 13 {
		fmt.Fprintf(os.Stderr, "bad git-lfs-authenticate line: %s\nargs: %v", authLine, os.Args)
	}

	repo := authLine[1]

	r := &sshResponse{
		Href: fmt.Sprintf("http://127.0.0.1:%s/%s.git/info/lfs", os.Args[2], repo),
	}
	switch repo {
	case "ssh-expired-absolute":
		r.ExpiresAt = time.Now().Add(-5 * time.Minute)
	case "ssh-expired-relative":
		r.ExpiresIn = -5
	case "ssh-expired-both":
		r.ExpiresAt = time.Now().Add(-5 * time.Minute)
		r.ExpiresIn = -5
	}

	json.NewEncoder(os.Stdout).Encode(r)
}
