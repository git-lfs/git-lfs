package tq

import (
	"sync"
	"testing"

	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestIsConfigurable(t *testing.T) {
	cli, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretries": "3",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	assert.Equal(t, 3, m.MaxRetries())
}

func TestManifestClampsValidValues(t *testing.T) {
	cli, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretries": "-1",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	assert.Equal(t, 8, m.MaxRetries())
}

func TestLazyManifestConcurrentUpgrade(t *testing.T) {
	cli, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, nil))
	require.Nil(t, err)

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
	cli, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretries": "not_an_int",
	}))
	require.Nil(t, err)

	m := NewManifest(nil, cli, "", "")
	assert.Equal(t, 8, m.MaxRetries())
}
