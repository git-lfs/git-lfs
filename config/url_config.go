package config

import (
	"fmt"
	"net/url"
	"regexp"
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
// rules in https://git-scm.com/docs/git-config#git-config-httplturlgt.
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
		key       string // The full configuration key
		hostScore int    // A score indicating the strength of the host match
		pathScore int    // A score indicating the strength of the path match
		userMatch int    // Whether we matched on a username. 1 for yes, else 0
	}

	searchURL, err := url.Parse(rawurl)
	if err != nil {
		return nil
	}

	config := c.git.All()

	re := regexp.MustCompile(fmt.Sprintf(`%s.(\S+).%s`, prefix, key))

	bestMatch := urlMatch{
		key:       "",
		hostScore: 0,
		pathScore: 0,
		userMatch: 0,
	}

	for k := range config {
		// Ensure we're examining the correct type of key and parse out the URL
		matches := re.FindStringSubmatch(k)
		if matches == nil {
			continue
		}
		configURL, err := url.Parse(matches[1])
		if err != nil {
			continue
		}

		match := urlMatch{
			key: k,
		}

		// Rule #1: Scheme must match exactly
		if searchURL.Scheme != configURL.Scheme {
			continue
		}

		// Rule #2: Hosts must match exactly, or through wildcards. More exact
		// matches should take priority over wildcard matches
		match.hostScore = compareHosts(searchURL.Hostname(), configURL.Hostname())

		if match.hostScore == 0 {
			continue
		}

		if match.hostScore < bestMatch.hostScore {
			continue
		}

		// Rule #3: Port Number must match exactly
		if portForURL(searchURL) != portForURL(configURL) {
			continue
		}

		// Rule #4: Configured path must match exactly, or as a prefix of
		// slash-delimited path elements
		match.pathScore = comparePaths(searchURL.Path, configURL.Path)

		if match.pathScore == 0 {
			continue
		}

		// Rule #5: Username must match exactly if present in the config.
		// If not present, config matches on any username but with lower
		// priority than an exact username match.
		if configURL.User != nil {
			if searchURL.User == nil {
				continue
			}

			if searchURL.User.Username() != configURL.User.Username() {
				continue
			}

			match.userMatch = 1
		}

		// Now combine our various scores to determine if we have found a best
		// match. Host score > path score > user score
		if match.hostScore > bestMatch.hostScore {
			bestMatch = match
			continue
		}

		if match.pathScore > bestMatch.pathScore {
			bestMatch = match
			continue
		}

		if match.pathScore == bestMatch.pathScore && match.userMatch > bestMatch.userMatch {
			bestMatch = match
			continue
		}
	}

	if bestMatch.key == "" {
		return nil
	}

	return c.git.GetAll(bestMatch.key)
}

func portForURL(u *url.URL) string {
	port := u.Port()
	if port != "" {
		return port
	}
	switch u.Scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	case "ssh":
		return "22"
	default:
		return ""
	}
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

	score := len(searchHost) + 1

	for i, subdomain := range searchHost {
		if configHost[i] == "*" {
			score--
			continue
		}

		if subdomain != configHost[i] {
			return 0
		}
	}

	return score
}

// comparePaths compares a path with a configuration path to determine a match.
// It returns an integer indicating the strength of the match, or 0 if the two
// paths did not match.
func comparePaths(rawSearchPath, rawConfigPath string) int {
	f := func(c rune) bool {
		return c == '/'
	}
	searchPath := strings.FieldsFunc(rawSearchPath, f)
	configPath := strings.FieldsFunc(rawConfigPath, f)

	if len(searchPath) < len(configPath) {
		return 0
	}

	// Start with a base score of 1, so we return something above 0 for a
	// zero-length path
	score := 1

	for i, element := range configPath {
		searchElement := searchPath[i]

		if element == searchElement {
			score += 2
			continue
		}

		if isDefaultLFSUrl(searchElement, searchPath, i+1) {
			if searchElement[0:len(searchElement)-4] == element {
				// Since we matched without the `.git` prefix, only add one
				// point to the score instead of 2
				score++
				continue
			}
		}

		return 0
	}

	return score
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
