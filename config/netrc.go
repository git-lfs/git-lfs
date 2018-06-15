package config

import (
	"os"
	"path/filepath"

	"github.com/git-lfs/go-netrc/netrc"
)

type netrcfinder interface {
	FindMachine(string) *netrc.Machine
}

type noNetrc struct{}

func (n *noNetrc) FindMachine(host string) *netrc.Machine {
	return nil
}

func (c *Configuration) parseNetrc() (netrcfinder, error) {
	home, _ := c.Os.Get("HOME")
	if len(home) == 0 {
		return &noNetrc{}, nil
	}

	nrcfilename := filepath.Join(home, netrcBasename)
	if _, err := os.Stat(nrcfilename); err != nil {
		return &noNetrc{}, nil
	}

	return netrc.ParseFile(nrcfilename)
}
