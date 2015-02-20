package commands

import (
	"fmt"
	"github.com/hawser/git-hawser/hawser"
	"github.com/hawser/git-hawser/pointer"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	pointersCmd = &cobra.Command{
		Use:   "pointers",
		Short: "Show pointers in the working directory that can be downloaded",
		Run:   pointersCommand,
	}
)

func pointersCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		walkForPointers(hawser.LocalWorkingDir, printPointer)
		return
	}

	rootDir := hawser.LocalWorkingDir + string(filepath.Separator)

	for _, path := range args {
		fullPath := filepath.Join(hawser.LocalWorkingDir, path)
		if !strings.HasPrefix(fullPath, rootDir) {
			Exit("Attempting to scan %s, outside the working directory of %s", fullPath, hawser.LocalWorkingDir)
		}
		walkForPointers(fullPath, printPointer)
	}
}

// This re-implements filepath.Walk() without sorting or reading an entire
// directory into memory at a time.
func walkForPointers(path string, cb func(string, os.FileInfo) error) {
	if cb == nil {
		Panic(fmt.Errorf("No pointer callback passed."), "Error attempting to scan %s", path)
	}

	if strings.HasSuffix(path, "/.git") {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		Panic(err, "Error scanning %s", path)
	}

	if !info.IsDir() {
		if err = cb(path, info); err != nil {
			Panic(err, "Error scanning %s", path)
		}
		return
	}

	if err = readDir(path, cb); err != nil {
		Panic(err, "Error scanning %s", path)
	}
}

func readDir(path string, cb func(string, os.FileInfo) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var dirs []os.FileInfo

	for {
		dirs, err = file.Readdir(50)
		if len(dirs) > 0 {
			for _, dir := range dirs {
				walkForPointers(filepath.Join(path, dir.Name()), cb)
			}
		}

		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func printPointer(path string, info os.FileInfo) error {
	if info.Size() > 200 {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = pointer.Decode(file)
	if err != nil {
		return nil
	}

	fmt.Println(path)
	return nil
}

func init() {
	RootCmd.AddCommand(pointersCmd)
}
