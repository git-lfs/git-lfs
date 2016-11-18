package lfs

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatFileBatchCheckScanner(t *testing.T) {
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

	var found []string
	for s.Scan() {
		found = append(found, s.BlobOID())
	}

	assert.Nil(t, s.Err())
	assert.Equal(t, 1, len(found))
	assert.Equal(t, "0000000000000000000000000000000000000002", found[0])
}
