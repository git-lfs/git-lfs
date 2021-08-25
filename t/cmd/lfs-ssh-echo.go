//go:build testtools
// +build testtools

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type sshResponse struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt time.Time         `json:"expires_at,omitempty"`
	ExpiresIn int               `json:"expires_in,omitempty"`
}

func shell() string {
	if runtime.GOOS == "windows" {
		return "bash"
	}
	return "sh"
}

func spawnCommand(command string) error {
	cmd := exec.Command(shell(), "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok {
		var ws syscall.WaitStatus
		ws, ok = e.ProcessState.Sys().(syscall.WaitStatus)
		if ok {
			os.Exit(ws.ExitStatus())
		}
	}
	return err
}

func main() {
	// expect args:
	//   lfs-ssh-echo -p PORT -- git@127.0.0.1 "git-lfs-authenticate REPO OPERATION"
	//   lfs-ssh-echo -p PORT -- git@127.0.0.1 "git-lfs-transfer REPO OPERATION"
	//   lfs-ssh-echo -- git@127.0.0.1 "git-lfs-transfer REPO OPERATION"
	//   lfs-ssh-echo git@127.0.0.1 "git-upload-pack REPO"
	//   lfs-ssh-echo git@127.0.0.1 "git-receive-pack REPO"
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "got %d args: %v", len(os.Args), os.Args)
		os.Exit(1)
	}

	offset := 4
	if os.Args[1] == "--" {
		offset = 2
	} else if os.Args[1] == "git@127.0.0.1" {
		offset = 1
	} else if os.Args[1] != "-p" {
		fmt.Fprintf(os.Stderr, "$1 expected \"-p\", got %q", os.Args[1])
		os.Exit(1)
	} else if os.Args[3] != "--" {
		fmt.Fprintf(os.Stderr, "$3 expected \"--\", got %q", os.Args[3])
		os.Exit(1)
	}

	if len(os.Args) < offset+2 {
		fmt.Fprintf(os.Stderr, "got %d args: %v", len(os.Args), os.Args)
		os.Exit(1)
	}

	if os.Args[offset] != "git@127.0.0.1" {
		fmt.Fprintf(os.Stderr, "$4 expected \"git@127.0.0.1\", got %q", os.Args[offset])
		os.Exit(1)
	}

	// just "git-lfs-(authenticate|transfer) REPO OPERATION" or "git-(upload|receive)-pack REPO"
	remoteCmd := strings.Split(os.Args[offset+1], " ")
	if len(remoteCmd) < 2 {
		fmt.Fprintf(os.Stderr, "bad command line: %s\nargs: %v", remoteCmd, os.Args)
		os.Exit(1)
	}

	if remoteCmd[0] == "git-lfs-transfer" || remoteCmd[0] == "git-upload-pack" || remoteCmd[0] == "git-receive-pack" {
		err := spawnCommand(os.Args[offset+1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running command %q: %v", remoteCmd[0], err)
			os.Exit(1)
		}
		return
	}

	repo := remoteCmd[1]

	r := &sshResponse{
		Href: fmt.Sprintf("http://127.0.0.1:%s/%s.git/info/lfs", os.Args[2], repo),
	}
	switch repo {
	case "/ssh-expired-absolute":
		r.ExpiresAt = time.Now().Add(-5 * time.Minute)
	case "/ssh-expired-relative":
		r.ExpiresIn = -5
	case "/ssh-expired-both":
		r.ExpiresAt = time.Now().Add(-5 * time.Minute)
		r.ExpiresIn = -5
	}

	json.NewEncoder(os.Stdout).Encode(r)
}
