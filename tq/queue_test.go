package tq_test

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/git-lfs/git-lfs/tq"
	"github.com/stretchr/testify/assert"
)

func TestQueueProcessesNewItems(t *testing.T) {
	var seen uint32

	q := tq.New(1, func(oid string) bool {
		if oid == "some-oid" {
			atomic.AddUint32(&seen, 1)
		}

		return true
	})

	q.Add("some-oid")
	q.Wait()

	assert.EqualValues(t, 1, seen)
}

func TestQueueRetriesFailedItems(t *testing.T) {
	var seen uint32

	q := tq.New(1, func(oid string) bool {
		if oid == "some-oid" {
			atomic.AddUint32(&seen, 1)
		}

		return atomic.LoadUint32(&seen) == 2
	})

	q.Add("some-oid")
	q.Wait()

	assert.EqualValues(t, 2, seen)
}

func TestQueueProcessesRetriedItemsBeforeNewItems(t *testing.T) {
	var order []string
	var retries uint32

	q := tq.New(1, func(oid string) bool {
		order = append(order, oid)

		if strings.HasSuffix(oid, "retry") {
			return atomic.AddUint32(&retries, 1) > 3
		}
		return true
	}, tq.WithBatchSize(3), tq.WithBufferDepth(4))

	q.Add("first-retry")
	q.Add("second-retry")
	q.Add("third-retry")
	q.Add("fourth-oid")

	q.Wait()

	assert.Equal(t, []string{
		"first-retry", "second-retry", "third-retry",
		"first-retry", "second-retry", "third-retry",
		"fourth-oid",
	}, order)
}
