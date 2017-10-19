package config

import (
	"sync"
)

// delayedEnvironment is an implementation of the Environment which wraps the legacy
// behavior of `*config.Configuration.loadGitConfig()`.
//
// It is functionally equivelant to call `cfg.loadGitConfig()` before calling
// methods on the Environment type.
type delayedEnvironment struct {
	env      Environment
	loading  sync.Mutex
	callback func() Environment
}

// Get is shorthand for calling the e.Load(), and then returning
// `e.env.Get(key)`.
func (e *delayedEnvironment) Get(key string) (string, bool) {
	e.Load()
	return e.env.Get(key)
}

// Get is shorthand for calling the e.Load(), and then returning
// `e.env.GetAll(key)`.
func (e *delayedEnvironment) GetAll(key string) []string {
	e.Load()
	return e.env.GetAll(key)
}

// Get is shorthand for calling the e.Load(), and then returning
// `e.env.Bool(key, def)`.
func (e *delayedEnvironment) Bool(key string, def bool) bool {
	e.Load()
	return e.env.Bool(key, def)
}

// Get is shorthand for calling the e.Load(), and then returning
// `e.env.Int(key, def)`.
func (e *delayedEnvironment) Int(key string, def int) int {
	e.Load()
	return e.env.Int(key, def)
}

// All returns a copy of all the key/value pairs for the current git config.
func (e *delayedEnvironment) All() map[string][]string {
	e.Load()
	return e.env.All()
}

// Load reads and parses the .gitconfig by calling ReadGitConfig. It
// also sets values on the configuration instance `g.config`.
//
// If Load has already been called, this method will bail out early,
// and return false. Otherwise it will preform the entire parse and return true.
//
// Load is safe to call across multiple goroutines.
func (e *delayedEnvironment) Load() {
	e.loading.Lock()
	defer e.loading.Unlock()

	if e.env != nil {
		return
	}

	e.env = e.callback()
}
