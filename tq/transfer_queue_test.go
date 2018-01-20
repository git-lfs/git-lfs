package tq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManifestDefaultsToFixedRetries(t *testing.T) {
	assert.Equal(t, 8, NewManifest(nil, nil, "", "").MaxRetries())
}

func TestRetryCounterDefaultsToFixedRetries(t *testing.T) {
	rc := newRetryCounter()
	assert.Equal(t, 8, rc.MaxRetries)
}

func TestRetryCounterIncrementsObjects(t *testing.T) {
	rc := newRetryCounter()
	rc.Increment("oid")
	assert.Equal(t, 1, rc.CountFor("oid"))
}

func TestRetryCounterCanNotRetryAfterExceedingRetryCount(t *testing.T) {
	rc := newRetryCounter()
	rc.MaxRetries = 1
	rc.Increment("oid")

	count, canRetry := rc.CanRetry("oid")
	assert.Equal(t, 1, count)
	assert.False(t, canRetry)
}

func TestBatchSizeReturnsBatchSize(t *testing.T) {
	q := NewTransferQueue(
		Upload, NewManifest(nil, nil, "", ""), "origin", WithBatchSize(3))

	assert.Equal(t, 3, q.BatchSize())
}
