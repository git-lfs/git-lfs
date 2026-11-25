package lfshttp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/v3/creds"
	"github.com/stretchr/testify/assert"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

var testCert = `-----BEGIN CERTIFICATE-----
MIIDyjCCArKgAwIBAgIJAMi9TouXnW+ZMA0GCSqGSIb3DQEBBQUAMEwxCzAJBgNV
BAYTAlVTMRMwEQYDVQQIEwpTb21lLVN0YXRlMRAwDgYDVQQKEwdnaXQtbGZzMRYw
FAYDVQQDEw1naXQtbGZzLmxvY2FsMB4XDTE2MDMwOTEwNTk1NFoXDTI2MDMwNzEw
NTk1NFowTDELMAkGA1UEBhMCVVMxEzARBgNVBAgTClNvbWUtU3RhdGUxEDAOBgNV
BAoTB2dpdC1sZnMxFjAUBgNVBAMTDWdpdC1sZnMubG9jYWwwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQCXmsI2w44nOsP7n3kL1Lz04U5FMZRErBSXLOE+
dpd4tMpgrjOncJPD9NapHabsVIOnuVvMDuBbWYwU9PwbN4tjQzch8DRxBju6fCp/
Pm+QF6p2Ga+NuSHWoVfNFuF2776aF9gSLC0rFnBekD3HCz+h6I5HFgHBvRjeVyAs
PRw471Y28Je609SoYugxaQNzRvahP0Qf43tE74/WN3FTGXy1+iU+uXpfp8KxnsuB
gfj+Wi6mPt8Q2utcA1j82dJ0K8ZbHSbllzmI+N/UuRLsbTUEdeFWYdZ0AlZNd/Vc
PlOSeoExwvOHIuUasT/cLIrEkdXNud2QLg2GpsB6fJi3NEUhAgMBAAGjga4wgasw
HQYDVR0OBBYEFC8oVPRQbekTwfkntgdL7PADXNDbMHwGA1UdIwR1MHOAFC8oVPRQ
bekTwfkntgdL7PADXNDboVCkTjBMMQswCQYDVQQGEwJVUzETMBEGA1UECBMKU29t
ZS1TdGF0ZTEQMA4GA1UEChMHZ2l0LWxmczEWMBQGA1UEAxMNZ2l0LWxmcy5sb2Nh
bIIJAMi9TouXnW+ZMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEFBQADggEBACIl
/CBLIhC3drrYme4cGArhWyXIyRpMoy9Z+9Dru8rSuOr/RXR6sbYhlE1iMGg4GsP8
4Cj7aIct6Vb9NFv5bGNyFJAmDesm3SZlEcWxU3YBzNPiJXGiUpQHCkp0BH+gvsXc
tb58XoiDZPVqrl0jNfX/nHpHR9c3DaI3Tjx0F/No0ZM6mLQ1cNMikFyEWQ4U0zmW
LvV+vvKuOixRqbcVnB5iTxqMwFG0X3tUql0cftGBgoCoR1+FSBOs0EXLODCck6ql
aW6vZwkA+ccj/pDTx8LBe2lnpatrFeIt6znAUJW3G8r6SFHKVBWHwmESZS4kxhjx
NpW5Hh0w4/5iIetCkJ0=
-----END CERTIFICATE-----`

var sslCAInfoConfigHostNames = []string{
	"git-lfs.local",
	"git-lfs.local/",
}
var sslCAInfoMatchedHostTests = []struct {
	hostName    string
	shouldMatch bool
}{
	{"git-lfs.local", true},
	{"git-lfs.local:8443", false},
	{"wronghost.com", false},
}

func clientForHost(c *Client, host string) *http.Client {
	u, _ := url.Parse(fmt.Sprintf("https://%v", host))
	client, _ := c.HttpClient(u, creds.BasicAccess)
	return client
}

func TestCertFromSSLCAInfoConfig(t *testing.T) {
	tempfile, err := os.CreateTemp("", "testcert")
	assert.Nil(t, err, "Error creating temp cert file")
	defer os.Remove(tempfile.Name())

	_, err = tempfile.WriteString(testCert)
	assert.Nil(t, err, "Error writing temp cert file")
	tempfile.Close()

	// Test http.<url>.sslcainfo
	for _, hostName := range sslCAInfoConfigHostNames {
		hostKey := fmt.Sprintf("http.https://%v.sslcainfo", hostName)
		c := NewClient(NewContext(nil, nil, map[string]string{
			hostKey: tempfile.Name(),
		}))

		for _, matchedHostTest := range sslCAInfoMatchedHostTests {
			pool := getRootCAsForHostFromGitconfig(c, matchedHostTest.hostName)

			var shouldOrShouldnt string
			if matchedHostTest.shouldMatch {
				shouldOrShouldnt = "should"
			} else {
				shouldOrShouldnt = "should not"
			}

			assert.Equal(t, matchedHostTest.shouldMatch, pool != nil,
				"Cert lookup for \"%v\" %v have succeeded with \"%v\"",
				matchedHostTest.hostName, shouldOrShouldnt, hostKey)
		}
	}

	// Test http.sslcainfo
	c := NewClient(NewContext(nil, nil, map[string]string{
		"http.sslcainfo": tempfile.Name(),
	}))

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHostFromGitconfig(c, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}
}

func TestCertFromSSLCAInfoEnv(t *testing.T) {
	tempfile, err := os.CreateTemp("", "testcert")
	assert.Nil(t, err, "Error creating temp cert file")
	defer os.Remove(tempfile.Name())

	_, err = tempfile.WriteString(testCert)
	assert.Nil(t, err, "Error writing temp cert file")
	tempfile.Close()

	c := NewClient(NewContext(nil, map[string]string{
		"GIT_SSL_CAINFO": tempfile.Name(),
	}, nil))

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHostFromGitconfig(c, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}
}

func TestCertFromSSLCAInfoEnvIsIgnoredForSchannelBackend(t *testing.T) {
	tempfile, err := os.CreateTemp("", "testcert")
	assert.Nil(t, err, "Error creating temp cert file")
	defer os.Remove(tempfile.Name())

	_, err = tempfile.WriteString(testCert)
	assert.Nil(t, err, "Error writing temp cert file")
	tempfile.Close()

	c := NewClient(NewContext(nil, map[string]string{
		"GIT_SSL_CAINFO": tempfile.Name(),
	}, map[string]string{
		"http.sslbackend": "schannel",
	}))

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHostFromGitconfig(c, matchedHostTest.hostName)
		assert.Nil(t, pool)
	}
}

func TestCertFromSSLCAInfoEnvWithSchannelBackend(t *testing.T) {
	tempfile, err := os.CreateTemp("", "testcert")
	assert.Nil(t, err, "Error creating temp cert file")
	defer os.Remove(tempfile.Name())

	_, err = tempfile.WriteString(testCert)
	assert.Nil(t, err, "Error writing temp cert file")
	tempfile.Close()

	c := NewClient(NewContext(nil, map[string]string{
		"GIT_SSL_CAINFO": tempfile.Name(),
	}, map[string]string{
		"http.sslbackend":           "schannel",
		"http.schannelusesslcainfo": "1",
	}))

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHostFromGitconfig(c, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}
}

func TestCertFromSSLCAPathConfig(t *testing.T) {
	tempdir := t.TempDir()

	err := os.WriteFile(filepath.Join(tempdir, "cert1.pem"), []byte(testCert), 0644)
	assert.Nil(t, err, "Error creating cert file")

	c := NewClient(NewContext(nil, nil, map[string]string{
		"http.sslcapath": tempdir,
	}))

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHostFromGitconfig(c, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}
}

func TestCertFromSSLCAPathEnv(t *testing.T) {
	tempdir := t.TempDir()

	err := os.WriteFile(filepath.Join(tempdir, "cert1.pem"), []byte(testCert), 0644)
	assert.Nil(t, err, "Error creating cert file")

	c := NewClient(NewContext(nil, map[string]string{
		"GIT_SSL_CAPATH": tempdir,
	}, nil))

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHostFromGitconfig(c, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}
}

func TestCertVerifyDisabledGlobalEnv(t *testing.T) {
	empty := NewClient(nil)
	httpClient := clientForHost(empty, "anyhost.com")
	tr, ok := httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	c := NewClient(NewContext(nil, map[string]string{
		"GIT_SSL_NO_VERIFY": "1",
	}, nil))

	httpClient = clientForHost(c, "anyhost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestCertVerifyDisabledGlobalConfig(t *testing.T) {
	def := NewClient(nil)
	httpClient := clientForHost(def, "anyhost.com")
	tr, ok := httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	c := NewClient(NewContext(nil, nil, map[string]string{
		"http.sslverify": "false",
	}))

	httpClient = clientForHost(c, "anyhost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestCertVerifyDisabledHostConfig(t *testing.T) {
	def := NewClient(nil)
	httpClient := clientForHost(def, "specifichost.com")
	tr, ok := httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	httpClient = clientForHost(def, "otherhost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	c := NewClient(NewContext(nil, nil, map[string]string{
		"http.https://specifichost.com/.sslverify": "false",
	}))

	httpClient = clientForHost(c, "specifichost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	httpClient = clientForHost(c, "otherhost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestParseP12(t *testing.T) {
	rootKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	rootTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "root"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTmpl, rootTmpl, &rootKey.PublicKey, rootKey)
	assert.Nil(t, err)
	root, _ := x509.ParseCertificate(rootDER)

	interKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	interTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "inter"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	interDER, err := x509.CreateCertificate(rand.Reader, interTmpl, root, &interKey.PublicKey, rootKey)
	assert.Nil(t, err)
	inter, _ := x509.ParseCertificate(interDER)

	leafKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(3), Subject: pkix.Name{CommonName: "client"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, inter, &leafKey.PublicKey, interKey)
	assert.Nil(t, err)
	leaf, _ := x509.ParseCertificate(leafDER)

	// Root and intermediate listed before the leaf's usual position.
	pfx, err := pkcs12.Modern.Encode(leafKey, leaf, []*x509.Certificate{root, inter}, "secret")
	assert.Nil(t, err)

	cert, err := parseP12(pfx, "secret")
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{leafDER, interDER}, cert.Certificate, "chain is leaf + intermediate, no root")

	_, err = parseP12(pfx, "wrong")
	assert.Equal(t, pkcs12.ErrIncorrectPassword, err)
	_, err = parseP12(pfx, "")
	assert.Equal(t, pkcs12.ErrIncorrectPassword, err)
}

func TestIsClientCertEnabledForP12(t *testing.T) {
	p12 := NewClient(NewContext(nil, nil, map[string]string{"http.https://h/.sslcert": "/x/a.p12"}))
	assert.True(t, isClientCertEnabledForHost(p12, "h"))
	typed := NewClient(NewContext(nil, nil, map[string]string{
		"http.https://h/.sslcert": "/x/a.bin", "http.https://h/.sslcerttype": "P12"}))
	assert.True(t, isClientCertEnabledForHost(typed, "h"))
	pem := NewClient(NewContext(nil, nil, map[string]string{"http.https://h/.sslcert": "/x/a.pem"}))
	assert.False(t, isClientCertEnabledForHost(pem, "h"))
}
