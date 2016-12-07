package tq

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerQueueProcessesBatchedItems(t *testing.T) {
	var called uint32

	wq := newWorkerQueue(1, func(oid string) bool {
		if oid == "some-oid" {
			atomic.AddUint32(&called, 1)
		}

		return true
	})

	retries := wq.Add([]string{"some-oid"})
	wq.Wait()

	assertChannelEmpty(t, retries, 100*time.Millisecond)
	assert.EqualValues(t, 1, called)
}

func TestWorkerQueueReturnsRetriedItems(t *testing.T) {
	retried := make(map[string]bool)

	wq := newWorkerQueue(1, func(oid string) bool {
		retried[oid] = false

		return false
	})

	retries := wq.Add([]string{"first-oid", "second-oid"})
	wq.Wait()

L:
	for {
		select {
		case retry, ok := <-retries:
			if !ok {
				break L
			}

			retried[retry] = true
		case <-time.After(100 * time.Millisecond):
			t.Errorf("tq: expected `retries` to be closed, wasn't after 100msec")

			break L
		}
	}

	assert.Len(t, retried, 2)
	assert.True(t, retried["first-oid"])
	assert.True(t, retried["second-oid"])
}

func TestWorkerQueueDistributesItemsAcrossManyWorkers(t *testing.T) {
	seen := make(map[string]bool)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	wq := newWorkerQueue(2, func(oid string) bool {
		wg.Done()
		wg.Wait()

		seen[oid] = true

		return true
	})

	retries := wq.Add([]string{"first-oid", "second-oid"})
	wq.Wait()

	assertChannelEmpty(t, retries, 100*time.Millisecond)

	assert.Len(t, seen, 2)
	assert.True(t, seen["first-oid"])
	assert.True(t, seen["second-oid"])
}

func assertChannelEmpty(t *testing.T, from <-chan string, after time.Duration) {
L:
	for {
		select {
		case v, ok := <-from:
			if !ok {
				break L
			}

			t.Errorf("tq: expected channel to be empty, got '%+v'", v)
		case <-time.After(after):
			t.Errorf("tq: expected channel to be closed after %s, wasn't", after)
		}
	}
}
