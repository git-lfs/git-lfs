package tq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManifestDefaultsToFixedRetries(t *testing.T) {
	assert.Equal(t, 8, NewManifest(nil, nil, "", "").MaxRetries())
}

func TestManifestDefaultsToFixedRetryDelay(t *testing.T) {
	assert.Equal(t, 10, NewManifest(nil, nil, "", "").MaxRetryDelay())
}

func TestRetryCounterDefaultsToFixedRetries(t *testing.T) {
	rc := newRetryCounter()
	assert.Equal(t, 8, rc.MaxRetries)
}

func TestRetryCounterDefaultsToFixedRetryDelay(t *testing.T) {
	rc := newRetryCounter()
	assert.Equal(t, 10, rc.MaxRetryDelay)
}

func TestRetryCounterIncrementsObjects(t *testing.T) {
	rc := newRetryCounter()
	assert.Equal(t, 1, rc.Increment("oid"))
	assert.Equal(t, 1, rc.CountFor("oid"))

	assert.Equal(t, 2, rc.Increment("oid"))
	assert.Equal(t, 2, rc.CountFor("oid"))
}

func TestRetryCounterCanNotRetryAfterExceedingRetryCount(t *testing.T) {
	rc := newRetryCounter()
	rc.MaxRetries = 1
	rc.Increment("oid")

	count, canRetry := rc.CanRetry("oid")
	assert.Equal(t, 1, count)
	assert.False(t, canRetry)
}

func TestRetryCounterDoesNotDelayFirstAttempt(t *testing.T) {
	rc := newRetryCounter()
	assert.Equal(t, time.Time{}, rc.ReadyTime("oid"))
}

func TestRetryCounterDelaysExponentially(t *testing.T) {
	rc := newRetryCounter()
	start := time.Now()

	rc.Increment("oid")
	ready1 := rc.ReadyTime("oid")
	assert.GreaterOrEqual(t, int64(ready1.Sub(start)/time.Millisecond), int64(baseRetryDelayMs))

	rc.Increment("oid")
	ready2 := rc.ReadyTime("oid")
	assert.GreaterOrEqual(t, int64(ready2.Sub(start)/time.Millisecond), int64(2*baseRetryDelayMs))
}

func TestRetryCounterLimitsDelay(t *testing.T) {
	rc := newRetryCounter()
	rc.MaxRetryDelay = 1

	for i := 0; i < 4; i++ {
		rc.Increment("oid")
	}

	rt := rc.ReadyTime("oid")
	assert.WithinDuration(t, time.Now(), rt, 1*time.Second)
}

func TestBatchSizeReturnsBatchSize(t *testing.T) {
	q := NewTransferQueue(
		Upload, NewManifest(nil, nil, "", ""), "origin", WithBatchSize(3))

	assert.Equal(t, 3, q.BatchSize())
}

func TestUseAdapterReusesWhenNameMatches(t *testing.T) {
	q := NewTransferQueue(
		Download, NewManifest(nil, nil, "", ""), "origin")

	// Set an initial adapter.
	q.useAdapter("basic")
	first := q.adapter
	assert.NotNil(t, first)
	assert.Equal(t, "basic", first.Name())

	// Calling with the same name should reuse the adapter instance.
	q.useAdapter("basic")
	assert.Same(t, first, q.adapter, "expected adapter to be reused when name matches")
}

func TestUseAdapterReusesWhenNameIsEmpty(t *testing.T) {
	q := NewTransferQueue(
		Download, NewManifest(nil, nil, "", ""), "origin")

	q.useAdapter("basic")
	first := q.adapter
	assert.NotNil(t, first)

	// An empty name means "use basic" per the spec. Since the current
	// adapter is already basic, it should be reused.
	q.useAdapter("")
	assert.Same(t, first, q.adapter, "expected basic adapter to be reused when name is empty")
}

func TestUseAdapterSwitchesFromNonDefaultWhenNameIsEmpty(t *testing.T) {
	q := NewTransferQueue(
		Download, NewManifest(nil, nil, "", ""), "origin")

	q.useAdapter("ssh")
	first := q.adapter
	assert.NotNil(t, first)
	assert.Equal(t, "ssh", first.Name())

	// An empty name means "use basic" per the spec, so it should
	// switch away from the SSH adapter.
	q.useAdapter("")
	assert.NotSame(t, first, q.adapter, "expected adapter to switch from ssh to basic")
	assert.Equal(t, "basic", q.adapter.Name())
}

func TestUseAdapterSwitchesWhenNameDiffers(t *testing.T) {
	q := NewTransferQueue(
		Download, NewManifest(nil, nil, "", ""), "origin")

	q.useAdapter("basic")
	first := q.adapter
	assert.NotNil(t, first)

	// A different, non-empty name should cause the adapter to switch.
	q.useAdapter("ssh")
	assert.NotNil(t, q.adapter)
	assert.NotSame(t, first, q.adapter, "expected a new adapter when name differs")
	assert.Equal(t, "ssh", q.adapter.Name())
}
