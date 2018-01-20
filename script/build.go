package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/git-lfs/git-lfs/config"
)

var (
	BuildOS      = flag.String("os", "", "OS to target: darwin,freebsd,linux,windows")
	BuildArch    = flag.String("arch", "", "Arch to target: 386,amd64")
	BuildAll     = flag.Bool("all", false, "Builds all architectures")
	BuildDwarf   = flag.Bool("dwarf", false, "Includes DWARF tables in build artifacts")
	BuildLdFlags = flag.String("ldflags", "", "-ldflags to pass to the compiler")
	BuildGcFlags = flag.String("gcflags", "", "-gcflags to pass to the compiler")
	ShowHelp     = flag.Bool("help", false, "Shows help")
	matrixKeys   = map[string]string{
		"darwin":  "Mac",
		"freebsd": "FreeBSD",
		"linux":   "Linux",
		"windows": "Windows",
		"amd64":   "AMD64",
	}
	LdFlags []string
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
		LdFlags = append(LdFlags, "-X", strings.TrimSpace(
			"github.com/git-lfs/git-lfs/config.GitCommit="+string(cmd),
		))
	}
	if !*BuildDwarf {
		LdFlags = append(LdFlags, "-s", "-w")
	}

	buildMatrix := make(map[string]Release)
	errored := false

	var platforms, arches []string
	if len(*BuildOS) > 0 {
		platforms = strings.Split(*BuildOS, ",")
	}
	if len(*BuildArch) > 0 {
		arches = strings.Split(*BuildArch, ",")
	}
	if *BuildAll {
		platforms = []string{"linux", "darwin", "freebsd", "windows"}
		arches = []string{"amd64", "386"}
	}

	if len(platforms) < 1 || len(arches) < 1 {
		if err := build("", "", buildMatrix); err != nil {
			log.Fatalln(err)
		}
		return // skip build matrix stuff
	}

	for _, buildos := range platforms {
		for _, buildarch := range arches {
			err := build(strings.TrimSpace(buildos), strings.TrimSpace(buildarch), buildMatrix)
			if err != nil {
				errored = true
			}
		}
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
	name := "git-lfs-" + config.Version
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

	cmdOS := runtime.GOOS
	if len(buildos) > 0 {
		cmdOS = buildos
	}
	if cmdOS == "windows" {
		bin = bin + ".exe"
	}

	args := make([]string, 1, 6)
	args[0] = "build"
	if len(*BuildLdFlags) > 0 {
		args = append(args, "-ldflags", *BuildLdFlags)
	} else if len(LdFlags) > 0 {
		args = append(args, "-ldflags", strings.Join(LdFlags, " "))
	}

	if len(*BuildGcFlags) > 0 {
		args = append(args, "-gcflags", *BuildGcFlags)
	}

	args = append(args, "-o", bin, ".")

	cmd := exec.Command("go", args...)
	if addenv {
		cmd.Env = buildGoEnv(buildos, buildarch)
	}

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println(string(output))
	}
	return err
}

func buildGoEnv(buildos, buildarch string) []string {
	env := make([]string, 6, 9)
	env[0] = "GOOS=" + buildos
	env[1] = "GOARCH=" + buildarch
	env[2] = "GOPATH=" + os.Getenv("GOPATH")
	env[3] = "GOROOT=" + os.Getenv("GOROOT")
	env[4] = "PATH=" + os.Getenv("PATH")
	env[5] = "GO15VENDOREXPERIMENT=" + os.Getenv("GO15VENDOREXPERIMENT")
	for _, key := range []string{"TMP", "TEMP", "TEMPDIR"} {
		v := os.Getenv(key)
		if len(v) == 0 {
			continue
		}
		env = append(env, key+"="+v)
	}
	return env
}

func setupInstaller(buildos, buildarch, dir string, buildMatrix map[string]Release) error {
	textfiles := []string{
		"README.md", "CHANGELOG.md",
	}

	if buildos == "windows" {
		return winInstaller(textfiles, buildos, buildarch, dir, buildMatrix)
	} else {
		return unixInstaller(textfiles, buildos, buildarch, dir, buildMatrix)
	}
}

func unixInstaller(textfiles []string, buildos, buildarch, dir string, buildMatrix map[string]Release) error {
	for _, filename := range textfiles {
		cmd := exec.Command("cp", filename, filepath.Join(dir, filename))
		if err := logAndRun(cmd); err != nil {
			return err
		}
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
	cmd = exec.Command("tar", "czf", "../"+name, filepath.Base(dir))
	cmd.Dir = filepath.Dir(dir)
	if err := logAndRun(cmd); err != nil {
		return nil
	}

	addToMatrix(buildMatrix, buildos, buildarch, name)
	return nil
}

func winInstaller(textfiles []string, buildos, buildarch, dir string, buildMatrix map[string]Release) error {
	for _, filename := range textfiles {
		by, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		winEndings := strings.Replace(string(by), "\n", "\r\n", -1)
		err = ioutil.WriteFile(filepath.Join(dir, filename), []byte(winEndings), 0644)
		if err != nil {
			return err
		}
	}

	installerPath := filepath.Dir(filepath.Dir(dir))
	name := zipName(buildos, buildarch) + ".zip"
	full := filepath.Join(installerPath, name)
	matches, err := filepath.Glob(dir + "/*")
	if err != nil {
		return err
	}

	args := make([]string, len(matches)+2)
	args[0] = "-j" // junk the zip paths
	args[1] = full
	copy(args[2:], matches)

	cmd := exec.Command("zip", args...)
	if err := logAndRun(cmd); err != nil {
		return err
	}

	addToMatrix(buildMatrix, buildos, buildarch, name)
	return nil
}

func addToMatrix(buildMatrix map[string]Release, buildos, buildarch, name string) {
	buildMatrix[fmt.Sprintf("%s-%s", buildos, buildarch)] = Release{
		Label:    releaseLabel(buildos, buildarch),
		Filename: name,
		SHA256:   hashRelease(name),
	}
}

func hashRelease(name string) string {
	full := filepath.Join("bin/releases", name)
	file, err := os.Open(full)
	if err != nil {
		fmt.Printf("unable to open release %q: %+v\n", full, err)
		os.Exit(1)
	}

	defer file.Close()

	h := sha256.New()
	if _, err = io.Copy(h, file); err != nil {
		fmt.Printf("error reading release %q: %+v\n", full, err)
		os.Exit(1)
	}

	return hex.EncodeToString(h.Sum(nil))
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
	return fmt.Sprintf("git-lfs-%s-%s-%s", os, arch, config.Version)
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
