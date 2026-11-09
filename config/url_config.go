package config

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

type URLConfig struct {
	git Environment
}

func NewURLConfig(git Environment) *URLConfig {
	if git == nil {
		git = EnvironmentOf(make(mapFetcher))
	}

	return &URLConfig{
		git: git,
	}
}

// Get retrieves a `http.{url}.{key}` for the given key and urls, following the
// rules in https://git-scm.com/docs/git-config#Documentation/git-config.txt-httplturlgt.
// The value for `http.{key}` is returned as a fallback if no config keys are
// set for the given urls.
func (c *URLConfig) Get(prefix, rawurl, key string) (string, bool) {
	if c == nil {
		return "", false
	}

	key = strings.ToLower(key)
	prefix = strings.ToLower(prefix)
	if v := c.getAll(prefix, rawurl, key); len(v) > 0 {
		return v[len(v)-1], true
	}
	return c.git.Get(strings.Join([]string{prefix, key}, "."))
}

func (c *URLConfig) GetAll(prefix, rawurl, key string) []string {
	if c == nil {
		return nil
	}

	key = strings.ToLower(key)
	prefix = strings.ToLower(prefix)
	if v := c.getAll(prefix, rawurl, key); len(v) > 0 {
		return v
	}
	return c.git.GetAll(strings.Join([]string{prefix, key}, "."))
}

func (c *URLConfig) Bool(prefix, rawurl, key string, def bool) bool {
	s, _ := c.Get(prefix, rawurl, key)
	return Bool(s, def)
}

func (c *URLConfig) getAll(prefix, rawurl, key string) []string {
	type urlMatch struct {
		values    []string // The values associated with the configuration key
		hostScore int      // A score indicating the strength of the host match
		pathScore int      // A score indicating the strength of the path match
		userMatch int      // Whether we matched on a username. 1 for yes, else 0
	}

	searchURL, err := normalizeURL(rawurl, false)
	if err != nil {
		return nil
	}

	config, ordered := c.configEntries()
	re := regexp.MustCompile(fmt.Sprintf(`\A%s\.(\S+)\.%s\z`, regexp.QuoteMeta(prefix), regexp.QuoteMeta(key)))

	bestMatch := urlMatch{}
	foundMatch := false

	for _, entry := range config {
		// Ensure we're examining the correct type of key and parse out the URL
		matches := re.FindStringSubmatch(entry.Key)
		if matches == nil {
			continue
		}
		configURL, err := normalizeURL(matches[1], true)
		if err != nil {
			continue
		}

		match := urlMatch{
			values: entry.Values,
		}

		// Rule #1: Scheme must match exactly
		if searchURL.scheme != configURL.scheme {
			continue
		}

		// Rule #2: Hosts must match exactly, or through wildcards. More exact
		// matches should take priority over wildcard matches
		match.hostScore = compareHosts(searchURL.hostname, configURL.hostname)

		if match.hostScore == 0 {
			continue
		}

		// Rule #3: Port Number must match exactly
		if searchURL.port != configURL.port {
			continue
		}

		// Rule #4: Configured path must match exactly, or as a prefix of
		// slash-delimited path elements
		match.pathScore = comparePaths(searchURL.path, configURL.path)

		if match.pathScore == 0 {
			continue
		}

		// Rule #5: Username must match exactly if present in the config.
		// If not present, config matches on any username but with lower
		// priority than an exact username match.
		if configURL.hasUserinfo && configURL.userinfo != "" {
			if !searchURL.hasUserinfo || searchURL.userinfo == "" {
				continue
			}

			if searchURL.username != configURL.username {
				continue
			}

			match.userMatch = 1
		}

		// Combine scores in the same order Git uses. When entries are in source
		// order, an equally specific later entry replaces the earlier one.
		better := !foundMatch ||
			match.hostScore > bestMatch.hostScore ||
			(match.hostScore == bestMatch.hostScore && match.pathScore > bestMatch.pathScore) ||
			(match.hostScore == bestMatch.hostScore && match.pathScore == bestMatch.pathScore && match.userMatch > bestMatch.userMatch)
		equal := foundMatch && match.hostScore == bestMatch.hostScore &&
			match.pathScore == bestMatch.pathScore && match.userMatch == bestMatch.userMatch
		if better || (ordered && equal) {
			bestMatch = match
			foundMatch = true
		}
	}

	if !foundMatch {
		return nil
	}

	return bestMatch.values
}

func (c *URLConfig) configEntries() ([]EnvironmentEntry, bool) {
	if entries := c.git.SortedAll(); entries != nil {
		return entries, true
	}

	all := c.git.All()
	keys := make([]string, 0, len(all))
	for key := range all {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]EnvironmentEntry, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, EnvironmentEntry{Key: key, Values: all[key]})
	}
	return entries, false
}

// compareHosts compares a hostname with a configuration hostname to determine
// a match. It returns an integer indicating the strength of the match, or 0 if
// the two hostnames did not match.
func compareHosts(searchHostname, configHostname string) int {
	searchHost := strings.Split(searchHostname, ".")
	configHost := strings.Split(configHostname, ".")

	if len(searchHost) != len(configHost) {
		return 0
	}

	for i, subdomain := range searchHost {
		if configHost[i] == "*" {
			continue
		}

		if subdomain != configHost[i] {
			return 0
		}
	}

	return len(configHostname) + 1
}

// comparePaths compares a path with a configuration path to determine a match.
// It returns an integer indicating the strength of the match, or 0 if the two
// paths did not match.
func comparePaths(rawSearchPath, rawConfigPath string) int {
	if score := pathMatchScore(rawSearchPath, rawConfigPath); score > 0 {
		return score
	}

	parts := strings.Split(rawSearchPath, slash)
	for i := 0; i+2 < len(parts); i++ {
		if !strings.HasSuffix(parts[i], gitExt) || parts[i+1] != infoPart || parts[i+2] != lfsPart {
			continue
		}

		parts[i] = strings.TrimSuffix(parts[i], gitExt)
		return pathMatchScore(strings.Join(parts, slash), rawConfigPath)
	}
	return 0
}

func pathMatchScore(searchPath, configPath string) int {
	if configPath == slash {
		return 1
	}

	configPath = strings.TrimSuffix(configPath, slash)
	if searchPath == configPath || strings.HasPrefix(searchPath, configPath+slash) {
		return len(configPath) + 1
	}
	return 0
}

func (c *URLConfig) hostsAndPaths(rawurl string) (hosts, paths []string) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, nil
	}

	return c.hosts(u), c.paths(u.Path)
}

func (c *URLConfig) hosts(u *url.URL) []string {
	hosts := make([]string, 0, 1)

	if u.User != nil {
		hosts = append(hosts, fmt.Sprintf("%s://%s@%s", u.Scheme, u.User.Username(), u.Host))
	}
	hosts = append(hosts, fmt.Sprintf("%s://%s", u.Scheme, u.Host))

	return hosts
}

func (c *URLConfig) paths(path string) []string {
	pLen := len(path)
	if pLen <= 2 {
		return nil
	}

	end := pLen
	if strings.HasSuffix(path, slash) {
		end--
	}
	return strings.Split(path[1:end], slash)
}

const (
	gitExt   = ".git"
	infoPart = "info"
	lfsPart  = "lfs"
	slash    = "/"
)

func isDefaultLFSUrl(path string, parts []string, index int) bool {
	if len(path) < 5 {
		return false // shorter than ".git"
	}

	if !strings.HasSuffix(path, gitExt) {
		return false
	}

	if index > len(parts)-2 {
		return false
	}

	return parts[index] == infoPart && parts[index+1] == lfsPart
}
