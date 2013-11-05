package main

import (
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
	if len(buildos) > 0 && len(buildarch) > 0 {
		fmt.Printf("Building for %s/%s\n", buildos, buildarch)
	}

	filepath.Walk("cmd", func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		return buildCommand(path, buildos, buildarch)
	})
}

func buildCommand(path, buildos, buildarch string) error {
	base := filepath.Base(path)
	base = base[0 : len(base)-3]
	bin := "bin"
	addenv := len(buildos) > 0 && len(buildarch) > 0

	if addenv {
		bin = filepath.Join(bin, buildos+"-"+buildarch, base)
	} else {
		bin = filepath.Join(bin, base)
	}

	cmd := exec.Command("go", "build", "-o", bin, path)
	var out bytes.Buffer
	cmd.Stderr = &out
	if addenv {
		cmd.Env = []string{"GOOS=" + buildos, "GOARCH=" + buildarch, "GOPATH=" + os.Getenv("GOPATH")}
	}

	if err := cmd.Run(); err != nil {
		fmt.Println(out.String())
		return err
	}

	return nil
}
