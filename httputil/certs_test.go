package httputil

import (
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

func TestCertFromSSLCAInfoConfig(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "testcert")
	assert.Nil(t, err, "Error creating temp cert file")
	defer os.Remove(tempfile.Name())

	_, err = tempfile.WriteString(testCert)
	assert.Nil(t, err, "Error writing temp cert file")
	tempfile.Close()

	cfg := config.New()
	cfg.SetConfig("http.https://git-lfs.local/.sslcainfo", tempfile.Name())

	t.Log("Should match")
	pool := getRootCAsForHost(cfg, "git-lfs.local")
	assert.NotNil(t, pool)

	t.Log("Should not match")
	pool = getRootCAsForHost(cfg, "wronghost.com")
	assert.Nil(t, pool)

	t.Log("Ports have to match")
	pool = getRootCAsForHost(cfg, "git-lfs.local:8443")
	assert.Nil(t, pool)

	cfg.SetConfig("http.sslcainfo", tempfile.Name())

	t.Log("Should match anything")
	pool = getRootCAsForHost(cfg, "git-lfs.local")
	assert.NotNil(t, pool)
	pool = getRootCAsForHost(cfg, "wronghost.com")
	assert.NotNil(t, pool)
	pool = getRootCAsForHost(cfg, "git-lfs.local:8443")
	assert.NotNil(t, pool)

}

func TestCertFromSSLCAInfoEnv(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "testcert")
	assert.Nil(t, err, "Error creating temp cert file")
	defer os.Remove(tempfile.Name())

	_, err = tempfile.WriteString(testCert)
	assert.Nil(t, err, "Error writing temp cert file")
	tempfile.Close()

	cfg := config.New()
	cfg.SetAllEnv(map[string]string{"GIT_SSL_CAINFO": tempfile.Name()})

	// Should match any host at all
	pool := getRootCAsForHost(cfg, "git-lfs.local")
	assert.NotNil(t, pool)
	pool = getRootCAsForHost(cfg, "wronghost.com")
	assert.NotNil(t, pool)
	pool = getRootCAsForHost(cfg, "notthisone.com:8888")
	assert.NotNil(t, pool)

}

func TestCertFromSSLCAPathConfig(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "testcertdir")
	assert.Nil(t, err, "Error creating temp cert dir")
	defer os.RemoveAll(tempdir)

	err = ioutil.WriteFile(filepath.Join(tempdir, "cert1.pem"), []byte(testCert), 0644)
	assert.Nil(t, err, "Error creating cert file")

	cfg := config.New()
	cfg.SetConfig("http.sslcapath", tempdir)

	// Should match any host at all
	pool := getRootCAsForHost(cfg, "git-lfs.local")
	assert.NotNil(t, pool)
	pool = getRootCAsForHost(cfg, "wronghost.com")
	assert.NotNil(t, pool)
	pool = getRootCAsForHost(cfg, "notthisone.com:8888")
	assert.NotNil(t, pool)

}

func TestCertFromSSLCAPathEnv(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "testcertdir")
	assert.Nil(t, err, "Error creating temp cert dir")
	defer os.RemoveAll(tempdir)

	err = ioutil.WriteFile(filepath.Join(tempdir, "cert1.pem"), []byte(testCert), 0644)
	assert.Nil(t, err, "Error creating cert file")

	cfg := config.New()
	cfg.SetAllEnv(map[string]string{"GIT_SSL_CAPATH": tempdir})

	// Should match any host at all
	pool := getRootCAsForHost(cfg, "git-lfs.local")
	assert.NotNil(t, pool)
	pool = getRootCAsForHost(cfg, "wronghost.com")
	assert.NotNil(t, pool)
	pool = getRootCAsForHost(cfg, "notthisone.com:8888")
	assert.NotNil(t, pool)

}

func TestCertVerifyDisabledGlobalEnv(t *testing.T) {
	cfg := config.New()
	assert.False(t, isCertVerificationDisabledForHost(cfg, "anyhost.com"))

	cfg.SetAllEnv(map[string]string{"GIT_SSL_NO_VERIFY": "1"})
	assert.True(t, isCertVerificationDisabledForHost(cfg, "anyhost.com"))
}

func TestCertVerifyDisabledGlobalConfig(t *testing.T) {
	cfg := config.New()
	assert.False(t, isCertVerificationDisabledForHost(cfg, "anyhost.com"))

	cfg.SetConfig("http.sslverify", "false")
	assert.True(t, isCertVerificationDisabledForHost(cfg, "anyhost.com"))
}

func TestCertVerifyDisabledHostConfig(t *testing.T) {
	cfg := config.New()
	assert.False(t, isCertVerificationDisabledForHost(cfg, "specifichost.com"))
	assert.False(t, isCertVerificationDisabledForHost(cfg, "otherhost.com"))

	cfg.SetConfig("http.https://specifichost.com/.sslverify", "false")
	assert.True(t, isCertVerificationDisabledForHost(cfg, "specifichost.com"))
	assert.False(t, isCertVerificationDisabledForHost(cfg, "otherhost.com"))
}
