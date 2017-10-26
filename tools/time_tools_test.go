package tools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeAtOrInNoDuration(t *testing.T) {
	now := time.Now()
	then := time.Now().Add(24 * time.Hour)

	got := TimeAtOrIn(now, then, time.Duration(0))

	assert.Equal(t, then, got)
}

func TestTimeAtOrInWithDuration(t *testing.T) {
	now := time.Now()
	duration := 5 * time.Minute
	expected := now.Add(duration)

	got := TimeAtOrIn(now, now, duration)

	assert.Equal(t, expected, got)
}

func TestTimeAtOrInZeroTime(t *testing.T) {
	now := time.Now()
	zero := time.Time{}

	got := TimeAtOrIn(now, zero, 0)

	assert.Equal(t, zero, got)
}

func TestIsExpiredAtOrInWithNonZeroTime(t *testing.T) {
	now := time.Now()
	within := 5 * time.Minute
	at := now.Add(10 * time.Minute)
	in := time.Duration(0)

	expired, ok := IsExpiredAtOrIn(now, within, at, in)

	assert.False(t, ok)
	assert.Equal(t, at, expired)
}

func TestIsExpiredAtOrInWithNonZeroDuration(t *testing.T) {
	now := time.Now()
	within := 5 * time.Minute
	at := time.Time{}
	in := 10 * time.Minute

	expired, ok := IsExpiredAtOrIn(now, within, at, in)

	assert.Equal(t, now.Add(in), expired)
	assert.False(t, ok)
}

func TestIsExpiredAtOrInWithNonZeroTimeExpired(t *testing.T) {
	now := time.Now()
	within := 5 * time.Minute
	at := now.Add(3 * time.Minute)
	in := time.Duration(0)

	expired, ok := IsExpiredAtOrIn(now, within, at, in)

	assert.True(t, ok)
	assert.Equal(t, at, expired)
}

func TestIsExpiredAtOrInWithNonZeroDurationExpired(t *testing.T) {
	now := time.Now()
	within := 5 * time.Minute
	at := time.Time{}
	in := -10 * time.Minute

	expired, ok := IsExpiredAtOrIn(now, within, at, in)

	assert.Equal(t, now.Add(in), expired)
	assert.True(t, ok)
}

func TestIsExpiredAtOrInWithAmbiguousTime(t *testing.T) {
	now := time.Now()
	within := 5 * time.Minute
	at := now.Add(-10 * time.Minute)
	in := 10 * time.Minute

	expired, ok := IsExpiredAtOrIn(now, within, at, in)

	assert.Equal(t, now.Add(in), expired)
	assert.False(t, ok)
}
