package lfs

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/tasklog"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/assert"
)

func TestBothCallbackReadersWithCallback(t *testing.T) {
	var calls int
	allReadSoFar := make([]int64, 0, 2)

	buf := []byte("BOOYA")
	bufSize := len(buf)

	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		calls++
		allReadSoFar = append(allReadSoFar, readSoFar)

		assert.EqualValues(t, bufSize, totalSize)

		return nil
	}

	readBuf := make([]byte, bufSize-2)
	readBufSize := len(readBuf)

	r := &tools.CallbackReader{
		C:         cb,
		TotalSize: int64(bufSize),
		Reader:    bytes.NewReader(buf),
	}
	br := tools.NewByteBodyWithCallback(buf, int64(bufSize), cb)

	for _, reader := range []io.Reader{r, br} {
		t.Logf("testing with reader: %T", reader)

		n, err := reader.Read(readBuf)

		// The underlying bytes.Reader should always return
		// a nil error when the last byte is read.
		assert.Equal(t, readBufSize, n)
		assert.Nil(t, err)
		assert.Equal(t, buf[:readBufSize], readBuf)

		n, err = reader.Read(readBuf)

		assert.Equal(t, bufSize-readBufSize, n)
		assert.Nil(t, err)
		assert.Equal(t, buf[readBufSize:], readBuf[:bufSize-readBufSize])

		assert.Equal(t, 2, calls)
		assert.Len(t, allReadSoFar, 2)
		assert.EqualValues(t, readBufSize, allReadSoFar[0])
		assert.EqualValues(t, bufSize, allReadSoFar[1])

		calls = 0
		allReadSoFar = allReadSoFar[:0]
	}
}

func TestCopyCallbackFileThrottle(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := tmpDir + "/git_lfs_progress.log"
	osMf := config.UniqMapFetcher(map[string]string{
		"GIT_LFS_PROGRESS": logFile,
	})
	gitMf := config.UniqMapFetcher(map[string]string{})

	fc := clock.NewFake()
	gf := GitFilter{
		cfg: &config.Configuration{
			Os:  config.EnvironmentOf(osMf),
			Git: config.EnvironmentOf(gitMf),
		},
		clk: fc,
	}

	bufSize := int64(128 * 1024)
	cb, f, err := gf.CopyCallbackFile("clean", "test_copy", 1, 1)
	assert.NoError(t, err)
	defer f.Close()

	r := &tools.CallbackReader{
		TotalSize: bufSize,
		Reader:    bytes.NewReader(make([]byte, bufSize)),
		C:         cb,
	}
	readbuf := make([]byte, 32*1024)
	r.Read(readbuf) // message skipped

	fc.Add(tasklog.DefaultLoggingThrottle)
	r.Read(readbuf) // message logged due to delay

	fc.Add(tasklog.DefaultLoggingThrottle / 2)
	r.Read(readbuf) // message skipped

	r.Read(readbuf) // message logged because reader is finished

	logBytes, err := os.ReadFile(logFile)
	assert.Nil(t, err)

	expectedLog := "clean 1/1 65536/131072 test_copy\nclean 1/1 131072/131072 test_copy\n"
	assert.Equal(t, expectedLog, string(logBytes))
}
