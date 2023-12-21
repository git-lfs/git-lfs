package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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

func readPoDir() (string, []os.DirEntry) {
	rootDirs := []string{
		".",
		"..",
		"../..",
	}

	var err error
	for _, rootDir := range rootDirs {
		fs, err := os.ReadDir(filepath.Join(rootDir, "po", "build"))
		if err == nil {
			return rootDir, fs
		}
	}

	// In this case, we don't care about the fact that the build dir doesn't
	// exist since that just means there are no translations built.  That's
	// fine for us, so just exit successfully.
	infof(os.Stderr, "Failed to open po dir: %v\n", err)
	os.Exit(0)
	return "", nil
}

var (
	verbose = flag.Bool("verbose", false, "Show verbose output.")
)

func main() {
	flag.Parse()

	infof(os.Stderr, "Converting po files into translations...\n")
	rootDir, fs := readPoDir()
	poDir := filepath.Join(rootDir, "po", "build")
	out, err := os.Create(filepath.Join(rootDir, "tr", "tr_gen.go"))
	if err != nil {
		warnf(os.Stderr, "Failed to create go file: %v\n", err)
		os.Exit(2)
	}
	out.WriteString("package tr\n\nfunc init() {\n")
	out.WriteString("\t// THIS FILE IS GENERATED, DO NOT EDIT\n")
	out.WriteString("\t// Use 'go generate ./tr/trgen' to update\n")
	fileregex := regexp.MustCompile(`([A-Za-z\-_]+).mo`)
	count := 0
	for _, f := range fs {
		if match := fileregex.FindStringSubmatch(f.Name()); match != nil {
			infof(os.Stderr, "%v\n", f.Name())
			cmd := match[1]
			content, err := os.ReadFile(filepath.Join(poDir, f.Name()))
			if err != nil {
				warnf(os.Stderr, "Failed to open %v: %v\n", f.Name(), err)
				os.Exit(2)
			}
			fmt.Fprintf(out, "\tlocales[\"%s\"] = \"%s\"\n", cmd, base64.StdEncoding.EncodeToString(content))
			count++
		}
	}
	out.WriteString("}\n")
	infof(os.Stderr, "Successfully processed %d translations.\n", count)

}
