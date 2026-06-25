package lfs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/internal/testutil"
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

	r := tools.NewCallbackReader(bytes.NewReader(buf), int64(bufSize), cb)
	br := tools.NewBodyWithCallback(tools.NewClosingByteReader(buf), int64(bufSize), cb)

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

type copyCallbackFileThrottleTestMode int

const (
	initialTotalCorrect copyCallbackFileThrottleTestMode = iota
	initialTotalUnknown
)

type copyCallbackFileThrottleTestCase struct {
	mode          copyCallbackFileThrottleTestMode
	readerBufSize int64
	delayedEOF    bool

	logFilePath    string
	fakeClock      clock.FakeClock
	callbackReader *tools.CallbackReader

	expectedFractions []string
}

func (c *copyCallbackFileThrottleTestCase) setup(t *testing.T) (*os.File, error) {
	tmpDir := t.TempDir()
	c.logFilePath = tmpDir + "/git_lfs_progress.log"
	osMf := config.UniqMapFetcher(map[string]string{
		"GIT_LFS_PROGRESS": c.logFilePath,
	})
	gitMf := config.UniqMapFetcher(map[string]string{})

	c.fakeClock = clock.NewFake()
	gf := GitFilter{
		cfg: &config.Configuration{
			Os:  config.EnvironmentOf(osMf),
			Git: config.EnvironmentOf(gitMf),
		},
		clk: c.fakeClock,
	}

	cb, f, err := gf.CopyCallbackFile("clean", "test_copy", 1, 1)
	if err != nil {
		return f, fmt.Errorf("unable to create progress log callback: %w", err)
	}

	// The Assert() method makes four separate reads.
	bufSize := 4 * c.readerBufSize
	initialTotalSize := bufSize
	switch c.mode {
	case initialTotalUnknown:
		initialTotalSize = -1
	}

	buf := make([]byte, bufSize)

	var r io.Reader
	if c.delayedEOF {
		r = testutil.NewDeferredEOFByteReader(buf)
	} else {
		r = testutil.NewEagerEOFByteReader(buf)
	}

	c.callbackReader = tools.NewCallbackReader(r, initialTotalSize, cb)

	return f, err
}

func (c *copyCallbackFileThrottleTestCase) Assert(t *testing.T) {
	buf := make([]byte, c.readerBufSize)

	// Read #1: No message should be logged as the deadline has not passed.
	n, err := c.callbackReader.Read(buf)
	assert.EqualValues(t, c.readerBufSize, n)
	assert.Nil(t, err)

	c.fakeClock.Add(tasklog.DefaultLoggingThrottle)

	// Read #2: A message should be logged as the deadline has passed.
	n, err = c.callbackReader.Read(buf)
	assert.EqualValues(t, c.readerBufSize, n)
	assert.Nil(t, err)

	c.fakeClock.Add(tasklog.DefaultLoggingThrottle / 2)

	// Read #3: No message should be logged as the full deadline
	//          has not passed.
	n, err = c.callbackReader.Read(buf)
	assert.EqualValues(t, c.readerBufSize, n)
	assert.Nil(t, err)

	// Read #4: When the total size is initially unknown and EOF is
	//          delayed, no message should be logged as the full deadline
	//          has not passed.
	//
	//          Otherwise, a message should be logged even though the
	//          full deadline has not passed because the callback
	//          sees that the initial total has been reached.
	n, err = c.callbackReader.Read(buf)
	assert.EqualValues(t, c.readerBufSize, n)
	if c.delayedEOF {
		assert.Nil(t, err)
	} else {
		assert.Equal(t, io.EOF, err)
	}

	if c.delayedEOF {
		// Final EOF Read: No message should be logged because no
		//                 callback will be made when no bytes are
		//                 available to be read.
		n, err = c.callbackReader.Read(buf)
		assert.Zero(t, n)
		assert.Equal(t, io.EOF, err)
	}

	logBytes, err := os.ReadFile(c.logFilePath)
	assert.Nil(t, err)

	var expectedLog string
	for _, expectedFraction := range c.expectedFractions {
		expectedLog += fmt.Sprintf("clean 1/1 %s test_copy\n", expectedFraction)
	}

	assert.Equal(t, expectedLog, string(logBytes))
}

func TestCopyCallbackFileThrottle(t *testing.T) {
	for desc, c := range map[string]*copyCallbackFileThrottleTestCase{
		"initial total correct": {
			mode:          initialTotalCorrect,
			readerBufSize: 32 * 1024,
			expectedFractions: []string{
				"65536/131072",
				"131072/131072",
			},
		},
		"initial total unknown": {
			mode:          initialTotalUnknown,
			readerBufSize: 32 * 1024,
			expectedFractions: []string{
				"65536/-1",
			},
		},
	} {
		for _, delayedEOF := range []bool{false, true} {
			if delayedEOF {
				c.delayedEOF = true
				desc += " with delayed EOF"
			}

			f, err := c.setup(t)
			if err != nil {
				t.Error(err)
				continue
			}
			defer f.Close()

			t.Run(desc, c.Assert)
		}
	}
}
