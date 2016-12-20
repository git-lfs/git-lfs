package lfsapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	c, err := NewClient(testEnv(map[string]string{}), testEnv(map[string]string{
		"lfs.dialtimeout":         "151",
		"lfs.keepalive":           "152",
		"lfs.tlstimeout":          "153",
		"lfs.concurrenttransfers": "154",
	}))

	require.Nil(t, err)
	assert.Equal(t, 151, c.DialTimeout)
	assert.Equal(t, 152, c.KeepaliveTimeout)
	assert.Equal(t, 153, c.TLSTimeout)
	assert.Equal(t, 154, c.ConcurrentTransfers)
}

func TestNewClientWithGitSSLVerify(t *testing.T) {
	c, err := NewClient(nil, nil)
	assert.Nil(t, err)
	assert.False(t, c.SkipSSLVerify)

	for _, value := range []string{"true", "1", "t"} {
		c, err = NewClient(testEnv(map[string]string{}), testEnv(map[string]string{
			"http.sslverify": value,
		}))
		t.Logf("http.sslverify: %q", value)
		assert.Nil(t, err)
		assert.False(t, c.SkipSSLVerify)
	}

	for _, value := range []string{"false", "0", "f"} {
		c, err = NewClient(testEnv(map[string]string{}), testEnv(map[string]string{
			"http.sslverify": value,
		}))
		t.Logf("http.sslverify: %q", value)
		assert.Nil(t, err)
		assert.True(t, c.SkipSSLVerify)
	}
}

func TestNewClientWithOSSSLVerify(t *testing.T) {
	c, err := NewClient(nil, nil)
	assert.Nil(t, err)
	assert.False(t, c.SkipSSLVerify)

	for _, value := range []string{"false", "0", "f"} {
		c, err = NewClient(testEnv(map[string]string{
			"GIT_SSL_NO_VERIFY": value,
		}), testEnv(map[string]string{}))
		t.Logf("GIT_SSL_NO_VERIFY: %q", value)
		assert.Nil(t, err)
		assert.False(t, c.SkipSSLVerify)
	}

	for _, value := range []string{"true", "1", "t"} {
		c, err = NewClient(testEnv(map[string]string{
			"GIT_SSL_NO_VERIFY": value,
		}), testEnv(map[string]string{}))
		t.Logf("GIT_SSL_NO_VERIFY: %q", value)
		assert.Nil(t, err)
		assert.True(t, c.SkipSSLVerify)
	}
}
