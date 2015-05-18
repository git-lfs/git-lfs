package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	commands = map[string]func(){
		"get":   fill,
		"store": noop,
		"erase": noop,
	}

	delim = '\n'

	hostRE = regexp.MustCompile(`\A127.0.0.1:\d+\z`)
)

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

		creds[parts[0]] = strings.TrimSpace(parts[1])
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading standard input: %v", err)
		os.Exit(1)
	}

	if _, ok := creds["username"]; !ok {
		creds["username"] = "user"
	}

	if _, ok := creds["password"]; !ok {
		creds["password"] = "pass"
	}

	if host := creds["host"]; !hostRE.MatchString(host) {
		fmt.Fprintf(os.Stderr, "invalid host: %s, should be '127.0.0.1:\\d+'\n", host)
		os.Exit(1)
	}

	for key, value := range creds {
		fmt.Fprintf(os.Stdout, "%s=%s\n", key, value)
	}
}

func noop() {}
