//go:build testtools
// +build testtools

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

type credential struct {
	authtype   string
	username   string
	password   string
	credential string
	matchState string
	state      string
	multistage bool
	skip       bool
}

func (c *credential) Serialize(capabilities map[string]struct{}, state []string, username []string) map[string][]string {
	formattedState := fmt.Sprintf("lfstest:%s", c.state)
	formattedMatchState := fmt.Sprintf("lfstest:%s", c.matchState)
	creds := make(map[string][]string)
	if c.skip {
		// Do nothing.
	} else if _, ok := capabilities["authtype"]; ok && len(c.authtype) != 0 && len(c.credential) != 0 {
		if _, ok := capabilities["state"]; len(c.matchState) == 0 || (ok && slices.Contains(state, formattedMatchState)) {
			creds["authtype"] = []string{c.authtype}
			creds["credential"] = []string{c.credential}
			if ok {
				creds["state[]"] = []string{formattedState}
				if c.multistage {
					creds["continue"] = []string{"1"}
				}
			}
		}
	} else if len(c.authtype) == 0 && (len(username) == 0 || username[0] == c.username) {
		if len(username) == 0 {
			creds["username"] = []string{c.username}
		}
		creds["password"] = []string{c.password}
	}
	return creds
}

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
	creds := map[string][]string{}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "bad line: %s\n", line)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "CREDS RECV: %s\n", line)
		if _, ok := creds[parts[0]]; ok {
			creds[parts[0]] = append(creds[parts[0]], strings.TrimSpace(parts[1]))
		} else {
			creds[parts[0]] = []string{strings.TrimSpace(parts[1])}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading standard input: %v", err)
		os.Exit(1)
	}

	hostPieces := strings.SplitN(firstEntryForKey(creds, "host"), ":", 2)
	credentials, err := credsForHostAndPath(hostPieces[0], firstEntryForKey(creds, "path"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	result := map[string][]string{}
	capas := discoverCapabilities(creds)
	for _, cred := range credentials {
		result = cred.Serialize(capas, creds["state[]"], creds["username"])
		if len(result) != 0 {
			break
		}
	}

	mode := os.Getenv("LFS_TEST_CREDS_WWWAUTH")
	wwwauth := firstEntryForKey(creds, "wwwauth[]")
	if mode == "required" && !strings.HasPrefix(wwwauth, "Basic ") {
		fmt.Fprintf(os.Stderr, "Missing required 'wwwauth[]' key in credentials\n")
		os.Exit(1)
	} else if mode == "forbidden" && wwwauth != "" {
		fmt.Fprintf(os.Stderr, "Unexpected 'wwwauth[]' key in credentials\n")
		os.Exit(1)
	}

	if len(result) == 0 {
		os.Exit(0)
	}

	// Send capabilities first to all for one-pass parsing, but only if
	// client advertised capabilities matching those of the per-host data.
	key := "capability[]"
	for entry, _ := range capas {
		fmt.Fprintf(os.Stderr, "CREDS SEND: %s=%s\n", key, entry)
		fmt.Fprintf(os.Stdout, "%s=%s\n", key, entry)
	}
	for key, value := range result {
		for _, entry := range value {
			fmt.Fprintf(os.Stderr, "CREDS SEND: %s=%s\n", key, entry)
			fmt.Fprintf(os.Stdout, "%s=%s\n", key, entry)
		}
	}
}

func discoverCapabilities(creds map[string][]string) map[string]struct{} {
	capas := make(map[string]struct{})
	supportedCapas := map[string]struct{}{
		"authtype": struct{}{},
		"state":    struct{}{},
	}
	for _, capa := range creds["capability[]"] {
		// Only pass on capabilities we support.
		if _, ok := supportedCapas[capa]; ok {
			capas[capa] = struct{}{}
		}
	}
	return capas
}

func credsForHostAndPath(host, path string) ([]credential, error) {
	if len(path) > 0 {
		pathFilename := fmt.Sprintf("%s--%s", host, strings.Replace(path, "/", "-", -1))
		cred, err := credsFromFilename(filepath.Join(credsDir, pathFilename))
		if err == nil {
			return cred, err
		}

		// Ideally we might run cygpath to convert paths like D:/...
		// to /d/... paths, but we only need to do this to support
		// one test of the deprecated git-lfs-clone command in our
		// CI suite, so for simplicity we just do basic rewriting.
		if len(path) > 2 && path[0] >= 'A' && path[0] <= 'Z' && path[1] == ':' {
			path = "/" + strings.ToLower(string(path[0])) + path[2:]
			pathFilename := fmt.Sprintf("%s--%s", host, strings.Replace(path, "/", "-", -1))
			cred, err := credsFromFilename(filepath.Join(credsDir, pathFilename))
			if err == nil {
				return cred, err
			}
		}
	}

	if len(host) == 0 {
		return nil, errors.New("No file available; empty 'host' key in credentials")
	}

	return credsFromFilename(filepath.Join(credsDir, host))
}

func parseOneCredential(s, file string) (credential, error) {
	// Each line in a file is of the following form:
	//
	// skip::
	//	The literal word "skip" means to skip emitting credentials.
	// AUTHTYPE::CREDENTIAL
	//	If the authtype is not empty, then this is an authtype and
	//	credential.
	// AUTHTYPE::CREDENTIAL:MATCH:STATE:MULTISTAGE
	//	Like above, but this matches only if MATCH is empty or if the
	//	state[] entry is present and matches "lfstest:MATCH".  If so,
	//	the value "lfstest:STATE" is emitted as the new state[] entry.
	//	If MULTISTAGE is set to "true", then the multistage flag is set.
	// :USERNAME:PASSWORD
	//	This is a normal username and password.
	credsPieces := strings.Split(strings.TrimSpace(s), ":")
	if len(credsPieces) != 3 && len(credsPieces) != 6 {
		return credential{}, fmt.Errorf("Invalid data %q while reading %q", string(s), file)
	}
	if credsPieces[0] == "skip" {
		return credential{skip: true}, nil
	} else if len(credsPieces[0]) == 0 {
		return credential{username: credsPieces[1], password: credsPieces[2]}, nil
	} else if len(credsPieces) == 3 {
		return credential{authtype: credsPieces[0], credential: credsPieces[2]}, nil
	} else {
		return credential{
			authtype:   credsPieces[0],
			credential: credsPieces[2],
			matchState: credsPieces[3],
			state:      credsPieces[4],
			multistage: credsPieces[5] == "true",
		}, nil
	}
}

func credsFromFilename(file string) ([]credential, error) {
	fileContents, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Error opening %q: %s", file, err)
	}
	lines := strings.Split(strings.TrimSpace(string(fileContents)), "\n")
	creds := make([]credential, 0, len(lines))
	for _, line := range lines {
		cred, err := parseOneCredential(line, file)
		if err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, nil
}

func log() {
	fmt.Fprintf(os.Stderr, "CREDS received command: %s (ignored)\n", os.Args[1])
}

func firstEntryForKey(input map[string][]string, key string) string {
	if val, ok := input[key]; ok && len(val) > 0 {
		return val[0]
	}
	return ""
}
