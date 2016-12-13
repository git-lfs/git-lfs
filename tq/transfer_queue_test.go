package tq

import (
	"testing"

	"github.com/git-lfs/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestRetryCounterDefaultsToFixedRetries(t *testing.T) {
	rc := newRetryCounter()
	assert.Equal(t, 1, rc.MaxRetries)
}

func TestRetryCounterIsConfigurable(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "3",
		},
	})

	tr := NewTransferQueue(Download, WithGitEnv(cfg.Git))
	assert.Equal(t, 3, tr.rc.MaxRetries)
}

func TestRetryCounterClampsValidValues(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "-1",
		},
	})

	tr := NewTransferQueue(Download, WithGitEnv(cfg.Git))
	assert.Equal(t, 1, tr.rc.MaxRetries)
}

func TestRetryCounterIgnoresNonInts(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.transfer.maxretries": "not_an_int",
		},
	})

	tr := NewTransferQueue(Download, WithGitEnv(cfg.Git))
	assert.Equal(t, 1, tr.rc.MaxRetries)
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
