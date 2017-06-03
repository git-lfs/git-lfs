package lfsapi

import "crypto/x509"

func appendRootCAsForHostFromPlatform(pool *x509.CertPool, host string) *x509.CertPool {
	// golang already supports Windows Certificate Store for self-signed certs
	return pool
}
