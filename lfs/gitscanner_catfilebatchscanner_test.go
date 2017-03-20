package lfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatFileBatchScannerWithValidOutput(t *testing.T) {
	blobs := []*Pointer{
		&Pointer{
			Version: "https://git-lfs.github.com/spec/v1",
			Oid:     "e71eefd918ea175b8f362611f981f648dbf9888ff74865077cb4c9077728f350",
			Size:    123,
			OidType: "sha256",
		},
		&Pointer{
			Version: "https://git-lfs.github.com/spec/v1",
			Oid:     "0eb69b651be65d5a61d6bebf2c53c811a5bf8031951111000e2077f4d7fe43b1",
			Size:    132,
			OidType: "sha256",
		},
	}

	reader := fakeReaderWithRandoData(t, blobs)
	if reader == nil {
		return
	}

	scanner := &catFileBatchScanner{r: bufio.NewReader(reader)}

	for i := 0; i < 5; i++ {
		assertNextEmptyPointer(t, scanner)
	}

	assertNextPointer(t, scanner, "e71eefd918ea175b8f362611f981f648dbf9888ff74865077cb4c9077728f350")

	for i := 0; i < 5; i++ {
		assertNextEmptyPointer(t, scanner)
	}

	assertNextPointer(t, scanner, "0eb69b651be65d5a61d6bebf2c53c811a5bf8031951111000e2077f4d7fe43b1")

	for i := 0; i < 5; i++ {
		assertNextEmptyPointer(t, scanner)
	}

	assertScannerDone(t, scanner)
	assert.Nil(t, scanner.Pointer())
}

func assertNextPointer(t *testing.T, scanner *catFileBatchScanner, oid string) {
	assertNextScan(t, scanner)
	p := scanner.Pointer()
	assert.NotNil(t, p)
	assert.Equal(t, oid, p.Oid)
}

func assertNextEmptyPointer(t *testing.T, scanner *catFileBatchScanner) {
	assertNextScan(t, scanner)
	assert.Nil(t, scanner.Pointer())
}

func fakeReaderWithRandoData(t *testing.T, blobs []*Pointer) io.Reader {
	buf := &bytes.Buffer{}
	rngbuf := make([]byte, 1000) // just under blob size cutoff
	rng := rand.New(rand.NewSource(0))

	for i := 0; i < 5; i++ {
		n, err := io.ReadFull(rng, rngbuf)
		if err != nil {
			t.Fatalf("error reading from rng: %+v", err)
		}
		writeFakeBuffer(t, buf, rngbuf, n)
	}

	for _, b := range blobs {
		ptrtext := b.Encoded()
		writeFakeBuffer(t, buf, []byte(ptrtext), len(ptrtext))
		for i := 0; i < 5; i++ {
			n, err := io.ReadFull(rng, rngbuf)
			if err != nil {
				t.Fatalf("error reading from rng: %+v", err)
			}
			writeFakeBuffer(t, buf, rngbuf, n)
		}
	}

	return bytes.NewBuffer(buf.Bytes())
}

func writeFakeBuffer(t *testing.T, buf *bytes.Buffer, by []byte, size int) {
	header := fmt.Sprintf("0000000000000000000000000000000000000000 blob %d", size)
	t.Log(header)
	buf.WriteString(header + "\n")
	buf.Write(by)
	buf.Write([]byte("\n"))
}
