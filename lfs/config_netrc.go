package lfs

import (
	"os"
	"path/filepath"

	"github.com/github/git-lfs/vendor/_nuts/github.com/bgentry/go-netrc/netrc"
)

// different on unix vs windows
var netrcBasename string

type netrcfinder interface {
	FindMachine(string) *netrc.Machine
}

type noNetrc struct{}

func (n *noNetrc) FindMachine(host string) *netrc.Machine {
	return nil
}

func (c *Configuration) parseNetrc() (netrcfinder, error) {
	home := c.Getenv("HOME")
	if len(home) == 0 {
		return &noNetrc{}, nil
	}

	nrcfilename := filepath.Join(home, netrcBasename)
	if _, err := os.Stat(nrcfilename); err != nil {
		return &noNetrc{}, nil
	}

	return netrc.ParseFile(nrcfilename)
}
