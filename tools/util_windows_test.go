// +build windows

package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/git-lfs/git-lfs/errors"

	"github.com/stretchr/testify/assert"
)

func TestCloneFile(t *testing.T) {
	testDir := os.Getenv("REFS_TEST_DIR")
	if testDir == "" {
		testDir, _ = Getwd()
	}

	t.Logf("testing on: %s", testDir)

	supported, err := CheckCloneFileSupported(testDir)
	if err != nil || !supported {
		t.Skip(err)
	}

	testCases := []struct {
		name string
		size int64
	}{
		{"Small", 123},
		{"Smaller than 4K", 4*1024 - 1},
		{"Equal to 4K", 4 * 1024},
		{"Larger than 4K", 4*1024 + 1},
		{"Smaller than 64K", 64*1024 - 1},
		{"Equal to 64K", 64 * 1024},
		{"Larger than 64K", 64*1024 + 1},
		{"Large", 12345678},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			as := assert.New(t)

			src, err := ioutil.TempFile(testDir, tc.name+"_src")
			as.NoError(err)
			dst, err := ioutil.TempFile(testDir, tc.name+"_dst")
			as.NoError(err)

			srcHash, err := fillFile(src, tc.size)
			as.NoError(err)

			ok, err := CloneFile(dst, src)
			as.NoError(err)
			as.True(ok)

			sha := sha256.New()
			dst.Seek(0, io.SeekStart)
			io.Copy(sha, dst)
			dstHash := hex.EncodeToString(sha.Sum(nil))

			as.Equal(srcHash, dstHash)
		})
	}
}

func fillFile(target *os.File, size int64) (hash string, err error) {
	str := make([]byte, 1024)
	for i := 0; i < 1023; i++ {
		str[i] = fmt.Sprintf("%x", i%16)[0]
	}
	str[1023] = '\n'

	for i := int64(0); i < size; i += 1024 {
		_, err := target.Write(str)
		if err != nil {
			panic(err)
		}
	}

	err = target.Truncate(size)
	if err != nil {
		return "", err
	}

	err = target.Sync()
	if err != nil {
		return "", err
	}

	_, err = target.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	sha := sha256.New()
	copySize, err := io.Copy(sha, target)
	if err != nil {
		return "", err
	}
	if size != copySize {
		return "", errors.New("size mismatch")
	}

	return hex.EncodeToString(sha.Sum(nil)), nil
}

func TestTryRenameDestExists(t *testing.T) {
	source, err := ioutil.TempFile("", "source")
	assert.NoError(t, err)
	assert.NoError(t, source.Close())
	defer os.Remove(source.Name())

	sourceData := []byte("source")
	assert.NoError(t, ioutil.WriteFile(source.Name(), sourceData, 0644))

	dest, err := ioutil.TempFile("", "dest")
	assert.NoError(t, err)
	defer os.Remove(dest.Name())
	assert.NoError(t, dest.Close())

	destData := []byte("dest")
	assert.NoError(t, ioutil.WriteFile(dest.Name(), destData, 0644))

	// Perform rename
	err = TryRename(source.Name(), dest.Name())
	assert.True(t, os.IsExist(err))

	sourceData2, err := ioutil.ReadFile(source.Name())
	assert.NoError(t, err)
	assert.Equal(t, sourceData, sourceData2)

	destData2, err := ioutil.ReadFile(dest.Name())
	assert.NoError(t, err)
	assert.Equal(t, destData, destData2)
}

func TestTryRename(t *testing.T) {
	source, err := ioutil.TempFile("", "source")
	assert.NoError(t, err)
	assert.NoError(t, source.Close())
	defer os.Remove(source.Name())

	sourceData := []byte("source")
	assert.NoError(t, ioutil.WriteFile(source.Name(), sourceData, 0644))

	dest, err := ioutil.TempFile("", "dest")
	assert.NoError(t, err)
	defer os.Remove(dest.Name())
	assert.NoError(t, dest.Close())

	// Remove destination file
	assert.NoError(t, os.Remove(dest.Name()))

	// Perform rename
	assert.NoError(t, TryRename(source.Name(), dest.Name()))

	_, err = os.Stat(source.Name())
	assert.True(t, os.IsNotExist(err))

	destData, err := ioutil.ReadFile(dest.Name())
	assert.NoError(t, err)
	assert.Equal(t, sourceData, destData)
}
