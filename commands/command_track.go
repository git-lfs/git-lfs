package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rubyist/tracerx"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tools/longpathos"
	"github.com/spf13/cobra"
)

var (
	prefixBlocklist = []string{
		".git", ".lfs",
	}

	trackVerboseLoggingFlag bool
	trackDryRunFlag         bool
)

func trackCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()

	if config.LocalGitDir == "" {
		Print("Not a git repository.")
		os.Exit(128)
	}

	if config.LocalWorkingDir == "" {
		Print("This operation must be run in a work tree.")
		os.Exit(128)
	}

	lfs.InstallHooks(false)
	knownPatterns := findPatterns()

	if len(args) == 0 {
		Print("Listing tracked patterns")
		for _, t := range knownPatterns {
			Print("    %s (%s)", t.Pattern, t.Source)
		}
		return
	}

	addTrailingLinebreak := needsTrailingLinebreak(".gitattributes")
	attributesFile, err := longpathos.OpenFile(".gitattributes", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		Print("Error opening .gitattributes file")
		return
	}
	defer attributesFile.Close()

	if addTrailingLinebreak {
		if _, werr := attributesFile.WriteString("\n"); werr != nil {
			Print("Error writing to .gitattributes")
		}
	}

	wd, _ := os.Getwd()
	relpath, err := filepath.Rel(config.LocalWorkingDir, wd)
	if err != nil {
		Exit("Current directory %q outside of git working directory %q.", wd, config.LocalWorkingDir)
	}

ArgsLoop:
	for _, unsanitizedPattern := range args {
		pattern := cleanRootPath(unsanitizedPattern)
		for _, known := range knownPatterns {
			if known.Pattern == filepath.Join(relpath, pattern) {
				Print("%s already supported", pattern)
				continue ArgsLoop
			}
		}

		// Make sure any existing git tracked files have their timestamp updated
		// so they will now show as modifed
		// note this is relative to current dir which is how we write .gitattributes
		// deliberately not done in parallel as a chan because we'll be marking modified
		//
		// NOTE: `git ls-files` does not do well with leading slashes.
		// Since all `git-lfs track` calls are relative to the root of
		// the repository, the leading slash is simply removed for its
		// implicit counterpart.
		if trackVerboseLoggingFlag {
			Print("Searching for files matching pattern: %s", pattern)
		}
		gittracked, err := git.GetTrackedFiles(pattern)
		if err != nil {
			Exit("Error getting tracked files for %q: %s", pattern, err)
		}

		if trackVerboseLoggingFlag {
			Print("Found %d files previously added to Git matching pattern: %s", len(gittracked), pattern)
		}

		var matchedBlocklist bool
		for _, f := range gittracked {
			if forbidden := blocklistItem(f); forbidden != "" {
				Print("Pattern %s matches forbidden file %s. If you would like to track %s, modify .gitattributes manually.", pattern, f, f)
				matchedBlocklist = true
			}

		}
		if matchedBlocklist {
			continue
		}

		if !trackDryRunFlag {
			encodedArg := strings.Replace(pattern, " ", "[[:space:]]", -1)
			_, err := attributesFile.WriteString(fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text\n", encodedArg))
			if err != nil {
				Print("Error adding pattern %s", pattern)
				continue
			}
		}
		Print("Tracking %s", pattern)

		for _, f := range gittracked {
			if trackVerboseLoggingFlag || trackDryRunFlag {
				Print("Git LFS: touching %s", f)
			}

			if !trackDryRunFlag {
				now := time.Now()
				err := longpathos.Chtimes(f, now, now)
				if err != nil {
					LoggedError(err, "Error marking %q modified", f)
					continue
				}
			}
		}
	}
}

type mediaPattern struct {
	Pattern string
	Source  string
}

func findPatterns() []mediaPattern {
	var patterns []mediaPattern

	for _, path := range findAttributeFiles() {
		attributes, err := longpathos.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(attributes)

		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "filter=lfs") {
				fields := strings.Fields(line)
				relfile, _ := filepath.Rel(config.LocalWorkingDir, path)
				pattern := fields[0]
				if reldir := filepath.Dir(relfile); len(reldir) > 0 {
					pattern = filepath.Join(reldir, pattern)
				}

				patterns = append(patterns, mediaPattern{Pattern: pattern, Source: relfile})
			}
		}
	}

	return patterns
}

func findAttributeFiles() []string {
	var paths []string

	repoAttributes := filepath.Join(config.LocalGitDir, "info", "attributes")
	if info, err := longpathos.Stat(repoAttributes); err == nil && !info.IsDir() {
		paths = append(paths, repoAttributes)
	}

	tools.FastWalkGitRepo(config.LocalWorkingDir, func(parentDir string, info os.FileInfo, err error) {
		if err != nil {
			tracerx.Printf("Error finding .gitattributes: %v", err)
			return
		}

		if info.IsDir() || info.Name() != ".gitattributes" {
			return
		}
		paths = append(paths, filepath.Join(parentDir, info.Name()))
	})

	return paths
}

func needsTrailingLinebreak(filename string) bool {
	file, err := longpathos.Open(filename)
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

// blocklistItem returns the name of the blocklist item preventing the given
// file-name from being tracked, or an empty string, if there is none.
func blocklistItem(name string) string {
	base := filepath.Base(name)

	for _, p := range prefixBlocklist {
		if strings.HasPrefix(base, p) {
			return p
		}
	}

	return ""
}

func init() {
	RegisterCommand("track", trackCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&trackVerboseLoggingFlag, "verbose", "v", false, "log which files are being tracked and modified")
		cmd.Flags().BoolVarP(&trackDryRunFlag, "dry-run", "d", false, "preview results of running `git lfs track`")
	})
}
