package locking

import (
	"regexp"
	"strings"
	"sync"

	"github.com/github/git-lfs/config"
)

var (
	// lockable patterns from .gitattributes
	cachedLockablePatterns []string
	cachedLockableMutex    sync.Mutex
)

// GetLockablePatterns returns a list of patterns in .gitattributes which are
// marked as lockable
func GetLockablePatterns() []string {
	cachedLockableMutex.Lock()
	defer cachedLockableMutex.Unlock()

	// Only load once
	if cachedLockablePatterns == nil {
		// Always make non-nil even if empty
		cachedLockablePatterns = make([]string, 0, 10)

		paths := config.GetAttributePaths()
		for _, p := range paths {
			if p.Lockable {
				cachedLockablePatterns = append(cachedLockablePatterns, p.Path)
			}
		}
	}

	return cachedLockablePatterns

}

// RefreshLockablePatterns causes us to re-read the .gitattributes and caches the result
func RefreshLockablePatterns() {
	cachedLockableMutex.Lock()
	defer cachedLockableMutex.Unlock()
	cachedLockablePatterns = nil
}

// IsFileLockable returns whether a specific file path is marked as Lockable,
// ie has the 'lockable' attribute in .gitattributes
// Lockable patterns are cached once for performance, unless you call RefreshLockablePatterns
// path should be relative to repository root
func IsFileLockable(path string) bool {
	patterns := GetLockablePatterns()
	for _, wildcard := range patterns {
		// Convert wildcards to regex
		regStr := "^" + regexp.QuoteMeta(wildcard)
		regStr = strings.Replace(regStr, "\\*", ".*", -1)
		regStr = strings.Replace(regStr, "\\?", ".", -1)
		reg := regexp.MustCompile(regStr)

		if reg.MatchString(path) {
			return true
		}
	}
	return false
}
