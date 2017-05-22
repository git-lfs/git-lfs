package odb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorerIncludesGivenEntries(t *testing.T) {
	sha := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	hex, err := hex.DecodeString(sha)

	assert.Nil(t, err)

	ms := newMemoryStorer(map[string]io.ReadWriter{
		sha: bytes.NewBuffer([]byte{0x1}),
	})

	buf, err := ms.Open(hex)
	assert.Nil(t, err)

	contents, err := ioutil.ReadAll(buf)
	assert.Nil(t, err)
	assert.Equal(t, []byte{0x1}, contents)
}

func TestMemoryStorerAcceptsNilEntries(t *testing.T) {
	ms := newMemoryStorer(nil)

	assert.NotNil(t, ms)
	assert.Equal(t, 0, len(ms.fs))
}

func TestMemoryStorerDoesntOpenMissingEntries(t *testing.T) {
	sha := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	hex, err := hex.DecodeString(sha)
	assert.Nil(t, err)

	defer func() {
		if err := recover(); err != nil {
			expeced := fmt.Sprintf("git/odb: memory storage cannot open %x, doesn't exist", hex)
			assert.Equal(t, expeced, err)
		} else {
			t.Fatal("expected panic()")
		}
	}()

	ms := newMemoryStorer(nil)

	ms.Open(hex)
}

func TestMemoryStorerStoresNewEntries(t *testing.T) {
	hex, err := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.Nil(t, err)

	ms := newMemoryStorer(nil)

	assert.Equal(t, 0, len(ms.fs))

	_, err = ms.Store(hex, strings.NewReader("hello"))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ms.fs))

	got, err := ms.Open(hex)
	assert.Nil(t, err)

	contents, err := ioutil.ReadAll(got)
	assert.Nil(t, err)
	assert.Equal(t, "hello", string(contents))
}

func TestMemoryStorerStoresNewEntriesExclusively(t *testing.T) {
	hex, err := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.Nil(t, err)

	ms := newMemoryStorer(nil)

	assert.Equal(t, 0, len(ms.fs))

	_, err = ms.Store(hex, new(bytes.Buffer))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ms.fs))

	defer func() {
		expected := fmt.Sprintf("git/odb: memory storage create %x, already exists", hex)
		if err := recover(); err != nil {
			assert.Equal(t, expected, err)
		} else {
			t.Fatal("expected panic()")
		}
	}()

	ms.Store(hex, new(bytes.Buffer))
}
