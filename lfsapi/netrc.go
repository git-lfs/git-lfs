package lfsapi

import (
	"os"
	"path/filepath"

	"github.com/bgentry/go-netrc/netrc"
)

type NetrcFinder interface {
	FindMachine(string) *netrc.Machine
}

func ParseNetrc(osEnv Env) (NetrcFinder, error) {
	home, _ := osEnv.Get("HOME")
	if len(home) == 0 {
		return &noFinder{}, nil
	}

	nrcfilename := filepath.Join(home, netrcBasename)
	if _, err := os.Stat(nrcfilename); err != nil {
		return &noFinder{}, nil
	}

	return netrc.ParseFile(nrcfilename)
}

type noFinder struct{}

func (f *noFinder) FindMachine(host string) *netrc.Machine {
	return nil
}
