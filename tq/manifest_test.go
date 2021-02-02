package tq

import (
	"testing"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestIsConfigurable(t *testing.T) {
	cli, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretries": "3",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	assert.Equal(t, 3, m.MaxRetries())
}

func TestManifestClampsValidValues(t *testing.T) {
	cli, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretries": "-1",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	assert.Equal(t, 8, m.MaxRetries())
}

func TestManifestIgnoresNonInts(t *testing.T) {
	cli, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretries": "not_an_int",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	assert.Equal(t, 8, m.MaxRetries())
}
