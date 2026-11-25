package lfshttp

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/creds"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
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
	hostSslCert, hostSslCertOk := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslCert")

	// A PKCS#12 file holds the key too, so no separate sslKey is needed.
	return hostSslCertOk && (hostSslKeyOk || isP12ForHost(c, host, hostSslCert))
}

// isP12ForHost returns whether the client certificate configured for the
// given host is a PKCS#12 file, either by http.sslCertType or by extension.
func isP12ForHost(c *Client, host, certPath string) bool {
	if t, ok := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslCertType"); ok {
		return strings.EqualFold(t, "P12")
	}
	ext := strings.ToLower(filepath.Ext(certPath))
	return ext == ".p12" || ext == ".pfx"
}

// fillPassphrase prompts for the passphrase protecting the file at path, using
// the credential helper. The caller must call Approve or Reject on the result.
func fillPassphrase(c *Client, path string) (*creds.CredentialHelperWrapper, creds.Creds, string, error) {
	fileurl := fmt.Sprintf("cert:///%s", filepath.ToSlash(path))
	url, err := url.Parse(fileurl)
	if err != nil {
		return nil, nil, "", err
	}
	credWrapper := c.credHelperContext.GetCredentialHelper(nil, url)

	credWrapper.Input["username"] = []string{""}

	filled, err := credWrapper.CredentialHelper.Fill(credWrapper.Input)
	if err != nil {
		tracerx.Printf("Error filling credentials for %q: %v", fileurl, err)
		return nil, nil, "", err
	}
	return &credWrapper, filled, filled["password"][0], nil
}

// decryptPEMBlock decrypts an encrypted PEM block representing a private key,
// prompting for credentials using the credential helper, and returns a
// decrypted PEM block representing that same private key.
func decryptPEMBlock(c *Client, block *pem.Block, path string, key []byte) ([]byte, error) {
	credWrapper, filled, pass, err := fillPassphrase(c, path)
	if err != nil {
		return nil, err
	}
	decrypted, err := x509.DecryptPEMBlock(block, []byte(pass))
	if err != nil {
		credWrapper.CredentialHelper.Reject(filled)
		return nil, err
	}
	credWrapper.CredentialHelper.Approve(filled)

	// decrypted is a DER blob, but we need a PEM-encoded block.
	toEncode := &pem.Block{Type: block.Type, Headers: nil, Bytes: decrypted}
	buf := pem.EncodeToMemory(toEncode)
	return buf, nil
}

// loadP12 reads a PKCS#12 file holding a client key and certificate chain.
// An empty password is tried first; otherwise the credential helper is asked.
func loadP12(c *Client, path string) (*tls.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		tracerx.Printf("Error reading client cert file %q: %v", path, err)
		return nil, errors.Wrap(err, tr.Tr.Get("Error reading client cert file %q", path))
	}

	cert, err := parseP12(data, "")
	if err == pkcs12.ErrIncorrectPassword {
		credWrapper, filled, pass, ferr := fillPassphrase(c, path)
		if ferr != nil {
			return nil, ferr
		}
		cert, err = parseP12(data, pass)
		if err == pkcs12.ErrIncorrectPassword {
			credWrapper.CredentialHelper.Reject(filled)
		} else if err == nil {
			credWrapper.CredentialHelper.Approve(filled)
		}
	}
	if err != nil {
		tracerx.Printf("Unable to read client cert file %q: %v", path, err)
		return nil, errors.Wrap(err, tr.Tr.Get("Error reading client cert file %q", path))
	}
	return cert, nil
}

// parseP12 decodes a PKCS#12 bundle into a TLS certificate. The leaf is the
// certificate matching the private key; the chain follows the issuers up to,
// but not including, a self-signed root, which the server does not need.
func parseP12(data []byte, password string) (*tls.Certificate, error) {
	key, first, others, err := pkcs12.DecodeChain(data, password)
	if err != nil {
		return nil, err
	}
	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, errors.New(tr.Tr.Get("unsupported private key type %T", key))
	}

	// Don't assume the leaf comes first: pick the cert matching the key.
	all := append([]*x509.Certificate{first}, others...)
	var leaf *x509.Certificate
	for _, cand := range all {
		if pub, ok := cand.PublicKey.(interface{ Equal(crypto.PublicKey) bool }); ok && pub.Equal(signer.Public()) {
			leaf = cand
			break
		}
	}
	if leaf == nil {
		return nil, errors.New(tr.Tr.Get("no certificate matches the private key"))
	}

	chain := [][]byte{leaf.Raw}
	cur := leaf
	for range all {
		var next *x509.Certificate
		for _, cand := range all {
			if bytes.Equal(cand.RawSubject, cur.RawIssuer) && !bytes.Equal(cand.Raw, cur.Raw) {
				next = cand
				break
			}
		}
		if next == nil || bytes.Equal(next.RawSubject, next.RawIssuer) {
			break // no issuer bundled, or reached the self-signed root
		}
		chain = append(chain, next.Raw)
		cur = next
	}

	return &tls.Certificate{Certificate: chain, PrivateKey: signer, Leaf: leaf}, nil
}

// getClientCertForHost returns a client certificate for a specific host (which may
// be "host:port" loaded from the gitconfig
func getClientCertForHost(c *Client, host string) (*tls.Certificate, error) {
	hostSslKey, hostSslKeyOk := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslKey")
	hostSslCert, _ := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslCert")

	hostSslCert, err := tools.ExpandPath(hostSslCert, false)
	if err != nil {
		return nil, errors.Wrap(err, tr.Tr.Get("Error resolving cert path %q", hostSslCert))
	}

	if !hostSslKeyOk && isP12ForHost(c, host, hostSslCert) {
		return loadP12(c, hostSslCert)
	}

	hostSslKey, err = tools.ExpandPath(hostSslKey, false)
	if err != nil {
		return nil, errors.Wrap(err, tr.Tr.Get("Error resolving key path %q", hostSslKey))
	}

	cert, err := os.ReadFile(hostSslCert)
	if err != nil {
		tracerx.Printf("Error reading client cert file %q: %v", hostSslCert, err)
		return nil, errors.Wrap(err, tr.Tr.Get("Error reading client cert file %q", hostSslCert))
	}
	key, err := os.ReadFile(hostSslKey)
	if err != nil {
		tracerx.Printf("Error reading client key file %q: %v", hostSslKey, err)
		return nil, errors.Wrap(err, tr.Tr.Get("Error reading client key file %q", hostSslKey))
	}

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New(tr.Tr.Get("Error decoding PEM block from %q", hostSslKey))
	}
	if x509.IsEncryptedPEMBlock(block) {
		key, err = decryptPEMBlock(c, block, hostSslKey, key)
		if err != nil {
			tracerx.Printf("Unable to decrypt client key file %q: %v", hostSslKey, err)
			return nil, errors.Wrap(err, tr.Tr.Get("Error reading client key file %q (not a PKCS#1 file?)", hostSslKey))
		}
	}

	certobj, err := tls.X509KeyPair(cert, key)
	if err != nil {
		tracerx.Printf("Error reading client cert/key %v", err)
		return nil, errors.Wrap(err, tr.Tr.Get("Error reading client cert/key"))
	}
	return &certobj, nil
}

// getRootCAsForHostFromGitconfig returns a certificate pool for that
// specific host (which may be "host:port" or just "host") loaded from the
// Git configuration and environment settings, if any are defined.
// May return nil if it doesn't have anything to add, in which case the default
// RootCAs will be used if passed to TLSClientConfig.RootCAs
func getRootCAsForHostFromGitconfig(c *Client, host string) *x509.CertPool {
	// don't init pool, want to return nil not empty if none found; init only on successful add cert
	var pool *x509.CertPool

	url := fmt.Sprintf("https://%v/", host)
	uc := config.NewURLConfig(c.gitEnv)

	backend, _ := uc.Get("http", url, "sslbackend")
	schannelUseSslCaInfoStrValue, _ := uc.Get("http", url, "schannelusesslcainfo")
	schannelUseSslCaInfo := config.Bool(schannelUseSslCaInfoStrValue, false)

	if backend == "schannel" && !schannelUseSslCaInfo {
		return pool
	}

	// Accumulate certs from all these locations:

	// GIT_SSL_CAINFO first
	if cafile, _ := c.osEnv.Get("GIT_SSL_CAINFO"); len(cafile) > 0 {
		return appendCertsFromFile(pool, cafile)
	}
	// http.<url>/.sslcainfo or http.<url>.sslcainfo
	if cafile, ok := uc.Get("http", url, "sslcainfo"); ok {
		return appendCertsFromFile(pool, cafile)
	}
	// GIT_SSL_CAPATH
	if cadir, _ := c.osEnv.Get("GIT_SSL_CAPATH"); len(cadir) > 0 {
		return appendCertsFromFilesInDir(pool, cadir)
	}
	// http.sslcapath
	if cadir, ok := c.gitEnv.Get("http.sslcapath"); ok {
		return appendCertsFromFilesInDir(pool, cadir)
	}

	return pool
}

func appendCertsFromFilesInDir(pool *x509.CertPool, dir string) *x509.CertPool {
	dirpath, errpath := tools.TranslateCygwinPath(dir)
	if errpath != nil {
		tracerx.Printf("Error reading cert dir %q: %v", dirpath, errpath)
	}
	files, err := os.ReadDir(dirpath)
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
	filenamepath, errfile := tools.TranslateCygwinPath(filename)
	if errfile != nil {
		tracerx.Printf("Error reading cert dir %q: %v", filenamepath, errfile)
	}
	data, err := os.ReadFile(filenamepath)
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
