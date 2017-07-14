package pack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexCount(t *testing.T) {
	fanout := make([]uint32, 256)
	for i := 0; i < len(fanout); i++ {
		fanout[i] = uint32(i)
	}

	idx := &Index{fanout: fanout}

	assert.EqualValues(t, 255, idx.Count())
}
