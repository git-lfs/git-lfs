package pack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodePackfileDecodesIntegerVersion(t *testing.T) {
	p, err := DecodePackfile(bytes.NewReader([]byte{
		'P', 'A', 'C', 'K', // Pack header.
		0x0, 0x0, 0x0, 0x2, // Pack version.
		0x0, 0x0, 0x0, 0x0, // Number of packed objects.
	}))

	assert.NoError(t, err)
	assert.EqualValues(t, 2, p.Version)
}

func TestDecodePackfileDecodesIntegerCount(t *testing.T) {
	p, err := DecodePackfile(bytes.NewReader([]byte{
		'P', 'A', 'C', 'K', // Pack header.
		0x0, 0x0, 0x0, 0x2, // Pack version.
		0x0, 0x0, 0x1, 0x2, // Number of packed objects.
	}))

	assert.NoError(t, err)
	assert.EqualValues(t, 258, p.Objects)
}

func TestDecodePackfileReportsBadHeaders(t *testing.T) {
	p, err := DecodePackfile(bytes.NewReader([]byte{
		'W', 'R', 'O', 'N', 'G', // Malformed pack header.
		0x0, 0x0, 0x0, 0x0, // Pack version.
		0x0, 0x0, 0x0, 0x0, // Number of packed objects.
	}))

	assert.Equal(t, errBadPackHeader, err)
	assert.Nil(t, p)
}
