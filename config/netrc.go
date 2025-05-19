package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/git-lfs/go-netrc/netrc"
)

type netrcfinder interface {
	FindMachine(string, string) *netrc.Machine
}

type noNetrc struct{}

func (n *noNetrc) FindMachine(host string, loginName string) *netrc.Machine {
	return nil
}

func (c *Configuration) parseNetrc() (netrcfinder, error) {
	home, _ := c.Os.Get("HOME")
	if len(home) == 0 {
		return &noNetrc{}, nil
	}

	nrcfilename := filepath.Join(home, netrcBasename)
	if _, err := os.Stat(nrcfilename); err != nil {
		// If on Windows, also try _netrc instead
		if runtime.GOOS == "windows" {
			altFilename := filepath.Join(home, "_netrc")
			if _, errAlt := os.Stat(altFilename); errAlt == nil {
				nrcfilename = altFilename
			} else {
				return &noNetrc{}, nil
			}
		} else {
			return &noNetrc{}, nil
		}
	}

	return netrc.ParseFile(nrcfilename)
}
