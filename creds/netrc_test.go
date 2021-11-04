package creds

import (
	"strings"
	"testing"

	"github.com/git-lfs/go-netrc/netrc"
)

func TestNetrcWithHostAndPort(t *testing.T) {
	var netrcHelper netrcCredentialHelper
	netrcHelper.netrcFinder = &fakeNetrc{}

	what := make(Creds)
	what["protocol"] = "http"
	what["host"] = "netrc-host:123"
	what["path"] = "/foo/bar"

	creds, err := netrcHelper.Fill(what)
	if err != nil {
		t.Fatalf("error retrieving netrc credentials: %s", err)
	}

	username := creds["username"]
	if username != "abc" {
		t.Fatalf("bad username: %s", username)
	}

	password := creds["password"]
	if password != "def" {
		t.Fatalf("bad password: %s", password)
	}
}

func TestNetrcWithHost(t *testing.T) {
	var netrcHelper netrcCredentialHelper
	netrcHelper.netrcFinder = &fakeNetrc{}

	what := make(Creds)
	what["protocol"] = "http"
	what["host"] = "netrc-host"
	what["path"] = "/foo/bar"

	creds, err := netrcHelper.Fill(what)
	if err != nil {
		t.Fatalf("error retrieving netrc credentials: %s", err)
	}

	username := creds["username"]
	if username != "abc" {
		t.Fatalf("bad username: %s", username)
	}

	password := creds["password"]
	if password != "def" {
		t.Fatalf("bad password: %s", password)
	}
}

func TestNetrcWithBadHost(t *testing.T) {
	var netrcHelper netrcCredentialHelper
	netrcHelper.netrcFinder = &fakeNetrc{}

	what := make(Creds)
	what["protocol"] = "http"
	what["host"] = "other-host"
	what["path"] = "/foo/bar"

	_, err := netrcHelper.Fill(what)
	if err != credHelperNoOp {
		t.Fatalf("expected no-op for unknown host other-host")
	}
}

type fakeNetrc struct{}

func (n *fakeNetrc) FindMachine(host string, loginName string) *netrc.Machine {
	if strings.Contains(host, "netrc") {
		return &netrc.Machine{Login: "abc", Password: "def"}
	}
	return nil
}
