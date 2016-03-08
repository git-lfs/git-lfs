package lfs

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

// getRootCAsForHost returns a certificate pool for that specific host (which may
// be "host:port" loaded from either the gitconfig or from a platform-specific
// source which is not included by default in the golang certificate search)
// May return nil if it doesn't have anything to add, in which case the default
// RootCAs will be used if passed to TLSClientConfig.RootCAs
func getRootCAsForHost(host string) *x509.CertPool {

	// don't init pool, want to return nil not empty if none found; init only on successful add cert
	var pool *x509.CertPool

	// gitconfig first
	pool = appendRootCAsForHostFromGitconfig(pool, host)
	// Platform specific
	return appendRootCAsForHostFromPlatform(pool, host)

}

func appendRootCAsForHostFromGitconfig(pool *x509.CertPool, host string) *x509.CertPool {
	// Accumulate certs from all these locations:

	loaded := false
	// GIT_SSL_CAINFO first
	if cafile, ok := Config.envVars["GIT_SSL_CAINFO"]; ok {
		pool = appendCertsFromFile(pool, cafile)
		loaded = true
	}
	// http.<url>.sslcainfo
	if !loaded {
		// match url-specific gitconfig settings
		// we know we have simply "host" or "host:port"
		key := fmt.Sprintf("http.https://%v/.sslcainfo", host)
		if cafile, ok := Config.GitConfig(key); ok {
			pool = appendCertsFromFile(pool, cafile)
			loaded = true
		}
	}
	// http.sslcainfo
	if !loaded {
		if cafile, ok := Config.GitConfig("http.sslcainfo"); ok {
			pool = appendCertsFromFile(pool, cafile)
			loaded = true
		}
	}
	// GIT_SSL_CAPATH
	if !loaded {
		if cadir, ok := Config.envVars["GIT_SSL_CAPATH"]; ok {
			pool = appendCertsFromFilesInDir(pool, cadir)
			loaded = true
		}
	}
	// http.sslcapath
	if !loaded {
		if cadir, ok := Config.GitConfig("http.sslcapath"); ok {
			pool = appendCertsFromFilesInDir(pool, cadir)
			loaded = true
		}
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
	return appendCertsFromPEMData(pool, data)
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
