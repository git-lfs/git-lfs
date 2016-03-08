package lfs

import "crypto/x509"

func getRootCAsForHostFromPlatform(host string) *x509.CertPool {
	// Do nothing, use golang default
	return nil
}
