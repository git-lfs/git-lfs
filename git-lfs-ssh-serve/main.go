package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
)

func main() {
	// Need to send the result code to the OS but also need to support 'defer'
	// os.Exit would finish before any defers, so wrap everything in mainImpl()
	os.Exit(MainImpl())

}

func MainImpl() int {

	// Generic panic handler so we get stack trace
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintf(os.Stderr, "git-lfs-ssh-serve panic: %v\n", e)
			fmt.Fprint(os.Stderr, string(debug.Stack()))
			os.Exit(99)
		}

	}()

	// Get set up
	cfg := LoadConfig()

	if cfg.BasePath == "" {
		fmt.Fprintf(os.Stderr, "Missing required configuration setting: base-path\n")
		return 12
	}
	if !dirExists(cfg.BasePath) {
		fmt.Fprintf(os.Stderr, "Invalid value for base-path: %v\nDirectory must exist.\n", cfg.BasePath)
		return 14
	}
	// Change to the base path directory so filepath.Clean() can work with relative dirs
	os.Chdir(cfg.BasePath)

	if cfg.DeltaCachePath != "" && !dirExists(cfg.DeltaCachePath) {
		// Create delta cache if doesn't exist, use same permissions as base path
		s, err := os.Stat(cfg.BasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid value for base-path: %v\nCannot stat: %v\n", cfg.BasePath, err.Error())
			return 16
		}
		err = os.MkdirAll(cfg.DeltaCachePath, s.Mode())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating delta cache path %v: %v\n", cfg.DeltaCachePath, err.Error())
			return 16
		}
	}

	// Get path argument
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Path argument missing, cannot continue\n")
		return 18
	}
	path := filepath.Clean(os.Args[1])
	if filepath.IsAbs(path) && !cfg.AllowAbsolutePaths {
		fmt.Fprintf(os.Stderr, "Path argument %v invalid, absolute paths are not allowed by this server\n", path)
		return 18
	}

	return Serve(os.Stdin, os.Stdout, os.Stderr, cfg, path)
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}
