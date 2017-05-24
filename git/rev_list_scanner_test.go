package git

import (
	"bufio"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ArgsTestCase struct {
	Left  string
	Right string
	Opt   *ScanRefsOptions

	ExpectedStdin string
	ExpectedArgs  []string
	ExpectedErr   string
}

func (c *ArgsTestCase) Assert(t *testing.T) {
	stdin, args, err := revListArgs(c.Left, c.Right, c.Opt)

	if len(c.ExpectedErr) > 0 {
		assert.EqualError(t, err, c.ExpectedErr)
	} else {
		assert.Nil(t, err)
	}

	require.Equal(t, len(c.ExpectedArgs), len(args))
	for i := 0; i < len(c.ExpectedArgs); i++ {
		assert.Equal(t, c.ExpectedArgs[i], args[i],
			"element #%d not equal: wanted %q, got %q", i, c.ExpectedArgs[i], args[i])
	}

	if stdin != nil {
		b, err := ioutil.ReadAll(stdin)
		assert.Nil(t, err)

		assert.Equal(t, c.ExpectedStdin, string(b))
	} else if len(c.ExpectedStdin) > 0 {
		t.Errorf("git: expected stdin contents %s, got none", c.ExpectedStdin)
	}
}

func TestRevListArgs(t *testing.T) {
	for desc, c := range map[string]*ArgsTestCase{
		"scan refs deleted, left and right": {
			Left: "left", Right: "right", Opt: &ScanRefsOptions{
				Mode:             ScanRefsMode,
				SkipDeletedBlobs: false,
			},
			ExpectedArgs: []string{"rev-list", "--objects", "--do-walk", "left", "right", "--"},
		},
		"scan refs not deleted, left and right": {
			Left: "left", Right: "right", Opt: &ScanRefsOptions{
				Mode:             ScanRefsMode,
				SkipDeletedBlobs: true,
			},
			ExpectedArgs: []string{"rev-list", "--objects", "--no-walk", "left", "right", "--"},
		},
		"scan refs deleted, left only": {
			Left: "left", Right: "", Opt: &ScanRefsOptions{
				Mode:             ScanRefsMode,
				SkipDeletedBlobs: false,
			},
			ExpectedArgs: []string{"rev-list", "--objects", "--do-walk", "left", "--"},
		},
		"scan refs not deleted, left only": {
			Left: "left", Right: "", Opt: &ScanRefsOptions{
				Mode:             ScanRefsMode,
				SkipDeletedBlobs: true,
			},
			ExpectedArgs: []string{"rev-list", "--objects", "--no-walk", "left", "--"},
		},
		"scan all": {
			Left: "left", Right: "right", Opt: &ScanRefsOptions{
				Mode: ScanAllMode,
			},
			ExpectedArgs: []string{"rev-list", "--objects", "--all", "--"},
		},
		"scan left to remote, no skipped refs": {
			Left: "left", Right: "right", Opt: &ScanRefsOptions{
				Mode:        ScanLeftToRemoteMode,
				Remote:      "origin",
				SkippedRefs: []string{},
			},
			ExpectedArgs: []string{"rev-list", "--objects", "left", "--not", "--remotes=origin", "--"},
		},
		"scan left to remote, skipped refs": {
			Left: "left", Right: "right", Opt: &ScanRefsOptions{
				Mode:        ScanLeftToRemoteMode,
				Remote:      "origin",
				SkippedRefs: []string{"a", "b", "c"},
			},
			ExpectedArgs:  []string{"rev-list", "--objects", "--stdin", "--"},
			ExpectedStdin: "left\na\nb\nc",
		},
		"scan unknown type": {
			Left: "left", Right: "right", Opt: &ScanRefsOptions{
				Mode: ScanningMode(-1),
			},
			ExpectedErr: "unknown scan type: -1",
		},
	} {
		t.Run(desc, c.Assert)
	}
}

func TestRevListScannerCallsClose(t *testing.T) {
	var called uint32
	err := errors.New("Hello world")

	s := &RevListScanner{
		closeFn: func() error {
			atomic.AddUint32(&called, 1)
			return err
		},
	}

	got := s.Close()

	assert.EqualValues(t, 1, atomic.LoadUint32(&called))
	assert.Equal(t, err, got)
}

func TestRevListScannerTreatsCloseFnAsOptional(t *testing.T) {
	s := &RevListScanner{
		closeFn: nil,
	}

	defer func() { assert.Nil(t, recover()) }()

	assert.Nil(t, s.Close())
}

func TestRevListScannerParsesLinesWithNames(t *testing.T) {
	given := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa name.dat"
	s := &RevListScanner{
		s: bufio.NewScanner(strings.NewReader(given)),
	}

	assert.True(t, s.Scan())
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", hex.EncodeToString(s.Object().Oid))
	assert.Equal(t, "name.dat", s.Object().Name)
	assert.Nil(t, s.Err())

	assert.False(t, s.Scan())
	assert.Nil(t, s.Object())
	assert.Nil(t, s.Err())
}

func TestRevListScannerParsesLinesWithoutName(t *testing.T) {
	given := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	s := &RevListScanner{
		s: bufio.NewScanner(strings.NewReader(given)),
	}

	assert.True(t, s.Scan())
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", hex.EncodeToString(s.Object().Oid))
	assert.Nil(t, s.Err())

	assert.False(t, s.Scan())
	assert.Nil(t, s.Object())
	assert.Nil(t, s.Err())
}
