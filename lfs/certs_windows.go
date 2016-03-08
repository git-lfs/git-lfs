package lfs

import "crypto/x509"

func getRootCAsForHostFromPlatform(host string) *x509.CertPool {
	// golang already supports Windows Certificate Store for self-signed certs
	return nil
}
