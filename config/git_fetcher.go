package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type GitFetcher struct {
	vmu  sync.RWMutex
	vals map[string]string
}

type GitConfig struct {
	Lines        []string
	OnlySafeKeys bool
}

func ReadGitConfig(configs ...*GitConfig) (gf *GitFetcher, extensions map[string]Extension, uniqRemotes map[string]bool) {
	vals := make(map[string]string)

	extensions = make(map[string]Extension)
	uniqRemotes = make(map[string]bool)

	for _, gc := range configs {
		uniqKeys := make(map[string]string)

		for _, line := range gc.Lines {
			pieces := strings.SplitN(line, "=", 2)
			if len(pieces) < 2 {
				continue
			}

			allowed := !gc.OnlySafeKeys
			key, val := strings.ToLower(pieces[0]), pieces[1]

			if origKey, ok := uniqKeys[key]; ok {
				if ShowConfigWarnings && vals[key] != val && strings.HasPrefix(key, gitConfigWarningPrefix) {
					fmt.Fprintf(os.Stderr, "WARNING: These git config values clash:\n")
					fmt.Fprintf(os.Stderr, "  git config %q = %q\n", origKey, vals[key])
					fmt.Fprintf(os.Stderr, "  git config %q = %q\n", pieces[0], val)
				}
			} else {
				uniqKeys[key] = pieces[0]
			}

			parts := strings.Split(key, ".")
			if len(parts) == 4 && parts[0] == "lfs" && parts[1] == "extension" {
				// prop: lfs.extension.<name>.<prop>
				name := parts[2]
				prop := parts[3]

				ext := extensions[name]
				ext.Name = name

				switch prop {
				case "clean":
					if gc.OnlySafeKeys {
						continue
					}
					ext.Clean = val
				case "smudge":
					if gc.OnlySafeKeys {
						continue
					}
					ext.Smudge = val
				case "priority":
					allowed = true
					p, err := strconv.Atoi(val)
					if err == nil && p >= 0 {
						ext.Priority = p
					}
				}

				extensions[name] = ext
			} else if len(parts) > 1 && parts[0] == "remote" {
				if gc.OnlySafeKeys && (len(parts) == 3 && parts[2] != "lfsurl") {
					continue
				}

				allowed = true
				remote := parts[1]
				uniqRemotes[remote] = remote == "origin"
			} else if len(parts) > 2 && parts[len(parts)-1] == "access" {
				allowed = true
			}

			if !allowed && keyIsUnsafe(key) {
				continue
			}

			vals[key] = val
		}
	}

	gf = &GitFetcher{vals: vals}

	return
}

func (g *GitFetcher) Get(key string) (val string) {
	g.vmu.RLock()
	defer g.vmu.RUnlock()

	return g.vals[key]
}

func keyIsUnsafe(key string) bool {
	for _, safe := range safeKeys {
		if safe == key {
			return false
		}
	}
	return true
}

var safeKeys = []string{
	"lfs.fetchexclude",
	"lfs.fetchinclude",
	"lfs.gitprotocol",
	"lfs.url",
}
