package lfs

import (
	"bytes"
	"os"
	"testing"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/tasklog"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/assert"
)

func TestBodyWithCallback(t *testing.T) {
	called := 0
	calledRead := make([]int64, 0, 2)

	cb := func(total int64, read int64, current int) error {
		called += 1
		calledRead = append(calledRead, read)
		assert.Equal(t, 5, int(total))
		return nil
	}
	reader := tools.NewByteBodyWithCallback([]byte("BOOYA"), 5, cb)

	readBuf := make([]byte, 3)
	n, err := reader.Read(readBuf)
	assert.Nil(t, err)
	assert.Equal(t, "BOO", string(readBuf[0:n]))

	n, err = reader.Read(readBuf)
	assert.Nil(t, err)
	assert.Equal(t, "YA", string(readBuf[0:n]))

	assert.Equal(t, 2, called)
	assert.Len(t, calledRead, 2)
	assert.Equal(t, 3, int(calledRead[0]))
	assert.Equal(t, 5, int(calledRead[1]))
}

func TestReadWithCallback(t *testing.T) {
	called := 0
	calledRead := make([]int64, 0, 2)

	reader := &tools.CallbackReader{
		TotalSize: 5,
		Reader:    bytes.NewBufferString("BOOYA"),
		C: func(total int64, read int64, current int) error {
			called += 1
			calledRead = append(calledRead, read)
			assert.Equal(t, 5, int(total))
			return nil
		},
	}

	readBuf := make([]byte, 3)
	n, err := reader.Read(readBuf)
	assert.Nil(t, err)
	assert.Equal(t, "BOO", string(readBuf[0:n]))

	n, err = reader.Read(readBuf)
	assert.Nil(t, err)
	assert.Equal(t, "YA", string(readBuf[0:n]))

	assert.Equal(t, 2, called)
	assert.Len(t, calledRead, 2)
	assert.Equal(t, 3, int(calledRead[0]))
	assert.Equal(t, 5, int(calledRead[1]))
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
