package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/git-lfs/git-lfs/git"
)

// gitEnvironment is an implementation of the Environment which wraps the legacy
// behavior or `*config.Configuration.loadGitConfig()`.
//
// It is functionally equivelant to call `cfg.loadGitConfig()` before calling
// methods on the Environment type.
type gitEnvironment struct {
	// git is the Environment which gitEnvironment wraps.
	git Environment

	// gitConfig can fetch or modify the current Git config and track the Git
	// version.
	gitConfig *git.Configuration

	// config is the *Configuration instance which is mutated by
	// `loadGitConfig`.
	config  *Configuration
	loading sync.Mutex // guards initialization of gitConfig and remotes
}

// Get is shorthand for calling the loadGitConfig, and then returning
// `g.git.Get(key)`.
func (g *gitEnvironment) Get(key string) (val string, ok bool) {
	g.loadGitConfig()

	return g.git.Get(key)
}

// Get is shorthand for calling the loadGitConfig, and then returning
// `g.git.GetAll(key)`.
func (g *gitEnvironment) GetAll(key string) []string {
	g.loadGitConfig()

	return g.git.GetAll(key)
}

// Get is shorthand for calling the loadGitConfig, and then returning
// `g.git.Bool(key, def)`.
func (g *gitEnvironment) Bool(key string, def bool) (val bool) {
	g.loadGitConfig()

	return g.git.Bool(key, def)
}

// Get is shorthand for calling the loadGitConfig, and then returning
// `g.git.Int(key, def)`.
func (g *gitEnvironment) Int(key string, def int) (val int) {
	g.loadGitConfig()

	return g.git.Int(key, def)
}

// All returns a copy of all the key/value pairs for the current git config.
func (g *gitEnvironment) All() map[string][]string {
	g.loadGitConfig()

	return g.git.All()
}

// loadGitConfig reads and parses the .gitconfig by calling ReadGitConfig. It
// also sets values on the configuration instance `g.config`.
//
// If loadGitConfig has already been called, this method will bail out early,
// and return false. Otherwise it will preform the entire parse and return true.
//
// loadGitConfig is safe to call across multiple goroutines.
func (g *gitEnvironment) loadGitConfig() bool {
	g.loading.Lock()
	defer g.loading.Unlock()

	if g.git != nil {
		return false
	}

	sources, err := g.gitConfig.Sources(filepath.Join(LocalWorkingDir, ".lfsconfig"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading git config: %s\n", err)
	}

	gf, extensions, uniqRemotes := ReadGitConfig(sources...)

	g.git = EnvironmentOf(gf)

	g.config.extensions = extensions

	g.config.remotes = make([]string, 0, len(uniqRemotes))
	for remote, isOrigin := range uniqRemotes {
		if isOrigin {
			continue
		}
		g.config.remotes = append(g.config.remotes, remote)
	}

	return true
}
