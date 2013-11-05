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
		err := setupInstaller(buildos, buildarch, dir)
		if err != nil {
			fmt.Println("Error setting up installer:\n", err.Error())
		}
	}
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
		return winInstaller(buildos, buildarch, dir)
	} else {
		return unixInstaller(buildos, buildarch, dir)
	}
}

func unixInstaller(buildos, buildarch, dir string) error {
	cmd := exec.Command("cp", "script/install.sh.example", filepath.Join(dir, "install.sh"))
	if err := logAndRun(cmd); err != nil {
		return err
	}

	name := zipName(buildos, buildarch) + ".tar.gz"
	cmd = exec.Command("tar", "czf", name, filepath.Base(dir))
	cmd.Dir = filepath.Dir(dir)
	return logAndRun(cmd)
}

func winInstaller(buildos, buildarch, dir string) error {
	cmd := exec.Command("cp", "script/install.bat.example", filepath.Join(dir, "install.bat"))
	if err := logAndRun(cmd); err != nil {
		return err
	}

	name := zipName(buildos, buildarch) + ".zip"
	cmd = exec.Command("zip", name, filepath.Base(dir)+"/*")
	cmd.Dir = filepath.Dir(dir)
	return logAndRun(cmd)
}

func logAndRun(cmd *exec.Cmd) error {
	fmt.Printf(" - %s\n", strings.Join(cmd.Args, " "))
	if len(cmd.Dir) > 0 {
		fmt.Printf("   - in %s\n", cmd.Dir)
	}

	return cmd.Run()
}

func zipName(os, arch string) string {
	return fmt.Sprintf("git-media-%s-%s-v%s", os, arch, gitmedia.Version)
}
