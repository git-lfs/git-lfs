package lfsapi

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/git-lfs/git-lfs/config"
	"github.com/rubyist/tracerx"
)

// isCertVerificationDisabledForHost returns whether SSL certificate verification
// has been disabled for the given host, or globally
func isCertVerificationDisabledForHost(c *Client, host string) bool {
	hostSslVerify, _ := c.uc.Get("http", fmt.Sprintf("https://%v", host), "sslverify")
	if hostSslVerify == "false" {
		return true
	}

	return c.SkipSSLVerify
}

// isClientCertEnabledForHost returns whether client certificate
// are configured for the given host
func isClientCertEnabledForHost(c *Client, host string) bool {
	_, hostSslKeyOk := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslKey")
	_, hostSslCertOk := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslCert")

	return hostSslKeyOk && hostSslCertOk
}

// getClientCertForHost returns a client certificate for a specific host (which may
// be "host:port" loaded from the gitconfig
func getClientCertForHost(c *Client, host string) tls.Certificate {
	hostSslKey, _ := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslKey")
	hostSslCert, _ := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslCert")
	cert, err := tls.LoadX509KeyPair(hostSslCert, hostSslKey)
	if err != nil {
		tracerx.Printf("Error reading client cert/key %v", err)
	}
	return cert
}

// getRootCAsForHost returns a certificate pool for that specific host (which may
// be "host:port" loaded from either the gitconfig or from a platform-specific
// source which is not included by default in the golang certificate search)
// May return nil if it doesn't have anything to add, in which case the default
// RootCAs will be used if passed to TLSClientConfig.RootCAs
func getRootCAsForHost(c *Client, host string) *x509.CertPool {
	// don't init pool, want to return nil not empty if none found; init only on successful add cert
	var pool *x509.CertPool

	// gitconfig first
	pool = appendRootCAsForHostFromGitconfig(c.osEnv, c.gitEnv, pool, host)
	// Platform specific
	return appendRootCAsForHostFromPlatform(pool, host)
}

func appendRootCAsForHostFromGitconfig(osEnv, gitEnv config.Environment, pool *x509.CertPool, host string) *x509.CertPool {
	// Accumulate certs from all these locations:

	// GIT_SSL_CAINFO first
	if cafile, _ := osEnv.Get("GIT_SSL_CAINFO"); len(cafile) > 0 {
		return appendCertsFromFile(pool, cafile)
	}
	// http.<url>/.sslcainfo or http.<url>.sslcainfo
	uc := config.NewURLConfig(gitEnv)
	if cafile, ok := uc.Get("http", fmt.Sprintf("https://%v/", host), "sslcainfo"); ok {
		return appendCertsFromFile(pool, cafile)
	}
	// GIT_SSL_CAPATH
	if cadir, _ := osEnv.Get("GIT_SSL_CAPATH"); len(cadir) > 0 {
		return appendCertsFromFilesInDir(pool, cadir)
	}
	// http.sslcapath
	if cadir, ok := gitEnv.Get("http.sslcapath"); ok {
		return appendCertsFromFilesInDir(pool, cadir)
	}

	return pool
}

func appendCertsFromFilesInDir(pool *x509.CertPool, dir string) *x509.CertPool {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		tracerx.Printf("Error reading cert dir %q: %v", dir, err)
		return pool
	}
	for _, f := range files {
		pool = appendCertsFromFile(pool, filepath.Join(dir, f.Name()))
	}
	return pool
}

func appendCertsFromFile(pool *x509.CertPool, filename string) *x509.CertPool {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		tracerx.Printf("Error reading cert file %q: %v", filename, err)
		return pool
	}
	// Firstly, try parsing as binary certificate
	if certs, err := x509.ParseCertificates(data); err == nil {
		return appendCerts(pool, certs)
	}
	// If not binary certs, try PEM data
	return appendCertsFromPEMData(pool, data)
}

func appendCerts(pool *x509.CertPool, certs []*x509.Certificate) *x509.CertPool {
	if len(certs) == 0 {
		// important to return unmodified (may be nil)
		return pool
	}

	if pool == nil {
		pool = x509.NewCertPool()
	}

	for _, cert := range certs {
		pool.AddCert(cert)
	}

	return pool
}

func appendCertsFromPEMData(pool *x509.CertPool, data []byte) *x509.CertPool {
	if len(data) == 0 {
		return pool
	}

	// Bit of a dance, need to ensure if AppendCertsFromPEM fails we still return
	// nil and not an empty pool, so system roots still get used
	var ret *x509.CertPool
	if pool == nil {
		ret = x509.NewCertPool()
	} else {
		ret = pool
	}
	if !ret.AppendCertsFromPEM(data) {
		// Return unmodified input pool (may be nil, do not replace with empty)
		return pool
	}
	return ret
}
