package lfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/gitobj/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPointerScannerWithValidOutput(t *testing.T) {
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

	be, _ := gitobj.NewMemoryBackend(nil)
	db, _ := gitobj.FromBackend(be)
	shas := fakeObjectsWithRandoData(t, db, blobs)

	scanner := &PointerScanner{
		scanner: git.NewObjectScannerFrom(db),
	}
	iter := 0

	for i := 0; i < 5; i++ {
		assertNextEmptyPointer(t, scanner, shas[iter])
		iter++
	}

	assertNextPointer(t, scanner, shas[iter], "e71eefd918ea175b8f362611f981f648dbf9888ff74865077cb4c9077728f350")
	iter++

	for i := 0; i < 5; i++ {
		assertNextEmptyPointer(t, scanner, shas[iter])
		iter++
	}

	assertNextPointer(t, scanner, shas[iter], "0eb69b651be65d5a61d6bebf2c53c811a5bf8031951111000e2077f4d7fe43b1")
	iter++

	for i := 0; i < 5; i++ {
		assertNextEmptyPointer(t, scanner, shas[iter])
		iter++
	}
}

func TestPointerScannerWithLargeBlobs(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 1025))
	sha := sha256.New()
	rng := rand.New(rand.NewSource(0))

	_, err := io.CopyN(io.MultiWriter(sha, buf), rng, 1025)
	require.Nil(t, err)

	be, _ := gitobj.NewMemoryBackend(nil)
	db, _ := gitobj.FromBackend(be)

	fake := bytes.NewBuffer(nil)
	oid := writeFakeBuffer(t, db, fake, buf.Bytes(), buf.Len())

	scanner := &PointerScanner{
		scanner: git.NewObjectScannerFrom(db),
	}

	require.True(t, scanner.Scan(oid))
	assert.Nil(t, scanner.Pointer())
	assert.Equal(t, fmt.Sprintf("%x", sha.Sum(nil)), scanner.ContentsSha())
}

func assertNextPointer(t *testing.T, scanner *PointerScanner, sha string, oid string) {
	assert.True(t, scanner.Scan(sha))
	assert.Nil(t, scanner.Err())

	p := scanner.Pointer()

	assert.NotNil(t, p)
	assert.Equal(t, oid, p.Oid)
}

func assertNextEmptyPointer(t *testing.T, scanner *PointerScanner, sha string) {
	assert.True(t, scanner.Scan(sha))
	assert.Nil(t, scanner.Err())

	assert.Nil(t, scanner.Pointer())
}

func fakeObjectsWithRandoData(t *testing.T, db *gitobj.ObjectDatabase, blobs []*Pointer) []string {
	buf := &bytes.Buffer{}
	rngbuf := make([]byte, 1000) // just under blob size cutoff
	rng := rand.New(rand.NewSource(0))
	oids := make([]string, 0)

	for i := 0; i < 5; i++ {
		n, err := io.ReadFull(rng, rngbuf)
		if err != nil {
			t.Fatalf("error reading from rng: %+v", err)
		}
		oids = append(oids, writeFakeBuffer(t, db, buf, rngbuf, n))
	}

	for _, b := range blobs {
		ptrtext := b.Encoded()
		oids = append(oids, writeFakeBuffer(t, db, buf, []byte(ptrtext), len(ptrtext)))
		for i := 0; i < 5; i++ {
			n, err := io.ReadFull(rng, rngbuf)
			if err != nil {
				t.Fatalf("error reading from rng: %+v", err)
			}
			oids = append(oids, writeFakeBuffer(t, db, buf, rngbuf, n))
		}
	}

	return oids
}

func writeFakeBuffer(t *testing.T, db *gitobj.ObjectDatabase, buf *bytes.Buffer, by []byte, size int) string {
	oid, _ := db.WriteBlob(gitobj.NewBlobFromBytes(by))
	return hex.EncodeToString(oid)
}
