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
