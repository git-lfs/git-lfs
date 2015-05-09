package commands

import (
	"bufio"
	"fmt"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	trackCmd = &cobra.Command{
		Use:   "track",
		Short: "Manipulate .gitattributes",
		Run:   trackCommand,
	}
)

// trackCommand takes a list of paths as an argument, and adds each path to the default
// attribtues file (.gitattributes), if it's not already exist.
func trackCommand(cmd *cobra.Command, args []string) {
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

	if addTrailingLinebreak {
		if _, err := attributesFile.WriteString("\n"); err != nil {
			Print("Error writing to .gitattributes")
		}
	}

	for _, t := range args {
		isKnownPath := false
		for _, k := range knownPaths {
			if t == k.Path {
				isKnownPath = true
				break
			}
		}

		if isKnownPath {
			Print("%s already supported", t)
			continue
		}

		_, err := attributesFile.WriteString(fmt.Sprintf("%s filter=lfs -crlf\n", t))
		if err != nil {
			Print("Error adding path %s", t)
			continue
		}
		Print("Tracking %s", t)
	}

	attributesFile.Close()
}

// mediaPath represents a tracked path.
type mediaPath struct {
	Path   string // the tracked path
	Source string // the source attributes file
}

// findAttributeFiles returns the absolute paths to files 
// in the current repository that Git uses to track attributes.
func findAttributeFiles() []string {
	paths := make([]string, 0)

	repoAttributes := filepath.Join(lfs.LocalGitDir, "info", "attributes")
	if _, err := os.Stat(repoAttributes); err == nil {
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

// findPaths extracts the paths tracked by Git LFS from the existing Git attributes files.
func findPaths() []mediaPath {
	paths := make([]mediaPath, 0)
	wd, _ := os.Getwd()

	for _, path := range findAttributeFiles() {
		attributes, err := os.Open(path)
		if err != nil {
			return paths
		}

		scanner := bufio.NewScanner(attributes)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			if strings.Contains(line, "filter=lfs") {
				fields := strings.Fields(line)
				relPath, _ := filepath.Rel(wd, path)
				paths = append(paths, mediaPath{Path: fields[0], Source: relPath})
			}
		}
	}

	return paths
}

// needsTrailingLinebreak returns 'true' if a file doesn't end in a newline,
// 'false' otherwise (even if an error occures).
func needsTrailingLinebreak(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}

	defer file.Close()

	// Reading the file in chuncks of 16384 bytes,
	// to avoid holding the entire file in memory.
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
