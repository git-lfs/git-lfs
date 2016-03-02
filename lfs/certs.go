package lfs

import (
	"crypto/x509"
)

var (
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
	addUserPlatformCerts()
	addGitConfigCerts()
}
