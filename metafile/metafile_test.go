package metafile

import (
	"bytes"
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
