package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	trackCmd = &cobra.Command{
		Use:   "track",
		Short: "Manipulate .gitattributes",
		Run:   trackCommand,
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

ArgsLoop:
	for _, t := range args {
		absT, relT := absRelPath(t, wd)

		if !filepath.HasPrefix(absT, lfs.LocalWorkingDir) {
			Print("%s is outside repository", t)
			os.Exit(128)
		}

		for _, k := range knownPaths {
			absK, _ := absRelPath(k.Path, filepath.Join(wd, filepath.Dir(k.Source)))
			if absT == absK {
				Print("%s already supported", t)
				continue ArgsLoop
			}
		}

		encodedArg := strings.Replace(relT, " ", "[[:space:]]", -1)
		_, err := attributesFile.WriteString(fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text\n", encodedArg))
		if err != nil {
			Print("Error adding path %s", t)
			continue
		}
		Print("Tracking %s", t)
	}
}

type mediaPath struct {
	Path   string
	Source string
}

func findPaths() []mediaPath {
	paths := make([]mediaPath, 0)
	wd, _ := os.Getwd()

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
				relPath, _ := filepath.Rel(wd, path)
				paths = append(paths, mediaPath{Path: fields[0], Source: relPath})
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

// absRelPath takes a path and a working directory and
// returns an absolute and a relative representation of path based on the working directory
func absRelPath(path, wd string) (string, string) {
	if filepath.IsAbs(path) {
		relPath, _ := filepath.Rel(wd, path)
		return path, relPath
	}

	absPath := filepath.Join(wd, path)
	return absPath, path
}

func init() {
	RootCmd.AddCommand(trackCmd)
}
