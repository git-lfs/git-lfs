package tq

import (
	"testing"

	"github.com/git-lfs/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestManifestDefaultsToFixedRetries(t *testing.T) {
	cfg := config.NewFrom(config.Values{})
	m := ConfigureManifest(NewManifest(), cfg)
	assert.Equal(t, 1, m.MaxRetries)
}

func TestManifestIsConfigurable(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "3",
		},
	})
	m := ConfigureManifest(NewManifest(), cfg)
	assert.Equal(t, 3, m.MaxRetries)
}

func TestManifestClampsValidValues(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "-1",
		},
	})
	m := ConfigureManifest(NewManifest(), cfg)
	assert.Equal(t, 1, m.MaxRetries)
}

func TestManifestIgnoresNonInts(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "not_an_int",
		},
	})
	m := ConfigureManifest(NewManifest(), cfg)
	assert.Equal(t, 1, m.MaxRetries)
}

func TestRetryCounterDefaultsToFixedRetries(t *testing.T) {
	rc := newRetryCounter()
	assert.Equal(t, 1, rc.MaxRetries)
}

func TestRetryCounterIncrementsObjects(t *testing.T) {
	rc := newRetryCounter()
	rc.Increment("oid")
	assert.Equal(t, 1, rc.CountFor("oid"))
}

func TestRetryCounterCanNotRetryAfterExceedingRetryCount(t *testing.T) {
	rc := newRetryCounter()
	rc.Increment("oid")

	count, canRetry := rc.CanRetry("oid")
	assert.Equal(t, 1, count)
	assert.False(t, canRetry)
}
