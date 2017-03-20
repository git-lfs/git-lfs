package config

// gitEnvironment is an implementation of the Environment which wraps the legacy
// behavior or `*config.Configuration.loadGitConfig()`.
//
// It is functionally equivelant to call `cfg.loadGitConfig()` before calling
// methods on the Environment type.
type gitEnvironment struct {
	// git is the Environment which gitEnvironment wraps.
	git Environment
	// config is the *Configuration instance which is mutated by
	// `loadGitConfig`.
	config *Configuration
}

// Get is shorthand for calling the loadGitConfig, and then returning
// `g.git.Get(key)`.
func (g *gitEnvironment) Get(key string) (val string, ok bool) {
	g.loadGitConfig()

	return g.git.Get(key)
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
func (g *gitEnvironment) All() map[string]string {
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
	g.config.loading.Lock()
	defer g.config.loading.Unlock()

	if g.git != nil {
		return false
	}

	gf, extensions, uniqRemotes := ReadGitConfig(getGitConfigs()...)

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
