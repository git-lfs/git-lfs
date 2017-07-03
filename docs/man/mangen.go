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

func readManDir() (string, []os.FileInfo) {
	rootDirs := []string{
		"..",
		"/tmp/docker_run/git-lfs",
	}

	var err error
	for _, rootDir := range rootDirs {
		fs, err := ioutil.ReadDir(filepath.Join(rootDir, "docs", "man"))
		if err == nil {
			return rootDir, fs
		}
	}

	fmt.Fprintf(os.Stderr, "Failed to open man dir: %v\n", err)
	os.Exit(2)
	return "", nil
}

// Reads all .ronn files & and converts them to string literals
// triggered by "go generate" comment
// Literals are inserted into a map using an init function, this means
// that there are no compilation errors if 'go generate' hasn't been run, just
// blank man files.
func main() {
	fmt.Fprintf(os.Stderr, "Converting man pages into code...\n")
	rootDir, fs := readManDir()
	manDir := filepath.Join(rootDir, "docs", "man")
	out, err := os.Create(filepath.Join(rootDir, "commands", "mancontent_gen.go"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create go file: %v\n", err)
		os.Exit(2)
	}
	out.WriteString("package commands\n\nfunc init() {\n")
	out.WriteString("// THIS FILE IS GENERATED, DO NOT EDIT\n")
	out.WriteString("// Use 'go generate ./commands' to update\n")
	fileregex := regexp.MustCompile(`git-lfs(?:-([A-Za-z\-]+))?.\d.ronn`)
	headerregex := regexp.MustCompile(`^###?\s+([A-Za-z0-9 ]+)`)
	// only pick up caps in links to avoid matching optional args
	linkregex := regexp.MustCompile(`\[([A-Z\- ]+)\]`)
	// man links
	manlinkregex := regexp.MustCompile(`(git)(?:-(lfs))?-([a-z\-]+)\(\d\)`)
	count := 0
	for _, f := range fs {
		if match := fileregex.FindStringSubmatch(f.Name()); match != nil {
			fmt.Fprintf(os.Stderr, "%v\n", f.Name())
			cmd := match[1]
			if len(cmd) == 0 {
				// This is git-lfs.1.ronn
				cmd = "git-lfs"
			}
			out.WriteString("ManPages[\"" + cmd + "\"] = `")
			contentf, err := os.Open(filepath.Join(manDir, f.Name()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open %v: %v\n", f.Name(), err)
				os.Exit(2)
			}
			// Process the ronn to make it nicer as help text
			scanner := bufio.NewScanner(contentf)
			firstHeaderDone := false
			skipNextLineIfBlank := false
			lastLineWasBullet := false
		scanloop:
			for scanner.Scan() {
				line := scanner.Text()
				trimmedline := strings.TrimSpace(line)
				if skipNextLineIfBlank && len(trimmedline) == 0 {
					skipNextLineIfBlank = false
					lastLineWasBullet = false
					continue
				}

				// Special case headers
				if hmatch := headerregex.FindStringSubmatch(line); hmatch != nil {
					header := strings.ToLower(hmatch[1])
					switch header {
					case "synopsis":
						// Ignore this, just go direct to command

					case "description":
						// Just skip the header & newline
						skipNextLineIfBlank = true
					case "options":
						out.WriteString("Options:" + "\n")
					case "see also":
						// don't include any content after this
						break scanloop
					default:
						out.WriteString(strings.ToUpper(header[:1]) + header[1:] + "\n")
						out.WriteString(strings.Repeat("-", len(header)) + "\n")
					}
					firstHeaderDone = true
					lastLineWasBullet = false
					continue
				}

				if lmatches := linkregex.FindAllStringSubmatch(line, -1); lmatches != nil {
					for _, lmatch := range lmatches {
						linktext := strings.ToLower(lmatch[1])
						line = strings.Replace(line, lmatch[0], `"`+strings.ToUpper(linktext[:1])+linktext[1:]+`"`, 1)
					}
				}
				if manmatches := manlinkregex.FindAllStringSubmatch(line, -1); manmatches != nil {
					for _, manmatch := range manmatches {
						line = strings.Replace(line, manmatch[0], strings.Join(manmatch[1:], " "), 1)
					}
				}

				// Skip content until after first header
				if !firstHeaderDone {
					continue
				}
				// OK, content here

				// remove characters that markdown would render invisible in a text env.
				for _, invis := range []string{"`", "<br>"} {
					line = strings.Replace(line, invis, "", -1)
				}

				// indent bullets
				if strings.HasPrefix(line, "*") {
					lastLineWasBullet = true
				} else if lastLineWasBullet && !strings.HasPrefix(line, " ") {
					// indent paragraphs under bullets if not already done
					line = "  " + line
				}

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
