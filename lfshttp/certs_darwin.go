package lfshttp

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"regexp"
	"strings"

	"github.com/git-lfs/git-lfs/v3/subprocess"
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
	cmd, err := subprocess.ExecCommand("/usr/bin/security", "list-keychains")
	if err != nil {
		tracerx.Printf("Error getting command to list keychains: %v", err)
		return nil
	}
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

func appendRootCAsFromKeychain(pool *x509.CertPool, hostname, keychain string) *x509.CertPool {
	cmd, err := subprocess.ExecCommand("/usr/bin/security", "find-certificate", "-a", "-p", "-c", hostname, keychain)
	if err != nil {
		tracerx.Printf("Error getting command to read keychain %q: %v", keychain, err)
		return pool
	}
	data, err := cmd.Output()
	if err != nil {
		tracerx.Printf("Error reading keychain %q: %v", keychain, err)
		return pool
	}

	// /usr/bin/security find-certificate does not support exact matching, so we have to do
	// some manual filtering.
	return appendMatchingCertsByCN(pool, data, hostname)
}

// appendRootCAsFromKeychain adds certificates from the specified keychain to the certificate pool
// for the given hostname. It filters out CA certificates that have the hostname as a substring
// in their CN but aren't an exact match for the hostname. This prevents issues where searching
// for "hostname" returns certificates like "hostname Some Authority" that shouldn't be used
// for TLS verification of the actual hostname.
func appendMatchingCertsByCN(pool *x509.CertPool, data []byte, hostname string) *x509.CertPool {
	if len(data) == 0 {
		return pool
	}

	// Parse PEM data into individual certificates
	pemBlocks := []byte(data)
	var validCerts [][]byte

	for len(pemBlocks) > 0 {
		var block *pem.Block
		block, pemBlocks = pem.Decode(pemBlocks)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		// Skip CA certificates where the hostname is a substring of the CN
		// but not an exact match (like "example.com JSS Built-in Certificate Authority")
		if cert.IsCA &&
			strings.Contains(cert.Subject.CommonName, hostname) &&
			!strings.EqualFold(cert.Subject.CommonName, hostname) {
			tracerx.Printf("Skipping CA certificate with problematic CN: %s", cert.Subject.CommonName)
			continue
		}

		// All other certificates are accepted
		validCerts = append(validCerts, pem.EncodeToMemory(block))

	}

	// Add only the filtered certificates
	if len(validCerts) > 0 {
		allCerts := bytes.Join(validCerts, []byte{})
		appendCertsFromPEMData(pool, allCerts)
	}

	return pool
}
