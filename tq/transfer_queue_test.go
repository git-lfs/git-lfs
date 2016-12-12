package tq

import (
	"testing"

	"github.com/git-lfs/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestRetryCounterDefaultsToFixedRetries(t *testing.T) {
	rc := newRetryCounter(config.NewFrom(config.Values{}))

	assert.Equal(t, 1, rc.MaxRetries)
}

func TestRetryCounterIsConfigurable(t *testing.T) {
	rc := newRetryCounter(config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "3",
		},
	}))

	assert.Equal(t, 3, rc.MaxRetries)
}

func TestRetryCounterClampsValidValues(t *testing.T) {
	rc := newRetryCounter(config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "-1",
		},
	}))

	assert.Equal(t, 1, rc.MaxRetries)
}

func TestRetryCounterIgnoresNonInts(t *testing.T) {
	rc := newRetryCounter(config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "not_an_int",
		},
	}))

	assert.Equal(t, 1, rc.MaxRetries)
}

func TestRetryCounterIncrementsObjects(t *testing.T) {
	rc := newRetryCounter(config.NewFrom(config.Values{}))

	rc.Increment("oid")

	assert.Equal(t, 1, rc.CountFor("oid"))
}

func TestRetryCounterCanNotRetryAfterExceedingRetryCount(t *testing.T) {
	rc := newRetryCounter(config.NewFrom(config.Values{}))

	rc.Increment("oid")
	count, canRetry := rc.CanRetry("oid")

	assert.Equal(t, 1, count)
	assert.False(t, canRetry)
}
