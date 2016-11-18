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

func TestCatFileBatchScanner(t *testing.T) {
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

	expected := map[string]bool{
		"e71eefd918ea175b8f362611f981f648dbf9888ff74865077cb4c9077728f350": true,
		"0eb69b651be65d5a61d6bebf2c53c811a5bf8031951111000e2077f4d7fe43b1": true,
	}
	for scanner.Scan() {
		p := scanner.Pointer()
		if !expected[p.Oid] {
			t.Errorf("Received unexpected OID: %q", p.Oid)
		} else {
			delete(expected, p.Oid)
		}
		assert.Equal(t, "0000000000000000000000000000000000000000", p.Sha1)
	}

	assert.Nil(t, scanner.Err())
	if len(expected) > 0 {
		t.Errorf("objects never received: %+v", expected)
	}
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
