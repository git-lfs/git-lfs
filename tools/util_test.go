package tools

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/bmizerany/assert"
)

func TestCopyWithCallback(t *testing.T) {
	buf := bytes.NewBufferString("BOOYA")

	called := 0
	calledWritten := make([]int64, 0, 2)

	n, err := CopyWithCallback(ioutil.Discard, buf, 5, func(total int64, written int64, current int) error {
		called += 1
		calledWritten = append(calledWritten, written)
		assert.Equal(t, 5, int(total))
		return nil
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, 5, int(n))

	assert.Equal(t, 1, called)
	assert.Equal(t, 1, len(calledWritten))
	assert.Equal(t, 5, int(calledWritten[0]))
}
