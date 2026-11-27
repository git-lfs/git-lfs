package tq

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimitWriterCapsFileButReportsEveryByte(t *testing.T) {
	var buf strings.Builder
	w := newLimitWriter(&buf, 5)

	n, err := w.Write([]byte("abc"))
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	n, err = w.Write([]byte("defgh"))
	require.NoError(t, err)
	assert.Equal(t, 5, n, "the whole input is reported as written")

	n, err = w.Write([]byte("ij"))
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	assert.Equal(t, "abcde", buf.String())
}

func TestLimitWriterWithZeroLimitWritesNothing(t *testing.T) {
	var buf strings.Builder
	w := newLimitWriter(&buf, 0)

	n, err := w.Write([]byte("abc"))
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Empty(t, buf.String())
}

func TestLimitWriterNegativeLimitWritesNothing(t *testing.T) {
	var buf strings.Builder
	w := newLimitWriter(&buf, -1)

	n, err := w.Write([]byte("abc"))
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Empty(t, buf.String())
}

type failingWriter struct {
	remaining int
}

func (w *failingWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, errors.New("write failed")
	}
	if len(p) > w.remaining {
		n := w.remaining
		w.remaining = 0
		return n, errors.New("write failed")
	}
	w.remaining -= len(p)
	return len(p), nil
}

func TestLimitWriterReturnsUnderlyingError(t *testing.T) {
	w := newLimitWriter(&failingWriter{remaining: 2}, 10)

	n, err := w.Write([]byte("abc"))
	assert.Equal(t, 2, n)
	assert.EqualError(t, err, "write failed")
}
