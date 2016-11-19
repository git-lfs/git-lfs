package lfs

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatFileBatchCheckScannerWithValidOutput(t *testing.T) {
	lines := []string{
		"short line",
		"0000000000000000000000000000000000000000 BLOB capitalized",
		"0000000000000000000000000000000000000001 blob not-a-size",
		"0000000000000000000000000000000000000002 blob 123",
		"0000000000000000000000000000000000000003 blob 1 0",
		"0000000000000000000000000000000000000004 blob 123456789",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))
	s := &catFileBatchCheckScanner{
		s:     bufio.NewScanner(r),
		limit: 1024,
	}

	assertNextOID(t, s, "")
	assertNextOID(t, s, "")
	assertNextOID(t, s, "")
	assertNextOID(t, s, "0000000000000000000000000000000000000002")
	assertNextOID(t, s, "")
	assertNextOID(t, s, "")
	assertScannerDone(t, s)
	assert.Equal(t, "", s.BlobOID())
}

type stringScanner interface {
	Next() (string, bool, error)
	Err() error
	Scan() bool
}

type genericScanner interface {
	Err() error
	Scan() bool
}

func assertNextScan(t *testing.T, scanner genericScanner) {
	assert.True(t, scanner.Scan())
	assert.Nil(t, scanner.Err())
}

func assertNextOID(t *testing.T, scanner *catFileBatchCheckScanner, oid string) {
	assertNextScan(t, scanner)
	assert.Equal(t, oid, scanner.BlobOID())
}

func assertScannerDone(t *testing.T, scanner genericScanner) {
	assert.False(t, scanner.Scan())
	assert.Nil(t, scanner.Err())
}
