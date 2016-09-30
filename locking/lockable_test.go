package locking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFileLockable(t *testing.T) {
	defer func() {
		cachedLockablePatterns = nil
	}()
	cachedLockablePatterns = []string{
		"*.jpg", "*.png",
		"singlefile.tst",
		"subfolderall/*",
		"subfolderpart/*.sub",
		"sub/sub/folder/*.dat",
		"m?xture/of/file??match.*",
	}

	assert.True(t, IsFileLockable("test.jpg"))
	assert.True(t, IsFileLockable("something/test.jpg"))
	assert.True(t, IsFileLockable("something/something/darkside.png"))
	assert.False(t, IsFileLockable("test.txt"))
	assert.False(t, IsFileLockable("something/test.txt"))
	assert.True(t, IsFileLockable("singlefile.tst"))
	assert.False(t, IsFileLockable("otherfile.tst"))
	assert.True(t, IsFileLockable("subfolderall/anything_here 2.dat"))
	assert.True(t, IsFileLockable("subfolderpart/only certain--things.sub"))
	assert.False(t, IsFileLockable("subfolderpart/other things.sug"))
	assert.True(t, IsFileLockable("sub/sub/folder/threelevels.dat"))
	assert.False(t, IsFileLockable("sub/folder/notenoughlevels.dat"))
	assert.True(t, IsFileLockable("mixture/of/file99match.dude"))
}
