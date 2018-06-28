package lfsapi

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/git-lfs/go-netrc/netrc"
)

func TestNetrcWithHostAndPort(t *testing.T) {
	netrcFinder := &fakeNetrc{}
	u, err := url.Parse("http://netrc-host:123/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL:    u,
		Header: http.Header{},
	}

	if !setAuthFromNetrc(netrcFinder, req) {
		t.Fatal("no netrc match")
	}

	auth := req.Header.Get("Authorization")
	if auth != "Basic YWJjOmRlZg==" {
		t.Fatalf("bad basic auth: %q", auth)
	}
}

func TestNetrcWithHost(t *testing.T) {
	netrcFinder := &fakeNetrc{}
	u, err := url.Parse("http://netrc-host/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL:    u,
		Header: http.Header{},
	}

	if !setAuthFromNetrc(netrcFinder, req) {
		t.Fatalf("no netrc match")
	}

	auth := req.Header.Get("Authorization")
	if auth != "Basic YWJjOmRlZg==" {
		t.Fatalf("bad basic auth: %q", auth)
	}
}

func TestNetrcWithBadHost(t *testing.T) {
	netrcFinder := &fakeNetrc{}
	u, err := url.Parse("http://other-host/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL:    u,
		Header: http.Header{},
	}

	if setAuthFromNetrc(netrcFinder, req) {
		t.Fatalf("unexpected netrc match")
	}

	auth := req.Header.Get("Authorization")
	if auth != "" {
		t.Fatalf("bad basic auth: %q", auth)
	}
}

type fakeNetrc struct{}

func (n *fakeNetrc) FindMachine(host string) *netrc.Machine {
	if strings.Contains(host, "netrc") {
		return &netrc.Machine{Login: "abc", Password: "def"}
	}
	return nil
}
