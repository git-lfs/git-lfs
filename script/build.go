package main

import (
	".."
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	BuildOS   = flag.String("os", "", "OS to target: darwin, freebsd, linux, windows")
	BuildArch = flag.String("arch", "", "Arch to target: 386, amd64")
	BuildAll  = flag.Bool("all", false, "Builds all architectures")
	ShowHelp  = flag.Bool("help", false, "Shows help")
)

func main() {
	cmd, err := exec.Command("script/fmt").Output()
	if err != nil {
		panic(err)
	}

	if len(cmd) > 0 {
		fmt.Println(string(cmd))
	}

	flag.Parse()
	if *ShowHelp {
		fmt.Println("usage: script/build [-os] [-arch] [-all]")
		flag.PrintDefaults()
		return
	}

	if *BuildAll {
		for _, buildos := range []string{"darwin", "freebsd", "linux", "windows"} {
			for _, buildarch := range []string{"386", "amd64"} {
				build(buildos, buildarch)
			}
		}
	} else {
		build(*BuildOS, *BuildArch)
	}
}

func build(buildos, buildarch string) {
	addenv := len(buildos) > 0 && len(buildarch) > 0
	name := "git-media-v" + gitmedia.Version
	dir := "bin"

	if addenv {
		fmt.Printf("Building for %s/%s\n", buildos, buildarch)
		dir = filepath.Join(dir, "installers", buildos+"-"+buildarch, name)
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
		setupInstaller(buildos, buildarch, dir)
	}
}

func buildCommand(path, dir, buildos, buildarch string) error {
	addenv := len(buildos) > 0 && len(buildarch) > 0
	name := filepath.Base(path)
	name = name[0 : len(name)-3]

	bin := filepath.Join(dir, name)

	cmd := exec.Command("go", "build", "-o", bin, path)
	var out bytes.Buffer
	cmd.Stderr = &out
	if addenv {
		cmd.Env = []string{
			"GOOS=" + buildos,
			"GOARCH=" + buildarch,
			"GOPATH=" + os.Getenv("GOPATH"),
		}
	}

	if err := cmd.Run(); err != nil {
		fmt.Println(out.String())
		return err
	}

	return nil
}

func setupInstaller(buildos, buildarch, dir string) error {
	if buildos == "windows" {
		return nil // Click here to install.bat
	}

	cmd := exec.Command("cp", "script/install.sh.example", filepath.Join(dir, "install.sh"))
	fmt.Printf(" - %s\n", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		return err
	}

	zipName := fmt.Sprintf("git-media-%s-%s-v%s.tar.gz", buildos, buildarch, gitmedia.Version)
	cmd = exec.Command("tar", "czf", zipName, filepath.Base(dir))
	fmt.Printf(" - %s\n", strings.Join(cmd.Args, " "))
	cmd.Dir = filepath.Dir(dir)
	return cmd.Run()
}
