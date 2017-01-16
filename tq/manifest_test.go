package tq

import (
	"testing"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestIsConfigurable(t *testing.T) {
	cli, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.transfer.maxretries": "3",
	}))
	require.Nil(t, err)

	m := NewManifestWithClient(cli)
	assert.Equal(t, 3, m.MaxRetries())
}

func TestManifestChecksNTLM(t *testing.T) {
	cli, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.url":                 "http://foo",
		"lfs.http://foo.access":   "ntlm",
		"lfs.concurrenttransfers": "3",
	}))
	require.Nil(t, err)

	m := NewManifestWithClient(cli)
	assert.Equal(t, 1, m.MaxRetries())
}

func TestManifestClampsValidValues(t *testing.T) {
	cli, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.transfer.maxretries": "-1",
	}))
	require.Nil(t, err)

	m := NewManifestWithClient(cli)
	assert.Equal(t, 1, m.MaxRetries())
}

func TestManifestIgnoresNonInts(t *testing.T) {
	cli, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.transfer.maxretries": "not_an_int",
	}))
	require.Nil(t, err)

	m := NewManifestWithClient(cli)
	assert.Equal(t, 1, m.MaxRetries())
}
