//go:build testtools
// +build testtools

package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/shlex"
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

func checkSufficientArgs(offset int) {
	if len(os.Args) < offset+2 {
		fmt.Fprintf(os.Stderr, "got %d args: %v", len(os.Args), os.Args)
		os.Exit(1)
	}
}

func main() {
	// expect args:
	//   lfs-ssh-echo [-v[v[v]]] [-p PORT [--]] git@127.0.0.1[.invalid] "git-lfs-authenticate REPO OPERATION"
	//   lfs-ssh-echo [-v[v[v]]] [-p PORT [--]] git@127.0.0.1[.invalid] "git-lfs-transfer REPO OPERATION"
	//   lfs-ssh-echo [-v[v[v]]] git@127.0.0.1 "git-upload-pack REPO"
	//   lfs-ssh-echo [-v[v[v]]] git@127.0.0.1 "git-receive-pack REPO"

	offset := 1

	checkSufficientArgs(offset)
	switch os.Args[offset] {
	case "-v", "-vv", "-vvv":
		offset += 1
	}

	checkSufficientArgs(offset)
	if masterArg, found := strings.CutPrefix(os.Args[offset], "-oControlMaster="); found {
		var master bool
		switch masterArg {
		case "yes":
			master = true
		case "no":
			master = false
		default:
			fmt.Fprintf(os.Stderr, "expected \"-oControlMaster=yes\" or \"-oControlMaster=no\", got %q", os.Args[offset])
			os.Exit(1)
		}
		if pathArg, found := strings.CutPrefix(os.Args[offset+1], "-oControlPath="); found {
			if master {
				if file, err := os.OpenFile(pathArg, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666); err != nil {
					if os.IsExist(err) {
						fmt.Fprintf(os.Stderr, "expected %q to not exist", pathArg)
					} else {
						fmt.Fprintf(os.Stderr, "unable to create %q: %s", pathArg, err)
					}
					os.Exit(1)
				} else {
					file.Close()
					defer os.Remove(pathArg)
				}
			} else {
				if file, err := os.OpenFile(pathArg, os.O_RDONLY, 0); err != nil {
					if os.IsNotExist(err) {
						fmt.Fprintf(os.Stderr, "expected %q to exist", pathArg)
					} else {
						fmt.Fprintf(os.Stderr, "unable to open %q: %s", pathArg, err)
					}
					os.Exit(1)
				} else {
					file.Close()
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "expected \"-oControlPath\"")
			os.Exit(1)
		}
		offset += 2
	}

	checkSufficientArgs(offset)
	var port string
	if os.Args[offset] == "-p" {
		p, err := strconv.Atoi(os.Args[offset+1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to parse port %q: %v", os.Args[offset+1], err)
			os.Exit(1)
		}
		if p < 1 {
			fmt.Fprintf(os.Stderr, "invalid port: %d", p)
			os.Exit(1)
		}
		port = fmt.Sprintf(":%d", p)
		offset += 2
	}

	checkSufficientArgs(offset)
	if os.Args[offset] == "--" {
		offset += 1
	}

	checkSufficientArgs(offset)
	if os.Args[offset] != "git@127.0.0.1" && os.Args[offset] != "git@127.0.0.1.invalid" {
		fmt.Fprintf(os.Stderr, "expected \"git@127.0.0.1\" or \"git@127.0.0.1.invalid\", got %q", os.Args[offset])
		os.Exit(1)
	}

	// Should be either "git-lfs-(authenticate|transfer) 'REPO' OPERATION"
	// or "git-(upload|receive)-pack 'REPO'".
	remoteCmd, err := shlex.Split(os.Args[offset+1])
	if len(remoteCmd) < 2 || err != nil {
		fmt.Fprintf(os.Stderr, "bad command line: %s\nargs: %v", remoteCmd, os.Args)
		os.Exit(1)
	}

	// If "git-lfs-(authenticate|transfer) REPO; touch exploit OPERATION"
	// is supplied (or an equivalent with another shell command delimiter),
	// then simulate the "touch" command and return an error.
	//
	// This condition should not occur unless the Git LFS client fails to
	// properly quote the REPO argument.  In this case we skip trying to
	// execute "git-lfs-authenticate" or "git-lfs-transfer" since the
	// latter would fail due to a missing final argument, and we want to
	// avoid inadvertently running an arbitrary program named
	// "git-lfs-authenticate" if one happens to exist.
	if len(remoteCmd) > 3 &&
		(remoteCmd[0] == "git-lfs-authenticate" || remoteCmd[0] == "git-lfs-transfer") &&
		remoteCmd[2] == "touch" && remoteCmd[3] == "exploit" {
		if file, err := os.OpenFile("exploit", os.O_WRONLY|os.O_CREATE, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error simulating command %q: %v", os.Args[offset+1], err)
		} else {
			file.Close()
		}
		fmt.Fprintf(os.Stderr, "simulated command %q", os.Args[offset+1])
		os.Exit(0)
	}

	if remoteCmd[0] == "git-lfs-transfer" || remoteCmd[0] == "git-upload-pack" || remoteCmd[0] == "git-receive-pack" {
		err := spawnCommand(os.Args[offset+1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running command %q: %v", os.Args[offset+1], err)
			os.Exit(1)
		}
		return
	}

	repo := strings.TrimLeft(remoteCmd[1], "/")

	// We avoid using the url.PathEscape() function because it URL-encodes
	// path separator ("/") characters.
	urlRepo := (&url.URL{Path: repo}).EscapedPath()

	switch repo {
	case "ssh-unavailable", "ssh-unavailable\n-newline":
		os.Exit(127)
	case "ssh-unavailable-message":
		// Simulate the error message returned by a third-party utility
		// which does not support the "git-lfs-authenticate" command.
		fmt.Fprintf(os.Stderr, "fatal: Gerrit Code Review: git-lfs-authenticate: not found")
		os.Exit(1)
	case "ssh-invalid-url":
		urlRepo = repo + "\n-newline"
	}

	r := &sshResponse{
		Href: fmt.Sprintf("http://127.0.0.1%s/%s.git/info/lfs", port, urlRepo),
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
