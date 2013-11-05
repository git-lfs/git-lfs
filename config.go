package gitmedia

type Configuration struct {
	Endpoint string
}

var config *Configuration

// Config gets the git media configuration for the current repository.  It
// reads .gitmedia, which is a toml file.
//
// https://github.com/mojombo/toml
func Config() *Configuration {
	if config == nil {
		config = &Configuration{"http://localhost:8080"}
	}

	return config
}
