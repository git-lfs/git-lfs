package lfs

import (
	"testing"

	"github.com/git-lfs/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestManifestIsConfigurable(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "3",
		},
	})
	m := TransferManifest(cfg)
	assert.Equal(t, 3, m.MaxRetries())
}

func TestManifestChecksNTLM(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.url":                 "http://foo",
			"lfs.http://foo.access":   "ntlm",
			"lfs.concurrenttransfers": "3",
		},
	})
	m := TransferManifest(cfg)
	assert.Equal(t, 1, m.MaxRetries())
}

func TestManifestClampsValidValues(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "-1",
		},
	})
	m := TransferManifest(cfg)
	assert.Equal(t, 1, m.MaxRetries())
}

func TestManifestIgnoresNonInts(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "not_an_int",
		},
	})
	m := TransferManifest(cfg)
	assert.Equal(t, 1, m.MaxRetries())
}
