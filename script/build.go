package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/github/git-lfs/lfs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
)

func mainBuild() {
	cmd, err := exec.Command("script/fmt").Output()
	if err != nil {
		panic(err)
	}

	if len(cmd) > 0 {
		fmt.Println(string(cmd))
	}

	if *ShowHelp {
		fmt.Println("usage: script/bootstrap [-os] [-arch] [-all]")
		flag.PrintDefaults()
		return
	}

	buildMatrix := make(map[string]Release)

	errored := false

	if *BuildAll {
		for _, buildos := range []string{"darwin", "freebsd", "linux", "windows"} {
			for _, buildarch := range []string{"386", "amd64"} {
				if err := build(buildos, buildarch, buildMatrix); err != nil {
					errored = true
				}
			}
		}
	} else {
		build(*BuildOS, *BuildArch, buildMatrix)
		errored = true // skip build matrix stuff
	}

	if !errored {
		by, err := json.Marshal(buildMatrix)
		if err != nil {
			fmt.Println("Error encoding build matrix to json:", err)
			os.Exit(1)
		}

		file, err := os.Create("bin/releases/build_matrix.json")
		if err != nil {
			fmt.Println("Error creating build_matrix.json:", err)
			os.Exit(1)
		}

		written, err := file.Write(by)
		file.Close()

		if err != nil {
			fmt.Println("Error writing build_matrix.json", err)
			os.Exit(1)
		}

		if jsonSize := len(by); written != jsonSize {
			fmt.Printf("Expected to write %d bytes, actually wrote %d.\n", jsonSize, written)
			os.Exit(1)
		}
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

	filepath.Walk("cmd", func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		cmd := filepath.Base(path)
		cmd = cmd[0 : len(cmd)-3]
		return buildCommand(path, dir, buildos, buildarch)
	})

	if addenv {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error setting up installer:\n", err.Error())
			return err
		}

		err = setupInstaller(buildos, buildarch, dir, buildMatrix)
		if err != nil {
			fmt.Println("Error setting up installer:\n", err.Error())
			return err
		}
	}

	return nil
}

func buildCommand(path, dir, buildos, buildarch string) error {
	addenv := len(buildos) > 0 && len(buildarch) > 0
	name := filepath.Base(path)
	name = name[0 : len(name)-3]

	bin := filepath.Join(dir, name)

	if buildos == "windows" {
		bin = bin + ".exe"
	}

	cmd := exec.Command("go", "build", "-o", bin, path)
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
	if buildos == "windows" {
		return winInstaller(buildos, buildarch, dir, buildMatrix)
	} else {
		return unixInstaller(buildos, buildarch, dir, buildMatrix)
	}
}

func unixInstaller(buildos, buildarch, dir string, buildMatrix map[string]Release) error {
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

func winInstaller(buildos, buildarch, dir string, buildMatrix map[string]Release) error {
	cmd := exec.Command("cp", "script/install.bat.example", filepath.Join(dir, "install.bat"))
	if err := logAndRun(cmd); err != nil {
		return err
	}

	installerPath := filepath.Dir(filepath.Dir(dir))

	name := zipName(buildos, buildarch) + ".zip"
	full := filepath.Join(installerPath, name)
	matches, err := filepath.Glob(dir + "/*")
	if err != nil {
		return err
	}

	addToMatrix(buildMatrix, buildos, buildarch, name)

	args := make([]string, len(matches)+2)
	args[0] = "-j" // junk the zip paths
	args[1] = full
	copy(args[2:], matches)

	cmd = exec.Command("zip", args...)
	return logAndRun(cmd)
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
