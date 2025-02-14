package config

import (
	"github.com/git-lfs/go-netrc/netrc"
)

type netrcfinder interface {
	FindMachine(string, string) *netrc.Machine
}

type noNetrc struct{}
