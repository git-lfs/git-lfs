package lfs

import (
	"strings"

	"github.com/github/git-lfs/subprocess"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

func addUserPlatformCerts() {
	// Go loads only the system root certificates by default
	// see https://github.com/golang/go/blob/master/src/crypto/x509/root_darwin.go
	// We want to load certs configured in the Login keychain and System keychain,
	// the 2 places people tend to add custom self-signed certs (former is per-user,
	// latter is system-wide)
	// To protect against these files moving, use security to list & match
	// For now, don't match all keychains to protect against something funky
	// both Adobe and Microsoft ship custom keychains on OS X

	// Unfortunately since tls.Config only allows the complete replacement of
	// all RootCAs, and golang doesn't expose the system certs it's already read,
	// we have to get the system root certs too (same technique)
	// This keychain is not included in 'security list-keychains'
	addCertsFromKeychain("/System/Library/Keychains/SystemRootCertificates.keychain")

	// Now find system.keychain and login.keychain for user-added certs
	cmd := subprocess.ExecCommand("/usr/bin/security", "list-keychains")
	kcout, err := cmd.Output()
	if err != nil {
		tracerx.Printf("Error listing keychains: %v", err)
		return
	}

	keychains := strings.Split(string(kcout), "\n")
	for _, keychain := range keychains {
		lc := strings.ToLower(keychain)
		if !strings.Contains(lc, "/login.keychain") && !strings.Contains(lc, "/system.keychain") {
			continue
		}
		keychain = strings.Trim(keychain, " \t\"")

		addCertsFromKeychain(keychain)
	}

}

func addCertsFromKeychain(keychain string) {
	// Extract all certs in the keychain in PEM format
	cmd := subprocess.ExecCommand("/usr/bin/security", "find-certificate", "-a", "-p", keychain)
	data, err := cmd.Output()
	if err != nil {
		tracerx.Printf("Error reading keychain %q: %v", keychain, err)
		return
	}
	addCertsFromPEMData(data)
}
