package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type genericScanner interface {
	Err() error
	Scan() bool
}

func assertNextScan(t *testing.T, scanner genericScanner) {
	assert.True(t, scanner.Scan())
	assert.Nil(t, scanner.Err())
}

func assertScannerDone(t *testing.T, scanner genericScanner) {
	assert.False(t, scanner.Scan())
	assert.Nil(t, scanner.Err())
}

func TestLsTreeParser(t *testing.T) {
	stdout := "100644 blob d899f6551a51cf19763c5955c7a06a2726f018e9      42	.gitattributes\000100644 blob 4d343e022e11a8618db494dc3c501e80c7e18197     126	PB SCN 16 Odhrán.wav"
	scanner := NewLsTreeScanner(strings.NewReader(stdout))

	assertNextTreeBlob(t, scanner, "d899f6551a51cf19763c5955c7a06a2726f018e9", ".gitattributes")
	assertNextTreeBlob(t, scanner, "4d343e022e11a8618db494dc3c501e80c7e18197", "PB SCN 16 Odhrán.wav")
	assertScannerDone(t, scanner)
}

func assertNextTreeBlob(t *testing.T, scanner *LsTreeScanner, oid, filename string) {
	assertNextScan(t, scanner)
	b := scanner.TreeBlob()
	assert.NotNil(t, b)
	assert.Equal(t, oid, b.Oid)
	assert.Equal(t, filename, b.Filename)
}

func BenchmarkLsTreeParser(b *testing.B) {
	stdout := "100644 blob d899f6551a51cf19763c5955c7a06a2726f018e9      42	.gitattributes\000100644 blob 4d343e022e11a8618db494dc3c501e80c7e18197     126	PB SCN 16 Odhrán.wav"

	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		scanner := NewLsTreeScanner(strings.NewReader(stdout))
		for scanner.Scan() {
		}
	}
}
