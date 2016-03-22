package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/github/git-lfs/git"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	trackCmd = &cobra.Command{
		Use: "track",
		Run: trackCommand,
	}
)

func trackCommand(cmd *cobra.Command, args []string) {
	if lfs.LocalGitDir == "" {
		Print("Not a git repository.")
		os.Exit(128)
	}

	if lfs.LocalWorkingDir == "" {
		Print("This operation must be run in a work tree.")
		os.Exit(128)
	}

	lfs.InstallHooks(false)
	knownPaths := findPaths()

	if len(args) == 0 {
		Print("Listing tracked paths")
		for _, t := range knownPaths {
			Print("    %s (%s)", t.Path, t.Source)
		}
		return
	}

	addTrailingLinebreak := needsTrailingLinebreak(".gitattributes")
	attributesFile, err := os.OpenFile(".gitattributes", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		Print("Error opening .gitattributes file")
		return
	}
	defer attributesFile.Close()

	if addTrailingLinebreak {
		if _, err := attributesFile.WriteString("\n"); err != nil {
			Print("Error writing to .gitattributes")
		}
	}

	wd, _ := os.Getwd()
	relpath, err := filepath.Rel(lfs.LocalWorkingDir, wd)
	if err != nil {
		Exit("Current directory %q outside of git working directory %q.", wd, lfs.LocalWorkingDir)
	}

ArgsLoop:
	for _, pattern := range args {
		for _, known := range knownPaths {
			if known.Path == filepath.Join(relpath, pattern) {
				Print("%s already supported", pattern)
				continue ArgsLoop
			}
		}

		encodedArg := strings.Replace(pattern, " ", "[[:space:]]", -1)
		_, err := attributesFile.WriteString(fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text\n", encodedArg))
		if err != nil {
			Print("Error adding path %s", pattern)
			continue
		}
		Print("Tracking %s", pattern)

		// Make sure any existing git tracked files have their timestamp updated
		// so they will now show as modifed
		// note this is relative to current dir which is how we write .gitattributes
		// deliberately not done in parallel as a chan because we'll be marking modified
		gittracked, err := git.GetTrackedFiles(pattern)
		if err != nil {
			LoggedError(err, "Error getting git tracked files")
			continue
		}
		now := time.Now()
		for _, f := range gittracked {
			err := os.Chtimes(f, now, now)
			if err != nil {
				LoggedError(err, "Error marking %q modified", f)
				continue
			}
		}

	}
}

type mediaPath struct {
	Path   string
	Source string
}

func findPaths() []mediaPath {
	paths := make([]mediaPath, 0)

	for _, path := range findAttributeFiles() {
		attributes, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(attributes)

		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "filter=lfs") {
				fields := strings.Fields(line)
				relfile, _ := filepath.Rel(lfs.LocalWorkingDir, path)
				pattern := fields[0]
				if reldir := filepath.Dir(relfile); len(reldir) > 0 {
					pattern = filepath.Join(reldir, pattern)
				}

				paths = append(paths, mediaPath{Path: pattern, Source: relfile})
			}
		}
	}

	return paths
}

func findAttributeFiles() []string {
	paths := make([]string, 0)

	repoAttributes := filepath.Join(lfs.LocalGitDir, "info", "attributes")
	if info, err := os.Stat(repoAttributes); err == nil && !info.IsDir() {
		paths = append(paths, repoAttributes)
	}

	filepath.Walk(lfs.LocalWorkingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (filepath.Base(path) == ".gitattributes") {
			paths = append(paths, path)
		}
		return nil
	})

	return paths
}

func needsTrailingLinebreak(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 16384)
	bytesRead := 0
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			return false
		}
		bytesRead = n
	}

	return !strings.HasSuffix(string(buf[0:bytesRead]), "\n")
}

func init() {
	RootCmd.AddCommand(trackCmd)
}
