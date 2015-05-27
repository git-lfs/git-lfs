package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type Config struct {
	BasePath           string
	AllowAbsolutePaths bool
	EnableDeltaReceive bool
	EnableDeltaSend    bool
	DeltaCachePath     string
	DeltaSizeLimit     int64
}

const defaultDeltaSizeLimit int64 = 2 * 1024 * 1024 * 1024

func NewConfig() *Config {
	return &Config{
		AllowAbsolutePaths: false,
		EnableDeltaReceive: true,
		EnableDeltaSend:    true,
		DeltaSizeLimit:     defaultDeltaSizeLimit, // 2GB
	}
}
func LoadConfig() *Config {
	// Support gitconfig-style configuration in:
	// Linux/Mac:
	// ~/.git-lfs-serve
	// /etc/git-lfs-serve.conf
	// Windows:
	// %USERPROFILE%\git-lfs-serve.ini
	// %PROGRAMDATA%\git-lfs\git-lfs-serve.ini

	var configFiles []string
	usr, err := user.Current()
	if err != nil {
		fmt.Fprint(os.Stderr, "Warning, couldn't locate home directory: %v", err.Error())
		return NewConfig()
	}
	home := usr.HomeDir

	// Order is important; read global config files first then user config files so settings
	// in the latter override the former
	if runtime.GOOS == "windows" {
		progdata := os.Getenv("PROGRAMDATA")
		if progdata != "" {
			configFiles = append(configFiles, filepath.Join(progdata, "git-lfs-serve.ini"))
		}
		if home != "" {
			configFiles = append(configFiles, filepath.Join(home, "git-lfs-serve.ini"))
		}
	} else {
		configFiles = append(configFiles, "/etc/git-lfs-serve.conf")
		if home != "" {
			configFiles = append(configFiles, filepath.Join(home, ".git-lfs-serve"))
		}
	}

	var settings = make(map[string]string)
	for _, conf := range configFiles {
		confsettings, err := ReadConfigFile(conf)
		if err == nil {
			for key, val := range confsettings {
				settings[key] = val
			}
		}
	}

	// Convert to Config
	cfg := NewConfig()
	if v := settings["base-path"]; v != "" {
		cfg.BasePath = filepath.Clean(v)
	}
	if v := strings.ToLower(settings["allow-absolute-paths"]); v != "" {
		if v == "true" {
			cfg.AllowAbsolutePaths = true
		} else if v == "false" {
			cfg.AllowAbsolutePaths = false
		}
	}
	if v := strings.ToLower(settings["enable-delta-receive"]); v != "" {
		if v == "true" {
			cfg.EnableDeltaReceive = true
		} else if v == "false" {
			cfg.EnableDeltaReceive = false
		}
	}
	if v := strings.ToLower(settings["enable-delta-send"]); v != "" {
		if v == "true" {
			cfg.EnableDeltaSend = true
		} else if v == "false" {
			cfg.EnableDeltaSend = false
		}
	}
	if v := settings["delta-cache-path"]; v != "" {
		cfg.DeltaCachePath = v
	}

	if cfg.DeltaCachePath == "" && cfg.BasePath != "" {
		cfg.DeltaCachePath = filepath.Join(cfg.BasePath, ".deltacache")
	}

	if v := settings["delta-size-limit"]; v != "" {
		var err error
		cfg.DeltaSizeLimit, err = strconv.ParseInt(v, 0, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid configuration: delta-size-limit=%v\n", v)
			cfg.DeltaSizeLimit = defaultDeltaSizeLimit
		}
	}

	return cfg
}

// Read a specific .gitconfig-formatted config file
// Returns a map of setting=value, where group levels are indicated by dot-notation
// e.g. git-lob.logfile=blah
// all keys are converted to lower case for easier matching
func ReadConfigFile(filepath string) (map[string]string, error) {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0644)
	if err != nil {
		return make(map[string]string), err
	}
	defer f.Close()

	// Need the directory for relative path includes
	dir := path.Dir(filepath)
	return ReadConfigStream(f, dir)

}
func ReadConfigStream(in io.Reader, dir string) (map[string]string, error) {
	ret := make(map[string]string, 10)
	sectionRegex := regexp.MustCompile(`^\[(.*)\]$`)                    // simple section regex ie [section]
	namedSectionRegex := regexp.MustCompile(`^\[(.*)\s+\"(.*)\"\s*\]$`) // named section regex ie [section "name"]

	scanner := bufio.NewScanner(in)
	var currentSection string
	var currentSectionName string
	for scanner.Scan() {
		// Reads lines by default, \n is already stripped
		line := strings.TrimSpace(scanner.Text())
		// Detect comments - discard any of the line after the comment but keep anything before
		commentPos := strings.IndexAny(line, "#;")
		if commentPos != -1 {
			// skip comments
			if commentPos == 0 {
				continue
			} else {
				// just strip rest of line after the comment
				line = strings.TrimSpace(line[0:commentPos])
				if len(line) == 0 {
					continue
				}
			}
		}

		// Check for sections
		if secmatch := sectionRegex.FindStringSubmatch(line); secmatch != nil {
			// named section? [section "name"]
			if namedsecmatch := namedSectionRegex.FindStringSubmatch(line); namedsecmatch != nil {
				// Named section
				currentSection = namedsecmatch[1]
				currentSectionName = namedsecmatch[2]

			} else {
				// Normal section
				currentSection = secmatch[1]
				currentSectionName = ""
			}
			continue
		}

		// Otherwise, probably a standard setting
		equalPos := strings.Index(line, "=")
		if equalPos != -1 {
			name := strings.TrimSpace(line[0:equalPos])
			value := strings.TrimSpace(line[equalPos+1:])
			if currentSection != "" {
				if currentSectionName != "" {
					name = fmt.Sprintf("%v.%v.%v", currentSection, currentSectionName, name)
				} else {
					name = fmt.Sprintf("%v.%v", currentSection, name)
				}
			}
			// convert key to lower case for easier matching
			name = strings.ToLower(name)

			// Check for includes and expand immediately
			if name == "include.path" {
				// if this is a relative, prepend containing dir context
				includeFile := value
				if !path.IsAbs(includeFile) {
					includeFile = path.Join(dir, includeFile)
				}
				includemap, err := ReadConfigFile(includeFile)
				if err == nil {
					for key, value := range includemap {
						ret[key] = value
					}
				}
			} else {
				ret[name] = value
			}
		}

	}
	if scanner.Err() != nil {
		// Problem (other than io.EOF)
		// return content we read up to here anyway
		return ret, scanner.Err()
	}

	return ret, nil

}
