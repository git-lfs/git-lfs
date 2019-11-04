package creds

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/go-netrc/netrc"
	"github.com/rubyist/tracerx"
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

// NetrcCredentialHelper retrieves credentials from a .netrc file
type netrcCredentialHelper struct {
	netrcFinder NetrcFinder
	mu          sync.Mutex
	skip        map[string]bool
}

var defaultNetrcFinder = &noFinder{}

// NewNetrcCredentialHelper creates a new netrc credential helper using a
// .netrc file gleaned from the OS environment
func newNetrcCredentialHelper(osEnv config.Environment) *netrcCredentialHelper {
	netrcFinder, netrcfile, err := ParseNetrc(osEnv)
	if err != nil {
		tracerx.Printf("bad netrc file %s: %s", netrcfile, err)
		return nil
	}

	if netrcFinder == nil {
		netrcFinder = defaultNetrcFinder
	}

	return &netrcCredentialHelper{netrcFinder: netrcFinder, skip: make(map[string]bool)}
}

func (c *netrcCredentialHelper) Fill(what Creds) (Creds, error) {
	host, err := getNetrcHostname(what["host"])
	if err != nil {
		return nil, credHelperNoOp
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.skip[host] {
		return nil, credHelperNoOp
	}
	if machine := c.netrcFinder.FindMachine(host); machine != nil {
		creds := make(Creds)
		creds["username"] = machine.Login
		creds["password"] = machine.Password
		creds["protocol"] = what["protocol"]
		creds["host"] = what["host"]
		creds["scheme"] = what["scheme"]
		creds["path"] = what["path"]
		creds["source"] = "netrc"
		tracerx.Printf("netrc: git credential fill (%q, %q, %q)",
			what["protocol"], what["host"], what["path"])
		return creds, nil
	}

	return nil, credHelperNoOp
}

func getNetrcHostname(hostname string) (string, error) {
	if strings.Contains(hostname, ":") {
		host, _, err := net.SplitHostPort(hostname)
		if err != nil {
			tracerx.Printf("netrc: error parsing %q: %s", hostname, err)
			return "", err
		}
		return host, nil
	}

	return hostname, nil
}

func (c *netrcCredentialHelper) Approve(what Creds) error {
	if what["source"] == "netrc" {
		host, err := getNetrcHostname(what["host"])
		if err != nil {
			return credHelperNoOp
		}
		tracerx.Printf("netrc: git credential approve (%q, %q, %q)",
			what["protocol"], what["host"], what["path"])
		c.mu.Lock()
		c.skip[host] = false
		c.mu.Unlock()
		return nil
	}
	return credHelperNoOp
}

func (c *netrcCredentialHelper) Reject(what Creds) error {
	if what["source"] == "netrc" {
		host, err := getNetrcHostname(what["host"])
		if err != nil {
			return credHelperNoOp
		}

		tracerx.Printf("netrc: git credential reject (%q, %q, %q)",
			what["protocol"], what["host"], what["path"])
		c.mu.Lock()
		c.skip[host] = true
		c.mu.Unlock()
		return nil
	}
	return credHelperNoOp
}
