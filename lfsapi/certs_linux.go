package lfsapi

import "crypto/x509"

func appendRootCAsForHostFromPlatform(pool *x509.CertPool, host string) *x509.CertPool {
	// Do nothing, use golang default
	return pool
}
