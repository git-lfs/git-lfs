package tq

import (
	"sync"
	"testing"

	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/stretchr/testify/assert"
)

func TestManifestIsConfigurable(t *testing.T) {
	cli := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretries":    "3",
		"lfs.transfer.maxretrydelay": "20",
		"lfs.transfer.maxretryafter": "60",
	}))
	defer cli.Close()

	m := NewManifest(nil, cli, "", "")
	assert.Equal(t, 3, m.MaxRetries())
	assert.Equal(t, 20, m.MaxRetryDelay())
	assert.Equal(t, 60, m.MaxRetryAfter())
}

func TestManifestClampsValidValues(t *testing.T) {
	cli := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretries":    "-1",
		"lfs.transfer.maxretrydelay": "-1",
		"lfs.transfer.maxretryafter": "-1",
	}))
	defer cli.Close()

	m := NewManifest(nil, cli, "", "")
	assert.Equal(t, 8, m.MaxRetries())
	assert.Equal(t, 10, m.MaxRetryDelay())
	assert.Equal(t, 300, m.MaxRetryAfter())
}

func TestLazyManifestConcurrentUpgrade(t *testing.T) {
	cli := lfsapi.NewClient(lfshttp.NewContext(nil, nil, nil))
	defer cli.Close()

	m := NewManifest(nil, cli, "", "")

	// Concurrent Upgrade calls must return the same concreteManifest
	// instance and not race on the nil check.
	start := make(chan struct{})
	results := make([]*concreteManifest, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			results[idx] = m.Upgrade()
		}(i)
	}
	close(start)
	wg.Wait()

	assert.Same(t, results[0], results[1], "concurrent Upgrade returned different instances")
}

func TestManifestIgnoresNonInts(t *testing.T) {
	cli := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretries":    "not_an_int",
		"lfs.transfer.maxretrydelay": "not_an_int",
		"lfs.transfer.maxretryafter": "not_an_int",
	}))
	defer cli.Close()

	m := NewManifest(nil, cli, "", "")
	assert.Equal(t, 8, m.MaxRetries())
	assert.Equal(t, 10, m.MaxRetryDelay())
	assert.Equal(t, 300, m.MaxRetryAfter())
}
