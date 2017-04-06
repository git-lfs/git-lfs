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
