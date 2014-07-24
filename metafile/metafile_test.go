package metafile

import (
	"bytes"
	"github.com/bmizerany/assert"
	"strings"
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

func TestIniV2Decode(t *testing.T) {
	ex := `[git-media]
version="http://git-media.io/v/2"
oid=sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size=12345`

	p, err := Decode(bytes.NewBufferString(ex))
	assertEqualWithExample(t, ex, nil, err)
	assertEqualWithExample(t, ex, latest, p.Version)
	assertEqualWithExample(t, ex, "4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393", p.Oid)
	assertEqualWithExample(t, ex, int64(12345), p.Size)
}

func TestAlphaDecode(t *testing.T) {
	examples := []string{
		"# git-media\nabc\n",
		"# external\nabc\n",
	}

	for _, ex := range examples {
		p, err := Decode(bytes.NewBufferString(ex))
		assertEqualWithExample(t, ex, nil, err)
		assertEqualWithExample(t, ex, "abc", p.Oid)
		assertEqualWithExample(t, ex, int64(0), p.Size)
		assertEqualWithExample(t, ex, "sha256", p.OidType)
		assertEqualWithExample(t, ex, alpha, p.Version)
	}
}

func TestDecodeInvalid(t *testing.T) {
	examples := []string{
		"invalid stuff",
		"# git-media",
	}

	for _, ex := range examples {
		p, err := Decode(bytes.NewBufferString(ex))
		if err == nil {
			t.Errorf("No error decoding: %v\nFrom:\n%s", p, strings.TrimSpace(ex))
		}
	}
}

func assertEqualWithExample(t *testing.T, example string, expected, actual interface{}) {
	assert.Equalf(t, expected, actual, "Example:\n%s", strings.TrimSpace(example))
}
