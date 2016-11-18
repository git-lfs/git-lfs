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

	assertNextEmptyString(t, s)
	assertNextEmptyString(t, s)
	assertNextEmptyString(t, s)
	assertNextOID(t, s, "0000000000000000000000000000000000000002")
	assertNextEmptyString(t, s)
	assertNextEmptyString(t, s)
	assertStringScannerDone(t, s)
}

type stringScanner interface {
	Next() (string, bool, error)
}

func assertNextOID(t *testing.T, scanner stringScanner, oid string) {
	actual, hasNext, err := scanner.Next()
	assert.Equal(t, oid, actual)
	assert.Nil(t, err)
	assert.True(t, hasNext)
}

func assertNextEmptyString(t *testing.T, scanner stringScanner) {
	actual, hasNext, err := scanner.Next()
	assert.Equal(t, "", actual)
	assert.Nil(t, err)
	assert.True(t, hasNext)
}

func assertStringScannerDone(t *testing.T, scanner stringScanner) {
	actual, hasNext, err := scanner.Next()
	assert.Equal(t, "", actual)
	assert.Nil(t, err)
	assert.False(t, hasNext)
}
