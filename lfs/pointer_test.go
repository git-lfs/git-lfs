package lfs

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	pointer := NewPointer("booya", 12345, nil)
	_, err := EncodePointer(&buf, pointer)
	assert.Nil(t, err)

	bufReader := bufio.NewReader(&buf)
	assertLine(t, bufReader, "version https://git-lfs.github.com/spec/v1\n")
	assertLine(t, bufReader, "oid sha256:booya\n")
	assertLine(t, bufReader, "size 12345\n")

	line, err := bufReader.ReadString('\n')
	if err == nil {
		t.Fatalf("More to read: %s", line)
	}
	assert.Equal(t, "EOF", err.Error())
}

func TestEncodeEmpty(t *testing.T) {
	var buf bytes.Buffer
	pointer := NewPointer("", 0, nil)
	_, err := EncodePointer(&buf, pointer)
	assert.Equal(t, nil, err)

	bufReader := bufio.NewReader(&buf)
	val, err := bufReader.ReadString('\n')
	assert.Equal(t, "", val)
	assert.Equal(t, "EOF", err.Error())
}

func TestEncodeExtensions(t *testing.T) {
	var buf bytes.Buffer
	exts := []*PointerExtension{
		NewPointerExtension("foo", 0, "foo_oid"),
		NewPointerExtension("bar", 1, "bar_oid"),
		NewPointerExtension("baz", 2, "baz_oid"),
	}
	pointer := NewPointer("main_oid", 12345, exts)
	_, err := EncodePointer(&buf, pointer)
	assert.Nil(t, err)

	bufReader := bufio.NewReader(&buf)
	assertLine(t, bufReader, "version https://git-lfs.github.com/spec/v1\n")
	assertLine(t, bufReader, "ext-0-foo sha256:foo_oid\n")
	assertLine(t, bufReader, "ext-1-bar sha256:bar_oid\n")
	assertLine(t, bufReader, "ext-2-baz sha256:baz_oid\n")
	assertLine(t, bufReader, "oid sha256:main_oid\n")
	assertLine(t, bufReader, "size 12345\n")

	line, err := bufReader.ReadString('\n')
	if err == nil {
		t.Fatalf("More to read: %s", line)
	}
	assert.Equal(t, "EOF", err.Error())
}

func assertLine(t *testing.T, r *bufio.Reader, expected string) {
	actual, err := r.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestDecodeTinyFile(t *testing.T) {
	ex := "this is not a git-lfs file!"
	p, err := DecodePointer(bytes.NewBufferString(ex))
	if p != nil {
		t.Errorf("pointer was decoded: %v", p)
	}

	if !errors.IsNotAPointerError(err) {
		t.Errorf("error is not a NotAPointerError: %s: '%v'", reflect.TypeOf(err), err)
	}
}

func TestDecode(t *testing.T) {
	ex := `version https://git-lfs.github.com/spec/v1
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`

	p, err := DecodePointer(bytes.NewBufferString(ex))
	assertEqualWithExample(t, ex, nil, err)
	assertEqualWithExample(t, ex, latest, p.Version)
	assertEqualWithExample(t, ex, "4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393", p.Oid)
	assertEqualWithExample(t, ex, "sha256", p.OidType)
	assertEqualWithExample(t, ex, int64(12345), p.Size)
}

func TestDecodeExtensions(t *testing.T) {
	ex := `version https://git-lfs.github.com/spec/v1
ext-0-foo sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
ext-1-bar sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
ext-2-baz sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`

	p, err := DecodePointer(bytes.NewBufferString(ex))
	assertEqualWithExample(t, ex, nil, err)
	assertEqualWithExample(t, ex, latest, p.Version)
	assertEqualWithExample(t, ex, "4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393", p.Oid)
	assertEqualWithExample(t, ex, int64(12345), p.Size)
	assertEqualWithExample(t, ex, "sha256", p.OidType)
	assertEqualWithExample(t, ex, "foo", p.Extensions[0].Name)
	assertEqualWithExample(t, ex, 0, p.Extensions[0].Priority)
	assertEqualWithExample(t, ex, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", p.Extensions[0].Oid)
	assertEqualWithExample(t, ex, "sha256", p.Extensions[0].OidType)
	assertEqualWithExample(t, ex, "bar", p.Extensions[1].Name)
	assertEqualWithExample(t, ex, 1, p.Extensions[1].Priority)
	assertEqualWithExample(t, ex, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", p.Extensions[1].Oid)
	assertEqualWithExample(t, ex, "sha256", p.Extensions[1].OidType)
	assertEqualWithExample(t, ex, "baz", p.Extensions[2].Name)
	assertEqualWithExample(t, ex, 2, p.Extensions[2].Priority)
	assertEqualWithExample(t, ex, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", p.Extensions[2].Oid)
	assertEqualWithExample(t, ex, "sha256", p.Extensions[2].OidType)
}

func TestDecodeExtensionsSort(t *testing.T) {
	ex := `version https://git-lfs.github.com/spec/v1
ext-2-baz sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
ext-0-foo sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
ext-1-bar sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`

	p, err := DecodePointer(bytes.NewBufferString(ex))
	assertEqualWithExample(t, ex, nil, err)
	assertEqualWithExample(t, ex, latest, p.Version)
	assertEqualWithExample(t, ex, "4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393", p.Oid)
	assertEqualWithExample(t, ex, int64(12345), p.Size)
	assertEqualWithExample(t, ex, "sha256", p.OidType)
	assertEqualWithExample(t, ex, "foo", p.Extensions[0].Name)
	assertEqualWithExample(t, ex, 0, p.Extensions[0].Priority)
	assertEqualWithExample(t, ex, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", p.Extensions[0].Oid)
	assertEqualWithExample(t, ex, "sha256", p.Extensions[0].OidType)
	assertEqualWithExample(t, ex, "bar", p.Extensions[1].Name)
	assertEqualWithExample(t, ex, 1, p.Extensions[1].Priority)
	assertEqualWithExample(t, ex, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", p.Extensions[1].Oid)
	assertEqualWithExample(t, ex, "sha256", p.Extensions[1].OidType)
	assertEqualWithExample(t, ex, "baz", p.Extensions[2].Name)
	assertEqualWithExample(t, ex, 2, p.Extensions[2].Priority)
	assertEqualWithExample(t, ex, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", p.Extensions[2].Oid)
	assertEqualWithExample(t, ex, "sha256", p.Extensions[2].OidType)
}

func TestDecodePreRelease(t *testing.T) {
	ex := `version https://hawser.github.com/spec/v1
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`

	p, err := DecodePointer(bytes.NewBufferString(ex))
	assertEqualWithExample(t, ex, nil, err)
	assertEqualWithExample(t, ex, latest, p.Version)
	assertEqualWithExample(t, ex, "4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393", p.Oid)
	assertEqualWithExample(t, ex, "sha256", p.OidType)
	assertEqualWithExample(t, ex, int64(12345), p.Size)
}

func TestDecodeFromEmptyReader(t *testing.T) {
	p, buf, err := DecodeFrom(strings.NewReader(""))
	by, rerr := ioutil.ReadAll(buf)

	assert.Nil(t, rerr)
	assert.EqualError(t, err, "Pointer file error: invalid header")
	assert.Nil(t, p)
	assert.Empty(t, by)
}

func TestDecodeInvalid(t *testing.T) {
	examples := []string{
		"invalid stuff",

		// no sha
		"# git-media",

		// bad oid
		`version https://git-lfs.github.com/spec/v1
oid sha256:boom
size 12345`,

		// bad oid type
		`version https://git-lfs.github.com/spec/v1
oid shazam:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`,

		// no oid
		`version https://git-lfs.github.com/spec/v1
size 12345`,

		// bad version
		`version http://git-media.io/v/whatever
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`,

		// no version
		`oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`,

		// bad size
		`version https://git-lfs.github.com/spec/v1
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size fif`,

		// no size
		`version https://git-lfs.github.com/spec/v1
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393`,

		// bad `key value` format
		`version=https://git-lfs.github.com/spec/v1
oid=sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size=fif`,

		// no git-media
		`version=http://wat.io/v/2
oid=sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size=fif`,

		// extra key
		`version https://git-lfs.github.com/spec/v1
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345
wat wat`,

		// keys out of order
		`version https://git-lfs.github.com/spec/v1
size 12345
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393`,

		// bad ext name
		`version https://git-lfs.github.com/spec/v1
ext-0-$$$$ sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`,

		// bad ext priority
		`version https://git-lfs.github.com/spec/v1
ext-#-foo sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`,

		// duplicate ext priority
		`version https://git-lfs.github.com/spec/v1
ext-0-foo sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
ext-0-bar sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`,

		// ext priority over 9
		`version https://git-lfs.github.com/spec/v1
ext-10-foo sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`,

		// bad ext oid
		`version https://git-lfs.github.com/spec/v1
ext-0-foo sha256:boom
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`,

		// bad ext oid type
		`version https://git-lfs.github.com/spec/v1
ext-0-foo boom:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345`,
	}

	for _, ex := range examples {
		p, err := DecodePointer(bytes.NewBufferString(ex))
		if err == nil {
			t.Errorf("No error decoding: %v\nFrom:\n%s", p, strings.TrimSpace(ex))
		}
	}
}

func assertEqualWithExample(t *testing.T, example string, expected, actual interface{}) {
	assert.Equal(t, expected, actual, "Example:\n%s", strings.TrimSpace(example))
}
