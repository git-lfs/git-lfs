package gitmedia

import (
	"github.com/pelletier/go-toml"
	"os"
	"path/filepath"
)

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
		config = &Configuration{}
		readToml(config)
	}

	return config
}

func readToml(config *Configuration) {
	tomlPath := filepath.Join(LocalWorkingDir, ".gitmedia")
	stat, _ := os.Stat(tomlPath)
	if stat != nil {
		readTomlFile(tomlPath, config)
	}
}

func readTomlFile(path string, config *Configuration) {
	tomlConfig, err := toml.LoadFile(path)
	if err != nil {
		Panic(err, "Error reading TOML file: %s", path)
	}

	if endpoint, ok := tomlConfig.Get("endpoint").(string); ok {
		config.Endpoint = endpoint
	}
}
