package metafile

import (
	"bytes"
	"github.com/bmizerany/assert"
	"testing"
)

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	pointer := NewPointer("abc", 0)
	n, err := Encode(&buf, pointer)
	if err != nil {
		t.Errorf("Error encoding: %s", err)
	}

	if n != len(MediaWarning)+4 {
		t.Errorf("wrong number of written bytes")
	}

	header := make([]byte, len(MediaWarning))
	buf.Read(header)

	if head := string(header); head != string(MediaWarning) {
		t.Errorf("Media warning not read: %s\n", head)
	}

	shabytes := make([]byte, 3)
	buf.Read(shabytes)

	if sha := string(shabytes); sha != "abc" {
		t.Errorf("Invalid sha: %#v", sha)
	}
}

func TestIniDecode(t *testing.T) {
	buf := bytes.NewBufferString(`[git-media]
version="http://git-media.io/v/2"
oid=sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size=12345
`)

	p, err := Decode(buf)
	assert.Equal(t, nil, err)
	assert.Equal(t, latest, p.Version)
	assert.Equal(t, "4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393", p.Oid)
	assert.Equal(t, int64(12345), p.Size)
}

func TestAlphaDecode(t *testing.T) {
	buf := bytes.NewBufferString("# git-media\nabc\n")
	if pointer, _ := Decode(buf); pointer.Oid != "abc" {
		t.Errorf("Invalid SHA: %#v", pointer.Oid)
	}
}

func TestAlphaDecodeExternal(t *testing.T) {
	buf := bytes.NewBufferString("# external\nabc\n")
	if pointer, _ := Decode(buf); pointer.Oid != "abc" {
		t.Errorf("Invalid SHA: %#v", pointer.Oid)
	}
}

func TestDecodeInvalid(t *testing.T) {
	buf := bytes.NewBufferString("invalid stuff")
	if _, err := Decode(buf); err == nil {
		t.Errorf("Decoded invalid sha")
	}
}

func TestAlphaDecodeWithValidHeaderNoSha(t *testing.T) {
	buf := bytes.NewBufferString("# git-media")
	if _, err := Decode(buf); err == nil {
		t.Errorf("Decoded with header but no sha")
	}
}
