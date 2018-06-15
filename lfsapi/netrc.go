package lfsapi

import (
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/go-netrc/netrc"
)

type NetrcFinder interface {
	FindMachine(string) *netrc.Machine
}

func ParseNetrc(osEnv config.Environment) (NetrcFinder, string, error) {
	home, _ := osEnv.Get("HOME")
	if len(home) == 0 {
		return &noFinder{}, "", nil
	}

	nrcfilename := filepath.Join(home, netrcBasename)
	if _, err := os.Stat(nrcfilename); err != nil {
		return &noFinder{}, nrcfilename, nil
	}

	f, err := netrc.ParseFile(nrcfilename)
	return f, nrcfilename, err
}

type noFinder struct{}

func (f *noFinder) FindMachine(host string) *netrc.Machine {
	return nil
}
