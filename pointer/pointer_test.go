package pointer

import (
	"bufio"
	"bytes"
	"github.com/bmizerany/assert"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	pointer := NewPointer("booya", 12345)
	_, err := Encode(&buf, pointer)
	assert.Equal(t, nil, err)

	bufReader := bufio.NewReader(&buf)
	assertLine(t, bufReader, "[git-media]\n")
	assertLine(t, bufReader, "version=http://git-media.io/v/2\n")
	assertLine(t, bufReader, "oid=sha256:booya\n")
	assertLine(t, bufReader, "size=12345\n")

	line, err := bufReader.ReadString('\n')
	if err == nil {
		t.Fatalf("More to read: %s", line)
	}
	assert.Equal(t, "EOF", err.Error())
}

func assertLine(t *testing.T, r *bufio.Reader, expected string) {
	actual, err := r.ReadString('\n')
	assert.Equal(t, nil, err)
	assert.Equal(t, expected, actual)
}

func TestIniV2Decode(t *testing.T) {
	ex := `[git-media]
version=http://git-media.io/v/2
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

		// no sha
		"# git-media",

		// invalid section
		`[git-media2]
version=http://git-media.io/v/2
oid=sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size=12345`,

		// bad oid type
		`[git-media]
version=http://git-media.io/v/2
oid=shazam:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size=12345`,

		// no oid
		`[git-media]
version=http://git-media.io/v/2
size=12345`,

		// bad version
		`[git-media]
version=http://git-media.io/v/whatever
oid=sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size=12345`,

		// no version
		`[git-media]
oid=sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size=12345`,

		// bad size
		`[git-media]
version=http://git-media.io/v/2
oid=sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size=fif`,

		// no size
		`[git-media2]
version=http://git-media.io/v/2
oid=sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393`,
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
