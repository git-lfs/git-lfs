package commands

import (
	"bufio"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	pathCmd = &cobra.Command{
		Use:   "path",
		Short: "Manipulate .gitattributes",
		Run:   pathCommand,
	}
)

func pathCommand(cmd *cobra.Command, args []string) {
	lfs.InstallHooks(false)

	Print("Listing paths")
	knownPaths := findPaths()
	for _, t := range knownPaths {
		Print("    %s (%s)", t.Path, t.Source)
	}
}

type mediaPath struct {
	Path   string
	Source string
}

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

func init() {
	RootCmd.AddCommand(pathCmd)
}
