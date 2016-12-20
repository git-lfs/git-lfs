package lfsapi

import (
	"crypto/x509"
	"regexp"
	"strings"

	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/rubyist/tracerx"
)

func appendRootCAsForHostFromPlatform(pool *x509.CertPool, host string) *x509.CertPool {
	// Go loads only the system root certificates by default
	// see https://github.com/golang/go/blob/master/src/crypto/x509/root_darwin.go
	// We want to load certs configured in the System keychain too, this is separate
	// from the system root certificates. It's also where other tools such as
	// browsers (e.g. Chrome) will load custom trusted certs from. They often
	// don't load certs from the login keychain so that's not included here
	// either, for consistency.

	// find system.keychain for user-added certs (don't assume location)
	cmd := subprocess.ExecCommand("/usr/bin/security", "list-keychains")
	kcout, err := cmd.Output()
	if err != nil {
		tracerx.Printf("Error listing keychains: %v", err)
		return nil
	}

	var systemKeychain string
	keychains := strings.Split(string(kcout), "\n")
	for _, keychain := range keychains {
		lc := strings.ToLower(keychain)
		if !strings.Contains(lc, "/system.keychain") {
			continue
		}
		systemKeychain = strings.Trim(keychain, " \t\"")
		break
	}

	if len(systemKeychain) == 0 {
		return nil
	}

	pool = appendRootCAsFromKeychain(pool, host, systemKeychain)

	// Also check host without port
	portreg := regexp.MustCompile(`([^:]+):\d+`)
	if match := portreg.FindStringSubmatch(host); match != nil {
		hostwithoutport := match[1]
		pool = appendRootCAsFromKeychain(pool, hostwithoutport, systemKeychain)
	}

	return pool
}

func appendRootCAsFromKeychain(pool *x509.CertPool, name, keychain string) *x509.CertPool {
	cmd := subprocess.ExecCommand("/usr/bin/security", "find-certificate", "-a", "-p", "-c", name, keychain)
	data, err := cmd.Output()
	if err != nil {
		tracerx.Printf("Error reading keychain %q: %v", keychain, err)
		return pool
	}
	return appendCertsFromPEMData(pool, data)
}
