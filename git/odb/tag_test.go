package odb

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagTypeReturnsCorrectObjectType(t *testing.T) {
	assert.Equal(t, TagObjectType, new(Tag).Type())
}

func TestTagEncode(t *testing.T) {
	tag := &Tag{
		Object:     []byte("aaaaaaaaaaaaaaaaaaaa"),
		ObjectType: CommitObjectType,
		Name:       "v2.4.0",
		Tagger:     "A U Thor <author@example.com>",

		Message: "The quick brown fox jumps over the lazy dog.",
	}

	buf := new(bytes.Buffer)

	n, err := tag.Encode(buf)

	assert.Nil(t, err)
	assert.EqualValues(t, buf.Len(), n)

	assertLine(t, buf, "object 6161616161616161616161616161616161616161")
	assertLine(t, buf, "type commit")
	assertLine(t, buf, "tag v2.4.0")
	assertLine(t, buf, "tagger A U Thor <author@example.com>")
	assertLine(t, buf, "")
	assertLine(t, buf, "The quick brown fox jumps over the lazy dog.")

	assert.Equal(t, 0, buf.Len())
}

func TestTagDecode(t *testing.T) {
	from := new(bytes.Buffer)

	fmt.Fprintf(from, "object 6161616161616161616161616161616161616161\n")
	fmt.Fprintf(from, "type commit\n")
	fmt.Fprintf(from, "tag v2.4.0\n")
	fmt.Fprintf(from, "tagger A U Thor <author@example.com>\n")
	fmt.Fprintf(from, "\n")
	fmt.Fprintf(from, "The quick brown fox jumps over the lazy dog.\n")

	flen := from.Len()

	tag := new(Tag)
	n, err := tag.Decode(from, int64(flen))

	assert.Nil(t, err)
	assert.Equal(t, n, flen)

	assert.Equal(t, []byte("aaaaaaaaaaaaaaaaaaaa"), tag.Object)
	assert.Equal(t, CommitObjectType, tag.ObjectType)
	assert.Equal(t, "v2.4.0", tag.Name)
	assert.Equal(t, "A U Thor <author@example.com>", tag.Tagger)
	assert.Equal(t, "The quick brown fox jumps over the lazy dog.", tag.Message)
}
