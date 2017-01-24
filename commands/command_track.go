package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	prefixBlocklist = []string{
		".git", ".lfs",
	}

	trackLockableFlag       bool
	trackNotLockableFlag    bool
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
	knownPatterns := git.GetAttributePaths(config.LocalWorkingDir, config.LocalGitDir)

	if len(args) == 0 {
		Print("Listing tracked patterns")
		for _, t := range knownPatterns {
			if t.Lockable {
				Print("    %s [lockable] (%s)", t.Path, t.Source)
			} else {
				Print("    %s (%s)", t.Path, t.Source)
			}
		}
		return
	}

	wd, _ := os.Getwd()
	relpath, err := filepath.Rel(config.LocalWorkingDir, wd)
	if err != nil {
		Exit("Current directory %q outside of git working directory %q.", wd, config.LocalWorkingDir)
	}

	changedAttribLines := make(map[string]string)
	var readOnlyPatterns []string
	var writeablePatterns []string
ArgsLoop:
	for _, unsanitizedPattern := range args {
		pattern := cleanRootPath(unsanitizedPattern)
		for _, known := range knownPatterns {
			if known.Path == filepath.Join(relpath, pattern) &&
				((trackLockableFlag && known.Lockable) || // enabling lockable & already lockable (no change)
					(trackNotLockableFlag && !known.Lockable) || // disabling lockable & not lockable (no change)
					(!trackLockableFlag && !trackNotLockableFlag)) { // leave lockable as-is in all cases
				Print("%s already supported", pattern)
				continue ArgsLoop
			}
		}

		// Generate the new / changed attrib line for merging
		encodedArg := strings.Replace(pattern, " ", "[[:space:]]", -1)
		lockableArg := ""
		if trackLockableFlag { // no need to test trackNotLockableFlag, if we got here we're disabling
			lockableArg = " " + git.LockableAttrib
		}

		changedAttribLines[pattern] = fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text%v\n", encodedArg, lockableArg)

		if trackLockableFlag {
			readOnlyPatterns = append(readOnlyPatterns, pattern)
		} else {
			writeablePatterns = append(writeablePatterns, pattern)
		}

		Print("Tracking %s", pattern)

	}

	// Now read the whole local attributes file and iterate over the contents,
	// replacing any lines where the values have changed, and appending new lines
	// change this:

	attribContents, err := ioutil.ReadFile(".gitattributes")
	// it's fine for file to not exist
	if err != nil && !os.IsNotExist(err) {
		Print("Error reading .gitattributes file")
		return
	}
	// Re-generate the file with merge of old contents and new (to deal with changes)
	attributesFile, err := os.OpenFile(".gitattributes", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0660)
	if err != nil {
		Print("Error opening .gitattributes file")
		return
	}
	defer attributesFile.Close()

	if len(attribContents) > 0 {
		scanner := bufio.NewScanner(bytes.NewReader(attribContents))
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			pattern := fields[0]
			if newline, ok := changedAttribLines[pattern]; ok {
				// Replace this line (newline already embedded)
				attributesFile.WriteString(newline)
				// Remove from map so we know we don't have to add it to the end
				delete(changedAttribLines, pattern)
			} else {
				// Write line unchanged (replace newline)
				attributesFile.WriteString(line + "\n")
			}
		}

		// Our method of writing also made sure there's always a newline at end
	}

	// Any items left in the map, write new lines at the end of the file
	// Note this is only new patterns, not ones which changed locking flags
	for pattern, newline := range changedAttribLines {
		// Newline already embedded
		attributesFile.WriteString(newline)

		// Also, for any new patterns we've added, make sure any existing git
		// tracked files have their timestamp updated so they will now show as
		// modifed note this is relative to current dir which is how we write
		// .gitattributes deliberately not done in parallel as a chan because
		// we'll be marking modified
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

		for _, f := range gittracked {
			if trackVerboseLoggingFlag || trackDryRunFlag {
				Print("Git LFS: touching %s", f)
			}

			if !trackDryRunFlag {
				now := time.Now()
				err := os.Chtimes(f, now, now)
				if err != nil {
					LoggedError(err, "Error marking %q modified", f)
					continue
				}
			}
		}
	}
	// now flip read-only mode based on lockable / not lockable changes
	lockClient := newLockClient(cfg.CurrentRemote)
	err = lockClient.FixFileWriteFlagsInDir(relpath, readOnlyPatterns, writeablePatterns)
	if err != nil {
		LoggedError(err, "Error changing lockable file permissions")
	}
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
		cmd.Flags().BoolVarP(&trackLockableFlag, "lockable", "l", false, "make pattern lockable, i.e. read-only unless locked")
		cmd.Flags().BoolVarP(&trackNotLockableFlag, "not-lockable", "", false, "remove lockable attribute from pattern")
		cmd.Flags().BoolVarP(&trackVerboseLoggingFlag, "verbose", "v", false, "log which files are being tracked and modified")
		cmd.Flags().BoolVarP(&trackDryRunFlag, "dry-run", "d", false, "preview results of running `git lfs track`")
	})
}
