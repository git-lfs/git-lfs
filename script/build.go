package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/github/git-lfs/lfs"
)

var (
	BuildOS    = flag.String("os", "", "OS to target: darwin, freebsd, linux, windows")
	BuildArch  = flag.String("arch", "", "Arch to target: 386, amd64")
	BuildAll   = flag.Bool("all", false, "Builds all architectures")
	ShowHelp   = flag.Bool("help", false, "Shows help")
	matrixKeys = map[string]string{
		"darwin":  "Mac",
		"freebsd": "FreeBSD",
		"linux":   "Linux",
		"windows": "Windows",
		"amd64":   "AMD64",
	}
	LdFlag string
)

func mainBuild() {
	if *ShowHelp {
		fmt.Println("usage: script/bootstrap [-os] [-arch] [-all]")
		flag.PrintDefaults()
		return
	}

	fmt.Printf("Using %s\n", runtime.Version())

	genOut, err := exec.Command("go", "generate", "./commands").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "go generate failed:\n%v", string(genOut))
		os.Exit(1)
	}
	cmd, _ := exec.Command("git", "rev-parse", "--short", "HEAD").Output()

	if len(cmd) > 0 {
		LdFlag = strings.TrimSpace("-X github.com/github/git-lfs/lfs.GitCommit " + string(cmd))
	}

	buildMatrix := make(map[string]Release)
	errored := false

	if *BuildAll {
		for _, buildos := range []string{"linux", "windows", "darwin", "freebsd"} {
			for _, buildarch := range []string{"amd64", "386"} {
				if err := build(buildos, buildarch, buildMatrix); err != nil {
					errored = true
				}
			}
		}
	} else {
		if err := build(*BuildOS, *BuildArch, buildMatrix); err != nil {
			log.Fatalln(err)
		}
		return // skip build matrix stuff
	}

	if errored {
		os.Exit(1)
	}

	by, err := json.Marshal(buildMatrix)
	if err != nil {
		log.Fatalln("Error encoding build matrix to json:", err)
	}

	file, err := os.Create("bin/releases/build_matrix.json")
	if err != nil {
		log.Fatalln("Error creating build_matrix.json:", err)
	}

	written, err := file.Write(by)
	file.Close()

	if err != nil {
		log.Fatalln("Error writing build_matrix.json", err)
	}

	if jsonSize := len(by); written != jsonSize {
		log.Fatalf("Expected to write %d bytes, actually wrote %d.\n", jsonSize, written)
	}
}

func build(buildos, buildarch string, buildMatrix map[string]Release) error {
	addenv := len(buildos) > 0 && len(buildarch) > 0
	name := "git-lfs-" + lfs.Version
	dir := "bin"

	if addenv {
		fmt.Printf("Building for %s/%s\n", buildos, buildarch)
		dir = filepath.Join(dir, "releases", buildos+"-"+buildarch, name)
	}

	if err := buildCommand(dir, buildos, buildarch); err != nil {
		return err
	}

	if addenv {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Println("Error setting up installer:\n", err.Error())
			return err
		}

		err = setupInstaller(buildos, buildarch, dir, buildMatrix)
		if err != nil {
			log.Println("Error setting up installer:\n", err.Error())
			return err
		}
	}

	return nil
}

func buildCommand(dir, buildos, buildarch string) error {
	addenv := len(buildos) > 0 && len(buildarch) > 0

	bin := filepath.Join(dir, "git-lfs")

	if buildos == "windows" {
		bin = bin + ".exe"
	}

	args := make([]string, 1, 6)
	args[0] = "build"
	if len(LdFlag) > 0 {
		args = append(args, "-ldflags", LdFlag)
	}
	args = append(args, "-o", bin, ".")

	cmd := exec.Command("go", args...)
	if addenv {
		cmd.Env = []string{
			"GOOS=" + buildos,
			"GOARCH=" + buildarch,
			"GOPATH=" + os.Getenv("GOPATH"),
			"GOROOT=" + os.Getenv("GOROOT"),
		}
	}

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println(string(output))
	}
	return err
}

func setupInstaller(buildos, buildarch, dir string, buildMatrix map[string]Release) error {
	files := []string{
		"README.md", "CHANGELOG.md",
	}

	for _, filename := range files {
		cmd := exec.Command("cp", filename, filepath.Join(dir, filename))
		if err := logAndRun(cmd); err != nil {
			return err
		}
	}

	// Windows installer is uploaded separately. See script/nsis.
	if buildos == "windows" {
		return nil
	}

	fullInstallPath := filepath.Join(dir, "install.sh")
	cmd := exec.Command("cp", "script/install.sh.example", fullInstallPath)
	if err := logAndRun(cmd); err != nil {
		return err
	}

	if err := os.Chmod(fullInstallPath, 0755); err != nil {
		return err
	}

	name := zipName(buildos, buildarch) + ".tar.gz"

	addToMatrix(buildMatrix, buildos, buildarch, name)

	cmd = exec.Command("tar", "czf", "../"+name, filepath.Base(dir))
	cmd.Dir = filepath.Dir(dir)
	return logAndRun(cmd)
}

func addToMatrix(buildMatrix map[string]Release, buildos, buildarch, name string) {
	buildMatrix[fmt.Sprintf("%s-%s", buildos, buildarch)] = Release{
		Label:    releaseLabel(buildos, buildarch),
		Filename: name,
	}
}

func logAndRun(cmd *exec.Cmd) error {
	fmt.Printf(" - %s\n", strings.Join(cmd.Args, " "))
	if len(cmd.Dir) > 0 {
		fmt.Printf("   - in %s\n", cmd.Dir)
	}

	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	return err
}

func zipName(os, arch string) string {
	return fmt.Sprintf("git-lfs-%s-%s-%s", os, arch, lfs.Version)
}

func releaseLabel(buildos, buildarch string) string {
	return fmt.Sprintf("%s %s", key(buildos), key(buildarch))
}

func key(k string) string {
	if s, ok := matrixKeys[k]; ok {
		return s
	}
	return k
}
