// +build testtools

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	commands = map[string]func(){
		"get":   fill,
		"store": log,
		"erase": log,
	}

	delim    = '\n'
	credsDir = ""
)

func init() {
	if len(credsDir) == 0 {
		credsDir = os.Getenv("CREDSDIR")
	}
}

func main() {
	if argsize := len(os.Args); argsize != 2 {
		fmt.Fprintf(os.Stderr, "wrong number of args: %d\n", argsize)
		os.Exit(1)
	}

	arg := os.Args[1]
	cmd := commands[arg]

	if cmd == nil {
		fmt.Fprintf(os.Stderr, "bad cmd: %s\n", arg)
		os.Exit(1)
	}

	cmd()
}

func fill() {
	scanner := bufio.NewScanner(os.Stdin)
	creds := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "bad line: %s\n", line)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "CREDS RECV: %s\n", line)
		creds[parts[0]] = strings.TrimSpace(parts[1])
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading standard input: %v", err)
		os.Exit(1)
	}

	hostPieces := strings.SplitN(creds["host"], ":", 2)
	user, pass, err := credsForHostAndPath(hostPieces[0], creds["path"])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if user != "skip" {
		if _, ok := creds["username"]; !ok {
			creds["username"] = user
		}

		if _, ok := creds["password"]; !ok {
			creds["password"] = pass
		}
	}

	for key, value := range creds {
		fmt.Fprintf(os.Stderr, "CREDS SEND: %s=%s\n", key, value)
		fmt.Fprintf(os.Stdout, "%s=%s\n", key, value)
	}
}

func credsForHostAndPath(host, path string) (string, string, error) {
	var hostFilename string

	// We need hostFilename to end in a slash so that our credentials all
	// end up in the same directory.  credsDir will come in from the
	// testsuite with a slash, but filepath.Join will strip it off if host
	// is empty, such as when we have a file:/// or cert:/// URL.
	if host != "" {
		hostFilename = filepath.Join(credsDir, host)
	} else {
		hostFilename = credsDir
	}

	if len(path) > 0 {
		pathFilename := fmt.Sprintf("%s--%s", hostFilename, strings.Replace(path, "/", "-", -1))
		u, p, err := credsFromFilename(pathFilename)
		if err == nil {
			return u, p, err
		}
	}

	return credsFromFilename(hostFilename)
}

func credsFromFilename(file string) (string, string, error) {
	userPass, err := ioutil.ReadFile(file)
	if err != nil {
		return "", "", fmt.Errorf("Error opening %q: %s", file, err)
	}
	credsPieces := strings.SplitN(strings.TrimSpace(string(userPass)), ":", 2)
	if len(credsPieces) != 2 {
		return "", "", fmt.Errorf("Invalid data %q while reading %q", string(userPass), file)
	}
	return credsPieces[0], credsPieces[1], nil
}

func log() {
	fmt.Fprintf(os.Stderr, "CREDS received command: %s (ignored)\n", os.Args[1])
}
