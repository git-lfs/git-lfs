//go:build windows
// +build windows

package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/git-lfs/git-lfs/v3/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

			src, err := os.CreateTemp(testDir, tc.name+"_src")
			require.NoError(t, err)
			srcName := src.Name()
			t.Cleanup(func() { os.Remove(srcName) })

			dst, err := os.CreateTemp(testDir, tc.name+"_dst")
			require.NoError(t, err)
			defer dst.Close()
			dstName := dst.Name()
			t.Cleanup(func() { os.Remove(dstName) })

			srcHash, err := fillFile(src, tc.size)
			require.NoError(t, err)
			require.NoError(t, src.Close())

			// Checkout opens LFS objects read-only before attempting a clone.
			src, err = os.Open(srcName)
			require.NoError(t, err)
			defer src.Close()

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

func TestSyncFileBeforeCloneWithReadOnlyHandle(t *testing.T) {
	writer, err := os.CreateTemp(t.TempDir(), "src")
	require.NoError(t, err)
	defer writer.Close()

	_, err = writer.WriteString("unflushed content")
	require.NoError(t, err)

	src, err := os.Open(writer.Name())
	require.NoError(t, err)
	defer src.Close()

	require.NoError(t, syncFileBeforeClone(src))
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
