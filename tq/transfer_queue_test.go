package tq

import (
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/stretchr/testify/assert"
)

func TestManifestDefaultsToFixedRetries(t *testing.T) {
	cli := lfsapi.NewClient(nil)
	defer cli.Close()

	assert.Equal(t, 8, NewManifest(nil, cli, "", "").MaxRetries())
}

func TestManifestDefaultsToFixedRetryDelay(t *testing.T) {
	cli := lfsapi.NewClient(nil)
	defer cli.Close()

	assert.Equal(t, 10, NewManifest(nil, cli, "", "").MaxRetryDelay())
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
	cli := lfsapi.NewClient(nil)
	defer cli.Close()

	q := NewTransferQueue(
		Upload, NewManifest(nil, cli, "", ""), "origin", WithBatchSize(3),
	)

	assert.Equal(t, 3, q.BatchSize())
}

func TestUseAdapterReusesWhenNameMatches(t *testing.T) {
	cli := lfsapi.NewClient(nil)
	defer cli.Close()

	q := NewTransferQueue(
		Download, NewManifest(nil, cli, "", ""), "origin",
	)

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
	cli := lfsapi.NewClient(nil)
	defer cli.Close()

	q := NewTransferQueue(
		Download, NewManifest(nil, cli, "", ""), "origin",
	)

	q.useAdapter("basic")
	first := q.adapter
	assert.NotNil(t, first)

	// An empty name means "use basic" per the spec. Since the current
	// adapter is already basic, it should be reused.
	q.useAdapter("")
	assert.Same(t, first, q.adapter, "expected basic adapter to be reused when name is empty")
}

func TestUseAdapterSwitchesFromNonDefaultWhenNameIsEmpty(t *testing.T) {
	cli := lfsapi.NewClient(nil)
	defer cli.Close()

	q := NewTransferQueue(
		Download, NewManifest(nil, cli, "", ""), "origin",
	)

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
	cli := lfsapi.NewClient(nil)
	defer cli.Close()

	q := NewTransferQueue(
		Download, NewManifest(nil, cli, "", ""), "origin",
	)

	q.useAdapter("basic")
	first := q.adapter
	assert.NotNil(t, first)

	// A different, non-empty name should cause the adapter to switch.
	q.useAdapter("ssh")
	assert.NotNil(t, q.adapter)
	assert.NotSame(t, first, q.adapter, "expected a new adapter when name differs")
	assert.Equal(t, "ssh", q.adapter.Name())
}

func TestCanRetryObjectLater(t *testing.T) {
	cli := lfsapi.NewClient(
		lfshttp.NewContext(
			nil,
			nil,
			map[string]string{
				"lfs.transfer.maxretryafter": "5",
			},
		),
	)
	defer cli.Close()

	tests := []struct {
		name              string
		err               error
		setup             func(*TransferQueue)
		wantCanRetry      bool
		wantReadyTimeZero bool
	}{
		{
			name: "retry count exceeded",
			err:  errors.NewRetriableLaterError(errors.New("rate limit"), "10"),
			setup: func(q *TransferQueue) {
				q.rc.MaxRetries = 1
				q.rc.Increment("oid")
			},
			wantCanRetry:      false,
			wantReadyTimeZero: true,
		},
		{
			name:              "non-retriable-later error",
			err:               errors.NewRetriableError(errors.New("network error")),
			wantCanRetry:      false,
			wantReadyTimeZero: true,
		},
		{
			name:              "retry-after exceeds threshold",
			err:               errors.NewRetriableLaterError(errors.New("rate limit"), "10"),
			wantCanRetry:      false,
			wantReadyTimeZero: true,
		},
		{
			name:              "retry-after within threshold",
			err:               errors.NewRetriableLaterError(errors.New("rate limit"), "3"),
			wantCanRetry:      true,
			wantReadyTimeZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewTransferQueue(
				Download,
				NewManifest(nil, cli, "", ""),
				"origin",
			)

			if tt.setup != nil {
				tt.setup(q)
			}

			retryAt, canRetry := q.canRetryObjectLater("oid", tt.err)

			assert.Equal(t, tt.wantCanRetry, canRetry)
			assert.Equal(t, tt.wantReadyTimeZero, retryAt.IsZero())
		})
	}
}
