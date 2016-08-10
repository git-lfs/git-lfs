package httputil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/github/git-lfs/config"
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
		cfg := config.NewFrom(config.Values{
			Git: map[string]string{hostKey: tempfile.Name()},
		})

		for _, matchedHostTest := range sslCAInfoMatchedHostTests {
			pool := getRootCAsForHost(cfg, matchedHostTest.hostName)

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
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{"http.sslcainfo": tempfile.Name()},
	})

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHost(cfg, matchedHostTest.hostName)
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

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSL_CAINFO": tempfile.Name(),
		},
	})

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHost(cfg, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}

}

func TestCertFromSSLCAPathConfig(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "testcertdir")
	assert.Nil(t, err, "Error creating temp cert dir")
	defer os.RemoveAll(tempdir)

	err = ioutil.WriteFile(filepath.Join(tempdir, "cert1.pem"), []byte(testCert), 0644)
	assert.Nil(t, err, "Error creating cert file")

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{"http.sslcapath": tempdir},
	})

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHost(cfg, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}

}

func TestCertFromSSLCAPathEnv(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "testcertdir")
	assert.Nil(t, err, "Error creating temp cert dir")
	defer os.RemoveAll(tempdir)

	err = ioutil.WriteFile(filepath.Join(tempdir, "cert1.pem"), []byte(testCert), 0644)
	assert.Nil(t, err, "Error creating cert file")

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSL_CAPATH": tempdir,
		},
	})

	// Should match any host at all
	for _, matchedHostTest := range sslCAInfoMatchedHostTests {
		pool := getRootCAsForHost(cfg, matchedHostTest.hostName)
		assert.NotNil(t, pool)
	}

}

func TestCertVerifyDisabledGlobalEnv(t *testing.T) {
	empty := config.NewFrom(config.Values{})
	assert.False(t, isCertVerificationDisabledForHost(empty, "anyhost.com"))

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSL_NO_VERIFY": "1",
		},
	})
	assert.True(t, isCertVerificationDisabledForHost(cfg, "anyhost.com"))
}

func TestCertVerifyDisabledGlobalConfig(t *testing.T) {
	def := config.New()
	assert.False(t, isCertVerificationDisabledForHost(def, "anyhost.com"))

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{"http.sslverify": "false"},
	})
	assert.True(t, isCertVerificationDisabledForHost(cfg, "anyhost.com"))
}

func TestCertVerifyDisabledHostConfig(t *testing.T) {
	def := config.New()
	assert.False(t, isCertVerificationDisabledForHost(def, "specifichost.com"))
	assert.False(t, isCertVerificationDisabledForHost(def, "otherhost.com"))

	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"http.https://specifichost.com/.sslverify": "false",
		},
	})
	assert.True(t, isCertVerificationDisabledForHost(cfg, "specifichost.com"))
	assert.False(t, isCertVerificationDisabledForHost(cfg, "otherhost.com"))
}
