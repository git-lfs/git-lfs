package tq

import (
	"fmt"
	"sync"
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

// TestPipelineBatchDoneNotStarved verifies that the collectBatches loop
// drains batchDone signals while still accepting new work. Without this,
// batch workers block sending to batchDone (which was only drained
// during shutdown), preventing them from reading the next batch and
// deadlocking the collector's send to batchCh.
func TestPipelineBatchDoneNotStarved(t *testing.T) {
	const (
		numObjects      = 300
		concurrentXfers = 4
		batchWorkers    = 4
		batchSize       = 20
	)

	adapter := &fakeSlowAdapter{jobs: make(chan *job, 100)}

	m := &concreteManifest{
		maxRetries:              1,
		maxRetryDelay:           1,
		concurrentTransfers:     concurrentXfers,
		concurrentBatchRequests: batchWorkers,
		standaloneTransferAgent: "fake",
		downloadAdapterFuncs:    map[string]NewAdapterFunc{},
		uploadAdapterFuncs:      map[string]NewAdapterFunc{},
	}
	m.RegisterNewAdapterFunc("fake", Download, func(string, Direction) Adapter {
		return adapter
	})

	q := NewTransferQueue(Download, m, "origin", WithBatchSize(batchSize))
	watch := q.Watch()

	var received int
	watchDone := make(chan struct{})
	go func() {
		for range watch {
			received++
		}
		close(watchDone)
	}()

	addDone := make(chan struct{})
	go func() {
		for i := 0; i < numObjects; i++ {
			q.Add(fmt.Sprintf("file-%d", i), "", fmt.Sprintf("%064x", i), int64(i+1), false, nil)
		}
		close(addDone)
	}()

	select {
	case <-addDone:
	case <-time.After(time.Second):
		t.Fatal("deadlock: Add() blocked")
	}

	waitDone := make(chan struct{})
	go func() {
		q.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-time.After(time.Second):
		t.Fatal("deadlock: queue did not finish")
	}

	<-watchDone
	assert.Equal(t, numObjects, received)
	assert.Empty(t, q.Errors())
}

// fakeSlowAdapter is a minimal Adapter for testing the transfer pipeline.
type fakeSlowAdapter struct {
	jobs    chan *job
	workers sync.WaitGroup
	jobWait sync.WaitGroup
}

func (a *fakeSlowAdapter) Name() string         { return "fake" }
func (a *fakeSlowAdapter) Direction() Direction { return Download }

func (a *fakeSlowAdapter) Begin(cfg AdapterConfig, cb ProgressCallback) error {
	n := cfg.ConcurrentTransfers()
	a.workers.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer a.workers.Done()
			for j := range a.jobs {
				j.results <- TransferResult{Transfer: j.T}
				j.wg.Done()
				j.done.Done()
			}
		}()
	}
	return nil
}

func (a *fakeSlowAdapter) Add(transfers ...*Transfer) <-chan TransferResult {
	results := make(chan TransferResult, len(transfers))
	var done sync.WaitGroup
	done.Add(len(transfers))
	a.jobWait.Add(len(transfers))
	go func() {
		for _, t := range transfers {
			a.jobs <- &job{T: t, results: results, wg: &a.jobWait, done: &done}
		}
		done.Wait()
		close(results)
	}()
	return results
}

func (a *fakeSlowAdapter) End() {
	a.jobWait.Wait()
	close(a.jobs)
	a.workers.Wait()
}
