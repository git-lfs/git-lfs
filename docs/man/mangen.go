package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func infof(w io.Writer, format string, a ...interface{}) {
	if !*verbose {
		return
	}
	fmt.Fprintf(w, format, a...)
}

func warnf(w io.Writer, format string, a ...interface{}) {
	fmt.Fprintf(w, format, a...)
}

func readManDir() (string, []os.DirEntry) {
	rootDirs := []string{
		"..",
		"/tmp/docker_run/git-lfs",
	}

	var err error
	for _, rootDir := range rootDirs {
		fs, err := os.ReadDir(filepath.Join(rootDir, "docs", "man"))
		if err == nil {
			return rootDir, fs
		}
	}

	warnf(os.Stderr, "Failed to open man dir: %v\n", err)
	os.Exit(2)
	return "", nil
}

func titleizeXref(s string) string {
	return strings.Replace(strings.ToTitle(s[1:2])+s[2:], "_", " ", -1)
}

var (
	verbose = flag.Bool("verbose", false, "Show verbose output.")
)

// Reads all .adoc files & and converts them to string literals
// triggered by "go generate" comment
// Literals are inserted into a map using an init function, this means
// that there are no compilation errors if 'go generate' hasn't been run, just
// blank man files.
func main() {
	flag.Parse()

	infof(os.Stderr, "Converting man pages into code...\n")
	rootDir, fs := readManDir()
	manDir := filepath.Join(rootDir, "docs", "man")
	out, err := os.Create(filepath.Join(rootDir, "commands", "mancontent_gen.go"))
	if err != nil {
		warnf(os.Stderr, "Failed to create go file: %v\n", err)
		os.Exit(2)
	}
	out.WriteString("package commands\n\nfunc init() {\n")
	out.WriteString("\t// THIS FILE IS GENERATED, DO NOT EDIT\n")
	out.WriteString("\t// Use 'go generate ./commands' to update\n")
	fileregex := regexp.MustCompile(`git-lfs(?:-([A-Za-z\-]+))?.adoc`)
	headerregex := regexp.MustCompile(`^(===?)\s+([A-Za-z0-9 -]+)`)
	// cross-references
	linkregex := regexp.MustCompile(`<<([^,>]+)(?:,([^>]+))?>>`)
	// man links
	manlinkregex := regexp.MustCompile(`(git)(?:-(lfs))?-([a-z\-]+)\(\d\)`)
	// source blocks
	sourceblockregex := regexp.MustCompile(`\[source(,.*)?\]`)
	// anchors
	anchorregex := regexp.MustCompile(`\[\[(.+)\]\]`)
	count := 0
	for _, f := range fs {
		if match := fileregex.FindStringSubmatch(f.Name()); match != nil {
			infof(os.Stderr, "%v\n", f.Name())
			cmd := match[1]
			if len(cmd) == 0 {
				// This is git-lfs.1.adoc
				cmd = "git-lfs"
			}
			out.WriteString("\tManPages[\"" + cmd + "\"] = `")
			contentf, err := os.Open(filepath.Join(manDir, f.Name()))
			if err != nil {
				warnf(os.Stderr, "Failed to open %v: %v\n", f.Name(), err)
				os.Exit(2)
			}
			// Process the asciidoc to make it nicer as help text
			scanner := bufio.NewScanner(contentf)
			firstHeaderDone := false
			skipNextLineIfBlank := false
			lastLineWasList := false
			isSourceBlock := false
			sourceBlockLine := ""
		scanloop:
			for scanner.Scan() {
				line := scanner.Text()
				trimmedline := strings.TrimSpace(line)
				if skipNextLineIfBlank && len(trimmedline) == 0 {
					skipNextLineIfBlank = false
					lastLineWasList = false
					continue
				}

				// Special case headers
				if hmatch := headerregex.FindStringSubmatch(line); hmatch != nil {
					if len(hmatch[1]) == 2 {
						header := strings.ToLower(hmatch[2])
						switch header {
						case "name":
							continue
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
					} else {
						out.WriteString(hmatch[2] + "\n")
						out.WriteString(strings.Repeat("~", len(hmatch[2])) + "\n")
					}
					lastLineWasList = false
					continue
				}

				if lmatches := linkregex.FindAllStringSubmatch(line, -1); lmatches != nil {
					for _, lmatch := range lmatches {
						if len(lmatch) > 2 && lmatch[2] != "" {
							line = strings.Replace(line, lmatch[0], `"`+lmatch[2]+`"`, 1)
						} else {
							line = strings.Replace(line, lmatch[0], `"`+titleizeXref(lmatch[1])+`"`, 1)
						}
					}
				}
				if manmatches := manlinkregex.FindAllStringSubmatch(line, -1); manmatches != nil {
					for _, manmatch := range manmatches {
						line = strings.Replace(line, manmatch[0], strings.Join(manmatch[1:], " "), 1)
					}
				}

				if sourceblockmatches := sourceblockregex.FindStringIndex(line); sourceblockmatches != nil {
					isSourceBlock = true
					continue
				}

				if anchormatches := anchorregex.FindStringIndex(line); anchormatches != nil {
					// Skip anchors.
					continue
				}

				// Skip content until after first header
				if !firstHeaderDone {
					continue
				}
				// OK, content here

				// handle source block headers
				if isSourceBlock {
					sourceBlockLine = line
					isSourceBlock = false
					line = ""
					continue
				} else if sourceBlockLine != "" && line == sourceBlockLine {
					line = ""
					sourceBlockLine = ""
				}

				// remove characters that asciidoc would render invisible in a text env.
				for _, invis := range []string{"`", "...."} {
					line = strings.Replace(line, invis, "", -1)
				}
				line = strings.TrimSuffix(line, " +")

				// indent bullets and definition lists
				if strings.HasPrefix(line, "*") {
					lastLineWasList = true
				} else if strings.HasSuffix(line, "::") {
					lastLineWasList = true
					line = strings.TrimSuffix(line, ":")
				} else if lastLineWasList && line == "+" {
					line = ""
				} else if lastLineWasList && line == "" {
					lastLineWasList = false
				} else if lastLineWasList && !strings.HasPrefix(line, " ") {
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
	infof(os.Stderr, "Successfully processed %d man pages.\n", count)

}
