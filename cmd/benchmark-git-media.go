package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	ListFiles  = flag.Bool("list", false, "List files only")
	SourcePath = flag.String("s", "/Users/rick/github/github", "Location of existing Git repository")
	DestPath   = flag.String("d", "", "Location to setup benchmark repository")
	Extensions = flag.String("ext", "", "Extensions to filter through git-media, separated by commas")
	WorkingDir string
)

func main() {
	flag.Parse()
	wd, _ := os.Getwd()
	WorkingDir = wd
	if *DestPath == "" {
		*DestPath = filepath.Join(WorkingDir, "benchmark")
	}

	os.Chdir(*SourcePath)

	if *ListFiles {
		list(findfiles())
	} else {
		bench(findfiles())
	}
}

func list(files []string) {
	for _, path := range files {
		fmt.Println(path)
	}
}

func bench(files []string) {
	os.MkdirAll(*DestPath, 0777)
	os.Chdir(*DestPath)

	fmt.Println(safeExec("git", "init"))

	if *Extensions != "" {
		fmt.Printf("Filtering extensions: %s\n", *Extensions)
		gitattr, err := os.Create(filepath.Join(*DestPath, ".gitattributes"))
		if err != nil {
			fmt.Println("Error opening .gitattributes")
			panic(err)
		}

		for _, ext := range strings.Split(*Extensions, ",") {
			gitattr.Write([]byte(fmt.Sprintf("*.%s\tfilter=media\n", ext)))
		}
		gitattr.Close()

		debugsuffix := " -debug 2> " + *DestPath + "/.git/media-debug.log"
		fmt.Println(WorkingDir, filepath.Join(WorkingDir, "bin", "git-media-clean"))
		safeExec("git", "config", "filter.media.clean", filepath.Join(WorkingDir, "bin", "git-media-clean"+debugsuffix))
		safeExec("git", "config", "filter.media.smudge", filepath.Join(WorkingDir, "bin", "git-media-smudge"+debugsuffix))
	}

	fmt.Printf("Copying %d files...\n", len(files))

	for _, path := range files {
		copyFile(path)
	}

	fmt.Println("Adding to git...")

	started := time.Now()
	if _, err := simpleExec("git", "add", "."); err != nil {
		fmt.Printf("Error with git add: %s\n", err)
	}
	cmd, err := simpleExec("git", "commit", "-a", "-m", "benchmark")
	dur := time.Since(started)
	if err != nil {
		fmt.Printf("Error with git add: %s\n", err)
	} else {
		fmt.Println(cmd)
	}
	fmt.Printf("Added in %s\n", dur.String())
}

func copyFile(path string) {
	sourcepath := filepath.Join(*SourcePath, path)
	destpath := filepath.Join(*DestPath, path)

	source, err := os.Open(sourcepath)
	if err != nil {
		return
	}

	os.MkdirAll(filepath.Dir(destpath), 0777)
	dest, err := os.Create(destpath)
	if err != nil {
		fmt.Printf("Error opening destination: %s\n", destpath)
		panic(err)
	}

	io.Copy(dest, source)
	source.Close()
	dest.Close()
}

func findfiles() []string {
	cmd := safeExec("git", "ls-tree", "HEAD", "-r")
	lines := strings.Split(cmd, "\n")
	files := make([]string, len(lines)-1)

	for i := range files {
		pieces := strings.Split(lines[i], "\t")
		files[i] = pieces[len(pieces)-1]
	}
	return files
}

func simpleExec(name string, args ...string) (string, error) {
	cmd, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return string(cmd), nil
}

func safeExec(name string, args ...string) string {
	str, err := simpleExec(name, args...)
	if err != nil {
		panic(err)
	}
	return str
}
