package commands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/git/gitattr"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
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
	trackNoModifyAttrsFlag  bool
	trackNoExcludedFlag     bool
	trackFilenameFlag       bool
	trackJSONFlag           bool
)

func trackCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()
	setupWorkingCopy()

	if trackDryRunFlag {
		trackNoModifyAttrsFlag = true
	}

	if !cfg.Os.Bool("GIT_LFS_TRACK_NO_INSTALL_HOOKS", false) {
		installHooks(false)
	}

	if len(args) == 0 {
		listPatterns()
		return
	}

	if trackJSONFlag {
		Exit(tr.Tr.Get("--json option can't be combined with arguments"))
	}

	mp := gitattr.NewMacroProcessor()

	// Intentionally do _not_ consider global- and system-level
	// .gitattributes here.  Parse them still to expand any macros.
	git.GetSystemAttributePaths(mp, cfg.Os)
	git.GetRootAttributePaths(mp, cfg.Git)
	knownPatterns := git.GetAttributePaths(mp, cfg.LocalWorkingDir(), cfg.LocalGitDir())
	lineEnd := getAttributeLineEnding(knownPatterns)
	if len(lineEnd) == 0 {
		lineEnd = gitLineEnding(cfg.Git)
	}

	wd, _ := tools.Getwd()
	wd = tools.ResolveSymlinks(wd)
	relpath, err := filepath.Rel(cfg.LocalWorkingDir(), wd)
	if err != nil {
		Exit(tr.Tr.Get("Current directory %q outside of Git working directory %q.", wd, cfg.LocalWorkingDir()))
	}

	changedAttribLines := make(map[string]string)
	var readOnlyPatterns []string
	var writeablePatterns []string
ArgsLoop:
	for _, unsanitizedPattern := range args {
		pattern := tools.TrimCurrentPrefix(cleanRootPath(unsanitizedPattern))

		// Generate the new / changed attrib line for merging
		var encodedArg string
		if trackFilenameFlag {
			encodedArg = escapeGlobCharacters(pattern)
			pattern = escapeGlobCharacters(pattern)
		} else {
			encodedArg = escapeAttrPattern(pattern)
		}

		if !trackNoModifyAttrsFlag {
			for _, known := range knownPatterns {
				if unescapeAttrPattern(known.Path) == path.Join(relpath, pattern) &&
					((trackLockableFlag && known.Lockable) || // enabling lockable & already lockable (no change)
						(trackNotLockableFlag && !known.Lockable) || // disabling lockable & not lockable (no change)
						(!trackLockableFlag && !trackNotLockableFlag)) { // leave lockable as-is in all cases
					Print(tr.Tr.Get("%q already supported", pattern))
					continue ArgsLoop
				}
			}
		}

		lockableArg := ""
		if trackLockableFlag { // no need to test trackNotLockableFlag, if we got here we're disabling
			lockableArg = " " + git.LockableAttrib
		}

		changedAttribLines[pattern] = fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text%v%s", encodedArg, lockableArg, lineEnd)

		if trackLockableFlag {
			readOnlyPatterns = append(readOnlyPatterns, pattern)
		} else {
			writeablePatterns = append(writeablePatterns, pattern)
		}

		Print(tr.Tr.Get("Tracking %q", unescapeAttrPattern(encodedArg)))
	}

	// Now read the whole local attributes file and iterate over the contents,
	// replacing any lines where the values have changed, and appending new lines
	// change this:

	var (
		attribContents []byte
		attributesFile *os.File
	)
	if !trackNoModifyAttrsFlag {
		attribContents, err = os.ReadFile(".gitattributes")
		// it's fine for file to not exist
		if err != nil && !os.IsNotExist(err) {
			Print(tr.Tr.Get("Error reading '.gitattributes' file"))
			return
		}
		// Re-generate the file with merge of old contents and new (to deal with changes)
		attributesFile, err = os.OpenFile(".gitattributes", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0660)
		if err != nil {
			Print(tr.Tr.Get("Error opening '.gitattributes' file"))
			return
		}
		defer attributesFile.Close()

		if len(attribContents) > 0 {
			scanner := bufio.NewScanner(bytes.NewReader(attribContents))
			for scanner.Scan() {
				line := scanner.Text()
				fields := strings.Fields(line)
				if len(fields) < 1 {
					continue
				}

				pattern := unescapeAttrPattern(fields[0])
				if newline, ok := changedAttribLines[pattern]; ok {
					// Replace this line (newline already embedded)
					attributesFile.WriteString(newline)
					// Remove from map so we know we don't have to add it to the end
					delete(changedAttribLines, pattern)
				} else {
					// Write line unchanged (replace newline)
					attributesFile.WriteString(line + lineEnd)
				}
			}

			// Our method of writing also made sure there's always a newline at end
		}
	}

	modified := false
	sawError := false
	// Any items left in the map, write new lines at the end of the file
	// Note this is only new patterns, not ones which changed locking flags
	for pattern, newline := range changedAttribLines {
		// Also, for any new patterns we've added, make sure any existing git
		// tracked files have their timestamp updated so they will now show as
		// modified note this is relative to current dir which is how we write
		// .gitattributes deliberately not done in parallel as a chan because
		// we'll be marking modified
		//
		// NOTE: `git ls-files` does not do well with leading slashes.
		// Since all `git-lfs track` calls are relative to the root of
		// the repository, the leading slash is simply removed for its
		// implicit counterpart.
		if trackVerboseLoggingFlag {
			Print(tr.Tr.Get("Searching for files matching pattern: %s", pattern))
		}

		gittracked, err := git.GetTrackedFiles(pattern)
		if err != nil {
			Exit(tr.Tr.Get("Error getting tracked files for %q: %s", pattern, err))
		}

		if trackVerboseLoggingFlag {
			Print(tr.Tr.Get("Found %d files previously added to Git matching pattern: %s", len(gittracked), pattern))
		}

		var matchedBlocklist bool
		for _, f := range gittracked {
			if forbidden := blocklistItem(f); forbidden != "" {
				Print(tr.Tr.Get("Pattern '%s' matches forbidden file '%s'. If you would like to track %s, modify '.gitattributes' manually.", pattern, f, f))
				matchedBlocklist = true
			}
		}
		if matchedBlocklist {
			continue
		}

		if !trackNoModifyAttrsFlag {
			// Newline already embedded
			attributesFile.WriteString(newline)
		}
		modified = true

		for _, f := range gittracked {
			if trackVerboseLoggingFlag || trackDryRunFlag {
				Print(tr.Tr.Get("Touching %q", f))
			}

			if !trackDryRunFlag {
				now := time.Now()
				err := os.Chtimes(f, now, now)
				if err != nil {
					LoggedError(err, tr.Tr.Get("Error marking %q modified: %s", f, err))
					sawError = true
					continue
				}
			}
		}
	}

	// now flip read-only mode based on lockable / not lockable changes
	lockClient := newLockClient()
	err = lockClient.FixFileWriteFlagsInDir(relpath, readOnlyPatterns, writeablePatterns)
	if err != nil {
		LoggedError(err, tr.Tr.Get("Error changing lockable file permissions: %s", err))
		sawError = true
	}

	if sawError {
		os.Exit(2)
	}
	// If we didn't modify things, but that's because the patterns
	// were already supported, don't return an error, since what the
	// user wanted has already been done.
	// Otherwise, if we didn't modify things but only because the
	// patterns were disallowed, return an error.
	if !modified && len(changedAttribLines) > 0 {
		os.Exit(1)
	}
}

type PatternData struct {
	Pattern  string `json:"pattern"`
	Source   string `json:"source"`
	Lockable bool   `json:"lockable"`
	Tracked  bool   `json:"tracked"`
}

func listPatterns() {
	knownPatterns, err := getAllKnownPatterns()
	if err != nil {
		Exit("unable to list patterns: %s", err)
	}
	if trackJSONFlag {
		patterns := struct {
			Patterns []PatternData `json:"patterns"`
		}{Patterns: make([]PatternData, 0, len(knownPatterns))}
		for _, p := range knownPatterns {
			patterns.Patterns = append(patterns.Patterns, PatternData{
				Pattern:  p.Path,
				Source:   p.Source.String(),
				Tracked:  p.Tracked,
				Lockable: p.Lockable,
			})
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", " ")
		err := encoder.Encode(patterns)
		if err != nil {
			ExitWithError(err)
		}
		return
	}

	if len(knownPatterns) < 1 {
		return
	}

	Print(tr.Tr.Get("Listing tracked patterns"))
	for _, t := range knownPatterns {
		if t.Lockable {
			// TRANSLATORS: Leading spaces here should be preserved.
			Print(tr.Tr.Get("    %s [lockable] (%s)", t.Path, t.Source))
		} else if t.Tracked {
			Print("    %s (%s)", t.Path, t.Source)
		}
	}

	if trackNoExcludedFlag {
		return
	}

	Print(tr.Tr.Get("Listing excluded patterns"))
	for _, t := range knownPatterns {
		if !t.Tracked && !t.Lockable {
			Print("    %s (%s)", t.Path, t.Source)
		}
	}
}

func getAllKnownPatterns() ([]git.AttributePath, error) {
	mp := gitattr.NewMacroProcessor()

	// Parse these in this order so that macros in one file are properly
	// expanded when referred to in a later file, then order them in the
	// order we want.
	systemPatterns, err := git.GetSystemAttributePaths(mp, cfg.Os)
	if err != nil {
		return nil, err
	}
	globalPatterns := git.GetRootAttributePaths(mp, cfg.Git)
	knownPatterns := git.GetAttributePaths(mp, cfg.LocalWorkingDir(), cfg.LocalGitDir())
	knownPatterns = append(knownPatterns, globalPatterns...)
	knownPatterns = append(knownPatterns, systemPatterns...)

	return knownPatterns, nil
}

func getAttributeLineEnding(attribs []git.AttributePath) string {
	for _, a := range attribs {
		if a.Source.Path == ".gitattributes" {
			return a.Source.LineEnding
		}
	}
	return ""
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

var (
	trackEscapePatterns = map[string]string{
		" ": "[[:space:]]",
		"#": "\\#",
	}
	trackEscapeStrings = []string{"*", "[", "]", "?"}
)

func escapeGlobCharacters(s string) string {
	var escaped string
	if runtime.GOOS == "windows" {
		escaped = strings.Replace(s, `\`, "/", -1)
	} else {
		escaped = strings.Replace(s, `\`, `\\`, -1)
	}

	for _, ch := range trackEscapeStrings {
		escaped = strings.Replace(escaped, ch, fmt.Sprintf("\\%s", ch), -1)
	}

	for from, to := range trackEscapePatterns {
		escaped = strings.Replace(escaped, from, to, -1)
	}
	return escaped
}

func escapeAttrPattern(s string) string {
	var escaped string
	if runtime.GOOS == "windows" {
		escaped = strings.Replace(s, `\`, "/", -1)
	} else {
		escaped = strings.Replace(s, `\`, `\\`, -1)
	}

	for from, to := range trackEscapePatterns {
		escaped = strings.Replace(escaped, from, to, -1)
	}

	return escaped
}

func unescapeAttrPattern(escaped string) string {
	var unescaped string = escaped

	for to, from := range trackEscapePatterns {
		unescaped = strings.Replace(unescaped, from, to, -1)
	}

	if runtime.GOOS != "windows" {
		unescaped = strings.Replace(unescaped, `\\`, `\`, -1)
	}

	return unescaped
}

func init() {
	RegisterCommand("track", trackCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&trackLockableFlag, "lockable", "l", false, "make pattern lockable, i.e. read-only unless locked")
		cmd.Flags().BoolVarP(&trackNotLockableFlag, "not-lockable", "", false, "remove lockable attribute from pattern")
		cmd.Flags().BoolVarP(&trackVerboseLoggingFlag, "verbose", "v", false, "log which files are being tracked and modified")
		cmd.Flags().BoolVarP(&trackDryRunFlag, "dry-run", "d", false, "preview results of running `git lfs track`")
		cmd.Flags().BoolVarP(&trackNoModifyAttrsFlag, "no-modify-attrs", "", false, "skip modifying .gitattributes file")
		cmd.Flags().BoolVarP(&trackNoExcludedFlag, "no-excluded", "", false, "skip listing excluded paths")
		cmd.Flags().BoolVarP(&trackFilenameFlag, "filename", "", false, "treat this pattern as a literal filename")
		cmd.Flags().BoolVarP(&trackJSONFlag, "json", "", false, "print output in JSON")
	})
}
