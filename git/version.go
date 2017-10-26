package git

import (
	"regexp"
	"strconv"
	"sync"

	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/rubyist/tracerx"
)

var (
	gitVersion   *string
	gitVersionMu sync.Mutex
)

func Version() (string, error) {
	gitVersionMu.Lock()
	defer gitVersionMu.Unlock()

	if gitVersion == nil {
		v, err := subprocess.SimpleExec("git", "version")
		gitVersion = &v
		return v, err
	}

	return *gitVersion, nil
}

// IsVersionAtLeast returns whether the git version is the one specified or higher
// argument is plain version string separated by '.' e.g. "2.3.1" but can omit minor/patch
func IsGitVersionAtLeast(ver string) bool {
	gitver, err := Version()
	if err != nil {
		tracerx.Printf("Error getting git version: %v", err)
		return false
	}
	return IsVersionAtLeast(gitver, ver)
}

// IsVersionAtLeast compares 2 version strings (ok to be prefixed with 'git version', ignores)
func IsVersionAtLeast(actualVersion, desiredVersion string) bool {
	// Capture 1-3 version digits, optionally prefixed with 'git version' and possibly
	// with suffixes which we'll ignore (e.g. unstable builds, MinGW versions)
	verregex := regexp.MustCompile(`(?:git version\s+)?(\d+)(?:.(\d+))?(?:.(\d+))?.*`)

	var atleast uint64
	// Support up to 1000 in major/minor/patch digits
	const majorscale = 1000 * 1000
	const minorscale = 1000

	if match := verregex.FindStringSubmatch(desiredVersion); match != nil {
		// Ignore errors as regex won't match anything other than digits
		major, _ := strconv.Atoi(match[1])
		atleast += uint64(major * majorscale)
		if len(match) > 2 {
			minor, _ := strconv.Atoi(match[2])
			atleast += uint64(minor * minorscale)
		}
		if len(match) > 3 {
			patch, _ := strconv.Atoi(match[3])
			atleast += uint64(patch)
		}
	}

	var actual uint64
	if match := verregex.FindStringSubmatch(actualVersion); match != nil {
		major, _ := strconv.Atoi(match[1])
		actual += uint64(major * majorscale)
		if len(match) > 2 {
			minor, _ := strconv.Atoi(match[2])
			actual += uint64(minor * minorscale)
		}
		if len(match) > 3 {
			patch, _ := strconv.Atoi(match[3])
			actual += uint64(patch)
		}
	}

	return actual >= atleast
}
