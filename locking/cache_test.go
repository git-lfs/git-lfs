package locking

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockCache(t *testing.T) {
	var err error

	tmpf, err := ioutil.TempFile("", "testCacheLock")
	assert.Nil(t, err)
	defer func() {
		os.Remove(tmpf.Name())
	}()
	tmpf.Close()

	cache, err := NewLockCache(tmpf.Name())
	assert.Nil(t, err)

	testLocks := []Lock{
		Lock{Path: "folder/test1.dat", Id: "101"},
		Lock{Path: "folder/test2.dat", Id: "102"},
		Lock{Path: "root.dat", Id: "103"},
	}

	for _, l := range testLocks {
		err = cache.Add(l)
		assert.Nil(t, err)
	}

	locks := cache.Locks()
	for _, l := range testLocks {
		assert.Contains(t, locks, l)
	}
	assert.Equal(t, len(testLocks), len(locks))

	err = cache.RemoveByPath("folder/test2.dat")
	assert.Nil(t, err)

	locks = cache.Locks()
	// delete item 1 from test locls
	testLocks = append(testLocks[:1], testLocks[2:]...)
	for _, l := range testLocks {
		assert.Contains(t, locks, l)
	}
	assert.Equal(t, len(testLocks), len(locks))

	err = cache.RemoveById("101")
	assert.Nil(t, err)

	locks = cache.Locks()
	testLocks = testLocks[1:]
	for _, l := range testLocks {
		assert.Contains(t, locks, l)
	}
	assert.Equal(t, len(testLocks), len(locks))
}
