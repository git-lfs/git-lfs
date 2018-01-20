package lfsapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestCertFromSSLCAInfoConfig(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "testcert")
	assert.Nil(t, err, "Error creating temp cert file")
	defer os.Remove(tempfile.Name())

	_, err = tempfile.WriteString(testCert)
	assert.Nil(t, err, "Error writing temp cert file")
	tempfile.Close()

	// Test http.<url>.sslcainfo
	for _, hostName := range sslCAInfoConfigHostNames {
		hostKey := fmt.Sprintf("http.https://%v.sslcainfo", hostName)
		c, err := NewClient(NewContext(nil, nil, map[string]string{
			hostKey: tempfile.Name(),
		}))
		assert.Nil(t, err)

		for _, matchedHostTest := range sslCAInfoMatchedHostTests {
			pool := getRootCAsForHost(c, matchedHostTest.hostName)

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
	c, err := NewClient(NewContext(nil, nil, map[string]string{
		"http.sslcainfo": tempfile.Name(),
	}))
	assert.Nil(t, err)

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHost(c, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}
}

func TestCertFromSSLCAInfoEnv(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "testcert")
	assert.Nil(t, err, "Error creating temp cert file")
	defer os.Remove(tempfile.Name())

	_, err = tempfile.WriteString(testCert)
	assert.Nil(t, err, "Error writing temp cert file")
	tempfile.Close()

	c, err := NewClient(NewContext(nil, map[string]string{
		"GIT_SSL_CAINFO": tempfile.Name(),
	}, nil))
	assert.Nil(t, err)

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHost(c, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}
}

func TestCertFromSSLCAPathConfig(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "testcertdir")
	assert.Nil(t, err, "Error creating temp cert dir")
	defer os.RemoveAll(tempdir)

	err = ioutil.WriteFile(filepath.Join(tempdir, "cert1.pem"), []byte(testCert), 0644)
	assert.Nil(t, err, "Error creating cert file")

	c, err := NewClient(NewContext(nil, nil, map[string]string{
		"http.sslcapath": tempdir,
	}))

	assert.Nil(t, err)

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHost(c, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}
}

func TestCertFromSSLCAPathEnv(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "testcertdir")
	assert.Nil(t, err, "Error creating temp cert dir")
	defer os.RemoveAll(tempdir)

	err = ioutil.WriteFile(filepath.Join(tempdir, "cert1.pem"), []byte(testCert), 0644)
	assert.Nil(t, err, "Error creating cert file")

	c, err := NewClient(NewContext(nil, map[string]string{
		"GIT_SSL_CAPATH": tempdir,
	}, nil))
	assert.Nil(t, err)

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHost(c, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}
}

func TestCertVerifyDisabledGlobalEnv(t *testing.T) {
	empty, _ := NewClient(nil)
	httpClient := empty.httpClient("anyhost.com")
	tr, ok := httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	c, err := NewClient(NewContext(nil, map[string]string{
		"GIT_SSL_NO_VERIFY": "1",
	}, nil))

	assert.Nil(t, err)

	httpClient = c.httpClient("anyhost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestCertVerifyDisabledGlobalConfig(t *testing.T) {
	def, _ := NewClient(nil)
	httpClient := def.httpClient("anyhost.com")
	tr, ok := httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	c, err := NewClient(NewContext(nil, nil, map[string]string{
		"http.sslverify": "false",
	}))
	assert.Nil(t, err)

	httpClient = c.httpClient("anyhost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestCertVerifyDisabledHostConfig(t *testing.T) {
	def, _ := NewClient(nil)
	httpClient := def.httpClient("specifichost.com")
	tr, ok := httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	httpClient = def.httpClient("otherhost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	c, err := NewClient(NewContext(nil, nil, map[string]string{
		"http.https://specifichost.com/.sslverify": "false",
	}))
	assert.Nil(t, err)

	httpClient = c.httpClient("specifichost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
	}

	httpClient = c.httpClient("otherhost.com")
	tr, ok = httpClient.Transport.(*http.Transport)
	if assert.True(t, ok) {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}
}
