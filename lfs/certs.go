package lfs

import (
	"crypto/x509"
)

var (
	// pool must include ALL certs to be trusted, including system root certs
	trustedCerts *x509.CertPool
)

// addGitConfigCerts adds any SSL certs configured in gitconfig
func addGitConfigCerts() {

}

func addCertsFromPEMData(data []byte) bool {
	if trustedCerts == nil {
		trustedCerts = x509.NewCertPool()
	}
	return trustedCerts.AppendCertsFromPEM(data)
}

func init() {
	// Platform-specific implementations of this method must load all
	// certs to be trusted into trustedCerts, including replicating Go's system
	// root loading. Go only supports passing a definitive list of certs or nil
	// to tls.Config, and does not expose the systemRootsPool() it uses as
	// the default if its nil, so we have to do everything.
	addTrustedPlatformCerts()
	addGitConfigCerts()
}
