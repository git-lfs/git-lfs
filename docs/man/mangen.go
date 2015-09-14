package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Reads all .ronn files & and converts them to string literals
// triggered by "go generate" comment
// Literals are inserted into a map using an init function, this means
// that there are no compilation errors if 'go generate' hasn't been run, just
// blank man files.
func main() {
	fmt.Fprintf(os.Stderr, "Converting man pages into code...\n")
	mandir := "../docs/man"
	fs, err := ioutil.ReadDir(mandir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open man dir: %v\n", err)
		os.Exit(2)
	}
	out, err := os.Create("../commands/mancontent_gen.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create go file: %v\n", err)
		os.Exit(2)
	}
	out.WriteString("package commands\n\nfunc init() {\n")
	out.WriteString("// THIS FILE IS GENERATED, DO NOT EDIT\n")
	out.WriteString("// Use 'go generate ./commands' to update\n")
	r := regexp.MustCompile(`git-lfs(?:-([A-Za-z\-]+))?.\d.ronn`)
	count := 0
	for _, f := range fs {
		if match := r.FindStringSubmatch(f.Name()); match != nil {
			fmt.Fprintf(os.Stderr, "%v\n", f.Name())
			cmd := match[1]
			if len(cmd) == 0 {
				// This is git-lfs.1.ronn
				cmd = "git-lfs"
			}
			out.WriteString("ManPages[\"" + cmd + "\"] = `")
			contentf, err := os.Open(filepath.Join(mandir, f.Name()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open %v: %v\n", f.Name(), err)
				os.Exit(2)
			}
			// Process the ronn to make it nicer as help text
			scanner := bufio.NewScanner(contentf)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				// Remove backticks since it won't format & that's delimiting the string
				line = strings.Replace(line, "`", "", -1)
				// Maybe more e.g. ## ?
				out.WriteString(line + "\n")
			}
			out.WriteString("`\n")
			contentf.Close()
			count++
		}
	}
	out.WriteString("}\n")
	fmt.Fprintf(os.Stderr, "Successfully processed %d man pages.\n", count)

}
