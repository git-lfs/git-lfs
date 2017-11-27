package git

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ArgsTestCase struct {
	Include []string
	Exclude []string
	Opt     *ScanRefsOptions

	ExpectedStdin string
	ExpectedArgs  []string
	ExpectedErr   string
}

func (c *ArgsTestCase) Assert(t *testing.T) {
	stdin, args, err := revListArgs(c.Include, c.Exclude, c.Opt)

	if len(c.ExpectedErr) > 0 {
		assert.EqualError(t, err, c.ExpectedErr)
	} else {
		assert.Nil(t, err)
	}

	assert.EqualValues(t, c.ExpectedArgs, args)

	if stdin != nil {
		b, err := ioutil.ReadAll(stdin)
		assert.Nil(t, err)

		assert.Equal(t, c.ExpectedStdin, string(b))
	} else if len(c.ExpectedStdin) > 0 {
		t.Errorf("git: expected stdin contents %s, got none", c.ExpectedStdin)
	}
}

var (
	s1 = "decafdecafdecafdecafdecafdecafdecafdecaf"
	s2 = "cafecafecafecafecafecafecafecafecafecafe"
)

func TestRevListArgs(t *testing.T) {
	for desc, c := range map[string]*ArgsTestCase{
		"scan refs deleted, left and right": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode:             ScanRefsMode,
				SkipDeletedBlobs: false,
			},
			ExpectedStdin: fmt.Sprintf("%s\n^%s", s1, s2),
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--do-walk", "--"},
		},
		"scan refs not deleted, left and right": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode:             ScanRefsMode,
				SkipDeletedBlobs: true,
			},
			ExpectedStdin: fmt.Sprintf("%s\n^%s", s1, s2),
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--no-walk", "--"},
		},
		"scan refs deleted, left only": {
			Include: []string{s1}, Opt: &ScanRefsOptions{
				Mode:             ScanRefsMode,
				SkipDeletedBlobs: false,
			},
			ExpectedStdin: s1,
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--do-walk", "--"},
		},
		"scan refs not deleted, left only": {
			Include: []string{s1}, Opt: &ScanRefsOptions{
				Mode:             ScanRefsMode,
				SkipDeletedBlobs: true,
			},
			ExpectedStdin: s1,
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--no-walk", "--"},
		},
		"scan all": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode: ScanAllMode,
			},
			ExpectedArgs: []string{"rev-list", "--stdin", "--objects", "--all", "--"},
		},
		"scan left to remote, no skipped refs": {
			Include: []string{s1}, Opt: &ScanRefsOptions{
				Mode:        ScanLeftToRemoteMode,
				Remote:      "origin",
				SkippedRefs: []string{},
			},
			ExpectedStdin: s1,
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--not", "--remotes=origin", "--"},
		},
		"scan left to remote, skipped refs": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode:        ScanLeftToRemoteMode,
				Remote:      "origin",
				SkippedRefs: []string{"a", "b", "c"},
			},
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--"},
			ExpectedStdin: s1 + "\n^" + s2 + "\na\nb\nc",
		},
		"scan unknown type": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode: ScanningMode(-1),
			},
			ExpectedErr: "unknown scan type: -1",
		},
		"scan date order": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode:  ScanRefsMode,
				Order: DateRevListOrder,
			},
			ExpectedStdin: fmt.Sprintf("%s\n^%s", s1, s2),
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--date-order", "--do-walk", "--"},
		},
		"scan author date order": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode:  ScanRefsMode,
				Order: AuthorDateRevListOrder,
			},
			ExpectedStdin: fmt.Sprintf("%s\n^%s", s1, s2),
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--author-date-order", "--do-walk", "--"},
		},
		"scan topo order": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode:  ScanRefsMode,
				Order: TopoRevListOrder,
			},
			ExpectedStdin: fmt.Sprintf("%s\n^%s", s1, s2),
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--topo-order", "--do-walk", "--"},
		},
		"scan commits only": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode:        ScanRefsMode,
				CommitsOnly: true,
			},
			ExpectedStdin: fmt.Sprintf("%s\n^%s", s1, s2),
			ExpectedArgs:  []string{"rev-list", "--stdin", "--do-walk", "--"},
		},
		"scan reverse": {
			Include: []string{s1}, Exclude: []string{s2}, Opt: &ScanRefsOptions{
				Mode:    ScanRefsMode,
				Reverse: true,
			},
			ExpectedStdin: fmt.Sprintf("%s\n^%s", s1, s2),
			ExpectedArgs:  []string{"rev-list", "--stdin", "--objects", "--reverse", "--do-walk", "--"},
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
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", hex.EncodeToString(s.OID()))
	assert.Equal(t, "name.dat", s.Name())
	assert.Nil(t, s.Err())

	assert.False(t, s.Scan())
	assert.Equal(t, "", s.Name())
	assert.Nil(t, s.OID())
	assert.Nil(t, s.Err())
}

func TestRevListScannerParsesLinesWithoutName(t *testing.T) {
	given := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	s := &RevListScanner{
		s: bufio.NewScanner(strings.NewReader(given)),
	}

	assert.True(t, s.Scan())
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", hex.EncodeToString(s.OID()))
	assert.Nil(t, s.Err())

	assert.False(t, s.Scan())
	assert.Equal(t, "", s.Name())
	assert.Nil(t, s.OID())
	assert.Nil(t, s.Err())
}
