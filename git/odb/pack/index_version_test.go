package pack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexVersionWidthV2(t *testing.T) {
	assert.EqualValues(t, 8, V2.Width())
}

func TestIndexVersionWidthPanicsOnUnknownVersion(t *testing.T) {
	v := IndexVersion(5)

	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("git/odb/pack: expected IndexVersion.Width() to panic()")
		}

		assert.Equal(t, "git/odb/pack: width unknown for pack version 5", err)
	}()

	v.Width()
}
