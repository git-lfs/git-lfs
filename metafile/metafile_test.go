package metafile

import (
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	n, err := Encode(&buf, "abc")
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

func TestDecode(t *testing.T) {
	buf := bytes.NewBufferString("# git-media\nabc\n")
	if sha, _ := Decode(buf); sha != "abc" {
		t.Errorf("Invalid SHA: %#v", sha)
	}
}

func TestDecodeExternal(t *testing.T) {
	buf := bytes.NewBufferString("# external\nabc\n")
	if sha, _ := Decode(buf); sha != "abc" {
		t.Errorf("Invalid SHA: %#v", sha)
	}
}
